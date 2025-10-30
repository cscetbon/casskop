package cassandracluster

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	api "github.com/cscetbon/casskop/api/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func stringOfSlice(a []string) string {
	q := make([]string, len(a))
	for i, s := range a {
		q[i] = fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("[%s]", strings.Join(q, ", "))
}

func registerJolokiaOperationJoiningNodes(host podName, numberOfJoiningNodes int) {
	joiningNodes := []string{}
	for i := 1; i <= numberOfJoiningNodes; i++ {
		joiningNodes = append(joiningNodes, "nodeX")
	}
	httpmock.RegisterResponder("POST", JolokiaURL(host.FullName, jolokiaPort),
		httpmock.NewStringResponder(200, fmt.Sprintf(`{"request":
											{"mbean": "org.apache.cassandra.db:type=StorageService",
											 "attribute": "joiningNodes",
											 "type": "read"},
										"value": %s,
										"timestamp": 1528850319,
										"status": 200}`, stringOfSlice(joiningNodes))))
}

func simulateNewPodsReady(t *testing.T, rcc *CassandraClusterReconciler, stfsName string, dc api.DC, scaleFrom, scaleTo int) {
	assert := assert.New(t)

	sts, err := rcc.GetStatefulSet(ctx, rcc.cc.Namespace, stfsName)
	assert.NoError(err, "get sts")

	//Now simulate sts to be ready for CassKop
	sts.Status.Replicas = *sts.Spec.Replicas
	sts.Status.ReadyReplicas = *sts.Spec.Replicas
	err = rcc.Client.Status().Update(ctx, sts)
	assert.NoError(err, "update sts status")

	// create new fake Pods (consequence of scale-out) so action may finish
	podTemplate := fakePodTemplate(rcc.cc, dc.Name, dc.Rack[0].Name)
	for i := scaleFrom; i < scaleTo; i++ {
		pod := podTemplate.DeepCopy()
		pod.Name = sts.Name + "-" + strconv.Itoa(i)
		pod.Spec.Hostname = pod.Name
		pod.Spec.Subdomain = rcc.cc.Name
		if err = rcc.CreatePod(ctx, pod); err != nil {
			t.Fatalf("can't create pod: (%v)", err)
		}
	}
}

func TestAddTwoNodes(t *testing.T) {
	overrideDelayWaitWithNoDelay()
	defer restoreDefaultDelayWait()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	assert := assert.New(t)

	rcc, req := createCassandraClusterWithNoDisruption(t, "cassandracluster-1DC.yaml")

	assert.Equal(int32(3), rcc.cc.Spec.NodesPerRacks)

	cassandraCluster := rcc.cc.DeepCopy()

	datacenters := cassandraCluster.Spec.Topology.DC
	assert.Equal(1, len(datacenters))
	assert.Equal(1, len(datacenters[0].Rack))

	dc := datacenters[0]
	stfsName := cassandraCluster.Name + fmt.Sprintf("-%s-%s", dc.Name, dc.Rack[0].Name)

	cassandraCluster.Spec.NodesPerRacks = 5
	rcc.Client.Update(context.TODO(), cassandraCluster)

	firstPod := podHost(stfsName, 0, rcc)
	reconcileValidation(t, rcc, *req)
	assert.GreaterOrEqual(jolokiaCallsCount(firstPod), 0)
	assertStatefulsetReplicas(ctx, t, rcc, 3, cassandraCluster.Namespace, stfsName)

	//Reconcile adds one node at a time when it's asked to add more than one node
	for expectedReplicas := 3; expectedReplicas <= 4; expectedReplicas++ {
		//Reconcile does not update the number of nodes when there are joining nodes
		registerJolokiaOperationJoiningNodes(firstPod, 1)
		for reconcileIteration := 0; reconcileIteration <= 2; reconcileIteration++ {
			reconcileValidation(t, rcc, *req)
			assert.GreaterOrEqual(jolokiaCallsCount(firstPod), 1)
			assertStatefulsetReplicas(ctx, t, rcc, expectedReplicas, cassandraCluster.Namespace, stfsName)
		}

		//Reconcile adds a node as soon as there are no longer joining nodes
		registerJolokiaOperationJoiningNodes(firstPod, 0)
		reconcileValidation(t, rcc, *req)
		assert.GreaterOrEqual(jolokiaCallsCount(firstPod), 1)
		assertStatefulsetReplicas(ctx, t, rcc, expectedReplicas+1, cassandraCluster.Namespace, stfsName)
	}

	assertClusterStatusPhase(assert, rcc, api.ClusterPhasePending)
	assertRackStatusPhase(assert, rcc, "dc1-rack1", api.ClusterPhasePending)
	assertClusterStatusLastAction(assert, rcc, api.ActionScaleUp, api.StatusOngoing)
	assertRackStatusLastAction(assert, rcc, "dc1-rack1", api.ActionScaleUp, api.StatusOngoing)

	//Reconcile does not end the action even when sts and all new pods are ready, because there are joining nodes
	simulateNewPodsReady(t, rcc, stfsName, dc, 3, 5)
	registerJolokiaOperationJoiningNodes(firstPod, 1)
	for reconcileIteration := 0; reconcileIteration <= 2; reconcileIteration++ {
		reconcileValidation(t, rcc, *req)
		assert.GreaterOrEqual(jolokiaCallsCount(firstPod), 1)
		assertClusterStatusPhase(assert, rcc, api.ClusterPhasePending)
		assertRackStatusPhase(assert, rcc, "dc1-rack1", api.ClusterPhaseRunning)
		assertClusterStatusLastAction(assert, rcc, api.ActionScaleUp, api.StatusOngoing)
		assertRackStatusLastAction(assert, rcc, "dc1-rack1", api.ActionScaleUp, api.StatusOngoing)
	}

	//Reconcile ends the action when sts and all new pods are ready and there are no joining nodes
	registerJolokiaOperationJoiningNodes(firstPod, 0)
	reconcileValidation(t, rcc, *req)
	assert.GreaterOrEqual(jolokiaCallsCount(firstPod), 1)
	assertClusterStatusPhase(assert, rcc, api.ClusterPhaseRunning)
	assertRackStatusPhase(assert, rcc, "dc1-rack1", api.ClusterPhaseRunning)
	assertClusterStatusLastAction(assert, rcc, api.ActionScaleUp, api.StatusDone)
	assertRackStatusLastAction(assert, rcc, "dc1-rack1", api.ActionScaleUp, api.StatusDone)
}
