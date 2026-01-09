package cassandracluster

import (
	"context"
	"fmt"
	"testing"

	api "github.com/cscetbon/casskop/api/v2"
	"github.com/cscetbon/casskop/pkg/k8s"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestStorageUpsize(t *testing.T) {
	overrideDelayWaitWithNoDelay()
	defer restoreDefaultDelayWait()

	const InitialCapacity = "3Gi"
	const NewCapacity = "10Gi"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	assert := assert.New(t)

	// setup cluster
	rcc, req := createCassandraClusterWithNoDisruption(t, "cassandracluster-2racks-with-storage.yaml")
	assert.Equal(int32(3), rcc.cc.Spec.NodesPerRacks)

	cassandraCluster := rcc.cc.DeepCopy()
	datacenters := cassandraCluster.Spec.Topology.DC
	assert.Equal(1, len(datacenters))
	assert.Equal(2, len(datacenters[0].Rack))
	assertClusterInitialized(assert, rcc)

	// check initial sts capacity
	for _, dc := range datacenters {
		for _, rack := range dc.Rack {
			assertStsCapacity(assert, rcc, dc, rack.Name, InitialCapacity)
		}
	}

	// simulate all PVCs ready
	simulatePVCsReadyForWholeCluster(assert, rcc, datacenters, InitialCapacity)

	// mock no joining nodes
	dc := datacenters[0]
	sts1Name := cassandraCluster.Name + fmt.Sprintf("-%s-%s", dc.Name, dc.Rack[0].Name)
	firstPod := podHost(sts1Name, 0, rcc)
	registerJolokiaOperationJoiningNodes(firstPod, 0)

	// request storage upsize
	cassandraCluster.Spec.DataCapacity = NewCapacity
	assert.NoError(rcc.Client.Update(context.TODO(), cassandraCluster))

	// reconcile storage upsize rack by rack
	for _, currentRack := range cassandraCluster.Spec.Topology.DC[0].Rack {
		currentDcRackName := cassandraCluster.GetDCRackName(dc.Name, currentRack.Name)

		// should start on current rack
		reconcileValidation(t, rcc, *req)
		assert.Equal(NewCapacity, cassandraCluster.Spec.DataCapacity)
		assertUpsizeInProgress(assert, rcc, currentDcRackName)

		// should remove current rack orphan option
		reconcileValidation(t, rcc, *req)
		assertStsNotFound(assert, rcc, dc, currentRack.Name)

		// should recreate current rack with new capacity
		reconcileValidation(t, rcc, *req)
		assertStsCapacity(assert, rcc, dc, currentRack.Name, NewCapacity)

		// simulate current rack sts is ready
		simulateStsIsReady(assert, rcc, dc, currentRack.Name)

		// should update current rack pvcs to new capacity
		reconcileValidation(t, rcc, *req)
		assertRackPVCs(assert, rcc, dc, currentRack.Name, NewCapacity)

		// should do nothing till current rack pvcs are not resized
		reconcileValidation(t, rcc, *req)
		assertUpsizeInProgress(assert, rcc, currentDcRackName)

		simulateCurrentRackPvcsAreUpsizedByProvisioner(assert, rcc, dc, currentRack.Name)

		// should finalize storage upsize on current rack
		reconcileValidation(t, rcc, *req)
		assertUpsizeDoneOnRack(assert, rcc, currentDcRackName)
	}

	// should finalize storage upsize globally
	reconcileValidation(t, rcc, *req)
	assertUpsizeDoneGlobally(assert, rcc)
}

func simulatePVCsReadyForWholeCluster(assert *assert.Assertions, rcc *CassandraClusterReconciler, datacenters api.DCSlice, expectedCapacity string) {
	for _, dc := range datacenters {
		for _, rack := range dc.Rack {
			stfsName := rcc.cc.Name + fmt.Sprintf("-%s-%s", dc.Name, rack.Name)
			for i := int32(0); i < rcc.cc.Spec.NodesPerRacks; i++ {
				pvcName := fmt.Sprintf("%s-%s-%d", "data", stfsName, i)
				pvc := corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pvcName,
						Namespace: rcc.cc.Namespace,
						Labels:    k8s.LabelsForCassandraDCRack(rcc.cc, dc.Name, rack.Name),
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								"storage": resource.MustParse(expectedCapacity),
							},
						},
					},
				}
				assert.NoError(rcc.Client.Create(context.TODO(), &pvc))
			}
		}
	}
}

func assertStsNotFound(assert *assert.Assertions, rcc *CassandraClusterReconciler, dc api.DC, rackName string) {
	_, err := getSts(rcc, dc, rackName)
	assert.True(apierrors.IsNotFound(err))
}

func assertStsCapacity(assert *assert.Assertions, rcc *CassandraClusterReconciler, dc api.DC, rackName string, expectedCapacity string) {
	currentRackSts, err := getSts(rcc, dc, rackName)
	assert.NoError(err)
	assert.Equal(resource.MustParse(expectedCapacity), currentRackSts.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests["storage"])
}

func simulateStsIsReady(assert *assert.Assertions, rcc *CassandraClusterReconciler, dc api.DC, rackName string) {
	currentRackSts, err := getSts(rcc, dc, rackName)
	assert.NoError(err)
	currentRackSts.Status.Replicas = *currentRackSts.Spec.Replicas
	currentRackSts.Status.ReadyReplicas = *currentRackSts.Spec.Replicas
	assert.NoError(rcc.Client.Status().Update(ctx, currentRackSts))
}

func getSts(rcc *CassandraClusterReconciler, dc api.DC, rackName string) (*appsv1.StatefulSet, error) {
	var currentRackSts = &appsv1.StatefulSet{}
	currentRackStsName := rcc.cc.Name + fmt.Sprintf("-%s-%s", dc.Name, rackName)
	currentRackStsNamespacedName := types.NamespacedName{Namespace: rcc.cc.Namespace, Name: currentRackStsName}
	err := rcc.Client.Get(context.TODO(), currentRackStsNamespacedName, currentRackSts)
	return currentRackSts, err
}

func assertRackPVCs(assert *assert.Assertions, rcc *CassandraClusterReconciler, dc api.DC, rackName string, expectedCapacity string) {
	currentRackPvcs, err := rcc.ListPVC(ctx, rcc.cc.Namespace, k8s.LabelsForCassandraDCRack(rcc.cc, dc.Name, rackName))
	assert.NoError(err)
	assert.Equal(3, len(currentRackPvcs.Items))
	for _, pvc := range currentRackPvcs.Items {
		assert.Equal(resource.MustParse(expectedCapacity), pvc.Spec.Resources.Requests["storage"])
	}
}

func simulateCurrentRackPvcsAreUpsizedByProvisioner(assert *assert.Assertions, rcc *CassandraClusterReconciler, dc api.DC, rackName string) {
	currentRackPvcs, err := rcc.ListPVC(ctx, rcc.cc.Namespace, k8s.LabelsForCassandraDCRack(rcc.cc, dc.Name, rackName))
	assert.NoError(err)
	for _, pvc := range currentRackPvcs.Items {
		if pvc.Status.Capacity == nil {
			pvc.Status.Capacity = corev1.ResourceList{}
		}
		pvc.Status.Capacity["storage"] = pvc.Spec.Resources.Requests["storage"]
		assert.NoError(rcc.Client.Status().Update(context.TODO(), &pvc))
	}
}

func assertClusterInitialized(assert *assert.Assertions, rcc *CassandraClusterReconciler) {
	assertClusterStatusPhase(assert, rcc, api.ClusterPhaseRunning)
	assertClusterStatusLastAction(assert, rcc, api.ClusterPhaseInitial, api.StatusDone)
	for dcRackName := range rcc.cc.Status.CassandraRackStatus {
		assertRackStatusPhase(assert, rcc, dcRackName, api.ClusterPhaseRunning)
		assertRackStatusLastAction(assert, rcc, dcRackName, api.ClusterPhaseInitial, api.StatusDone)
	}
}

func assertUpsizeInProgress(assert *assert.Assertions, rcc *CassandraClusterReconciler, dcRackName string) {
	assert.NotEmpty(rcc.cc.Status.CassandraRackStatus[dcRackName].StatefulSetSnapshotBeforeStorageResize)

	assertClusterStatusPhase(assert, rcc, api.ClusterPhasePending)
	assertClusterStatusLastAction(assert, rcc, api.ActionStorageUpsize, api.StatusOngoing)
	assertRackStatusPhase(assert, rcc, dcRackName, api.ClusterPhasePending)
	assertRackStatusLastAction(assert, rcc, dcRackName, api.ActionStorageUpsize, api.StatusOngoing)
}

func assertUpsizeDoneOnRack(assert *assert.Assertions, rcc *CassandraClusterReconciler, dcRackName string) {
	assertClusterStatusPhase(assert, rcc, api.ClusterPhasePending)
	assertClusterStatusLastAction(assert, rcc, api.ActionStorageUpsize, api.StatusOngoing)
	assertRackStatusPhase(assert, rcc, dcRackName, api.ClusterPhasePending)
	assertRackStatusLastAction(assert, rcc, dcRackName, api.ActionStorageUpsize, api.StatusDone)
}

func assertUpsizeDoneGlobally(assert *assert.Assertions, rcc *CassandraClusterReconciler) {
	assertClusterStatusPhase(assert, rcc, api.ClusterPhaseRunning)
	assertClusterStatusLastAction(assert, rcc, api.ActionStorageUpsize, api.StatusDone)
	for dcRackName := range rcc.cc.Status.CassandraRackStatus {
		assertRackStatusPhase(assert, rcc, dcRackName, api.ClusterPhaseRunning)
		assertRackStatusLastAction(assert, rcc, dcRackName, api.ActionStorageUpsize, api.StatusDone)
	}
}

func TestStorageUpsizeDoesNotStartWhenOtherOperationInProgress(t *testing.T) {

	overrideDelayWaitWithNoDelay()
	defer restoreDefaultDelayWait()

	const InitialCapacity = "3Gi"
	const NewCapacity = "10Gi"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	assert := assert.New(t)

	// setup cluster
	rcc, req := createCassandraClusterWithNoDisruption(t, "cassandracluster-1DC.yaml")
	assert.Equal(int32(3), rcc.cc.Spec.NodesPerRacks)

	cassandraCluster := rcc.cc.DeepCopy()
	datacenters := cassandraCluster.Spec.Topology.DC
	assert.Equal(1, len(datacenters))
	assert.Equal(1, len(datacenters[0].Rack))
	assertClusterInitialized(assert, rcc)

	// check initial sts capacity
	dc := datacenters[0]
	rack := dc.Rack[0]
	assertStsCapacity(assert, rcc, dc, rack.Name, InitialCapacity)

	// mock no joining nodes
	sts1Name := cassandraCluster.Name + fmt.Sprintf("-%s-%s", dc.Name, rack.Name)
	firstPod := podHost(sts1Name, 0, rcc)
	registerJolokiaOperationJoiningNodes(firstPod, 0)

	// request scale out - assert operation started
	cassandraCluster.Spec.NodesPerRacks = 4
	rcc.Client.Update(context.TODO(), cassandraCluster)

	reconcileValidation(t, rcc, *req)
	assert.GreaterOrEqual(jolokiaCallsCount(firstPod), 0)
	assertStatefulsetReplicas(ctx, t, rcc, 4, cassandraCluster.Namespace, sts1Name)
	assertClusterStatusLastAction(assert, rcc, api.ActionScaleUp, api.StatusOngoing)
	assertRackStatusLastAction(assert, rcc, "dc1-rack1", api.ActionScaleUp, api.StatusOngoing)

	// request storage upsize while scale-out ongoing - assert new capacity accepted but upsize is not started and sts capacity untouched
	cassandraCluster = rcc.cc.DeepCopy()
	cassandraCluster.Spec.DataCapacity = NewCapacity
	assert.NoError(rcc.Client.Update(context.TODO(), cassandraCluster))

	reconcileValidation(t, rcc, *req)

	cassandraCluster = rcc.cc.DeepCopy()
	assert.Equal(NewCapacity, cassandraCluster.Spec.DataCapacity)
	assertStsCapacity(assert, rcc, dc, rack.Name, InitialCapacity)

	assertStatefulsetReplicas(ctx, t, rcc, 4, cassandraCluster.Namespace, sts1Name)
	assertClusterStatusLastAction(assert, rcc, api.ActionScaleUp, api.StatusOngoing)
	assertRackStatusLastAction(assert, rcc, "dc1-rack1", api.ActionScaleUp, api.StatusOngoing)

	// scale-out finishes - capacity still unchanged
	simulateNewPodsReady(t, rcc, sts1Name, dc, 3, 4)
	registerJolokiaOperationJoiningNodes(firstPod, 0)

	reconcileValidation(t, rcc, *req)

	assert.GreaterOrEqual(jolokiaCallsCount(firstPod), 1)
	assertClusterStatusPhase(assert, rcc, api.ClusterPhaseRunning)
	assertRackStatusPhase(assert, rcc, "dc1-rack1", api.ClusterPhaseRunning)
	assertClusterStatusLastAction(assert, rcc, api.ActionScaleUp, api.StatusDone)
	assertRackStatusLastAction(assert, rcc, "dc1-rack1", api.ActionScaleUp, api.StatusDone)
	assertStsCapacity(assert, rcc, dc, rack.Name, InitialCapacity)

	// next reconcile should initialize storage upsize action
	simulatePVCsReadyForWholeCluster(assert, rcc, datacenters, InitialCapacity)

	reconcileValidation(t, rcc, *req)

	assertUpsizeInProgress(assert, rcc, cassandraCluster.GetDCRackName(dc.Name, rack.Name))
	assertStsCapacity(assert, rcc, dc, rack.Name, InitialCapacity)

	// next two reconciles do: 1. sts removal,2.  sts re-creation with new capacity
	reconcileValidation(t, rcc, *req)
	reconcileValidation(t, rcc, *req)
	assertStsCapacity(assert, rcc, dc, rack.Name, NewCapacity)
}
