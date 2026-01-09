package storageupsize

import (
	"context"

	api "github.com/cscetbon/casskop/api/v2"
	"github.com/cscetbon/casskop/controllers/cassandracluster/pods"
	"github.com/cscetbon/casskop/controllers/cassandracluster/storagestateclient"
	as "github.com/cscetbon/casskop/controllers/cassandracluster/storageupsize/actionstep"
	"github.com/cscetbon/casskop/controllers/cassandracluster/sts"
	"github.com/cscetbon/casskop/controllers/cassandracluster/view"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func ShouldBeStarted(rack view.RackView, requestedCapacity string) bool {
	requested := silentParseResourceQuantity(requestedCapacity)
	_, current := findDataCapacity(rack.LivingStatefulSet().Spec.VolumeClaimTemplates)
	if !requested.Equal(current) {
		rack.Log().Infof("Storage upsize should be started: ask %v and have %v", requested, current)
		return true
	}
	return false
}

func Start(rack view.RackView) {
	startUpsizeAction(rack)
}

func IsStarted(dcRackStatus *api.CassandraRackStatus) bool {
	return dcRackStatus.CassandraLastAction.Name == api.ActionStorageUpsize.Name &&
		dcRackStatus.CassandraLastAction.Status != api.StatusDone
}

// Reconcile performs the storage upsize action steps
// Each step may
// - return an error
// - execute an action and break the loop (if action was not finished before or even not started yet)
// - do nothing and continue to the next step pass (if action was finished before)
// Usually step do its job once and break the loop, then in the next reconcile loop this step "pass" and the next step is executed
func Reconcile(ctx context.Context, cc *api.CassandraCluster, rack view.RackView, newDataCapacity resource.Quantity,
	storageStateClient storagestateclient.StorageStateClient, stsClient sts.StsClient, podsClient pods.PodsClient) error {

	dataPVCs := make([]corev1.PersistentVolumeClaim, 0)

	steps := []func() as.StepResult{
		func() as.StepResult { return makeOldStatefulSetSnapshot(rack) },
		func() as.StepResult { return removeStatefulSetOrphan(ctx, cc, rack, stsClient) },
		func() as.StepResult { return recreateStatefulSetWithNewCapacity(ctx, rack, newDataCapacity, stsClient) },
		func() as.StepResult { return fetchDataPvcs(ctx, cc, rack, storageStateClient, &dataPVCs) },
		func() as.StepResult { return ensureAllPVCsHaveNewCapacity(ctx, cc, dataPVCs, rack, storageStateClient) },
		func() as.StepResult { return waitTillAllFilesystemsHaveNewCapacity(cc, dataPVCs, rack) },
		func() as.StepResult { return waitTillStatefulSetAndAllPodsAreReady(ctx, cc, rack, podsClient) },
	}
	for _, executeStep := range steps {
		if stepResult := executeStep(); stepResult.HasError() {
			return stepResult.Error()
		} else if stepResult.ShouldBreakReconcileLoop() {
			return nil
		}
	}

	return nil
}

// RevertAnyStorageUpsizeBeyondUpsizeAction reverts any storage capacity changes if upsize action IS NOT started
// current action should finish, then upsize action should be started and then these changes should be applied
func RevertAnyStorageUpsizeBeyondUpsizeAction(rack view.RackView, newStatefulSet *appsv1.StatefulSet) {
	if !IsStarted(rack.RackStatus()) {
		_, current := findDataCapacity(rack.LivingStatefulSet().Spec.VolumeClaimTemplates)
		index, requested := findDataCapacity(newStatefulSet.Spec.VolumeClaimTemplates)
		if !requested.Equal(current) {
			dataPvcResources := &newStatefulSet.Spec.VolumeClaimTemplates[index].Spec.Resources
			if dataPvcResources.Requests == nil {
				dataPvcResources.Requests = corev1.ResourceList{}
			}
			dataPvcResources.Requests[corev1.ResourceStorage] = current
			rack.Log().
				Infof("Storage Resize request detected, postponing resize from %s to %s until other actions are done",
					requested.String(), current.String())
		}
	}
}
