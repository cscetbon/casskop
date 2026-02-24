package storageupsize

import (
	"context"
	"errors"
	"fmt"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	api "github.com/cscetbon/casskop/api/v2"
	"github.com/cscetbon/casskop/controllers/cassandracluster/consts"
	"github.com/cscetbon/casskop/controllers/cassandracluster/pods"
	"github.com/cscetbon/casskop/controllers/cassandracluster/storageupsize/actionstep"
	"github.com/cscetbon/casskop/controllers/cassandracluster/storageupsize/lastapplied"
	"github.com/cscetbon/casskop/controllers/cassandracluster/sts"
	"github.com/cscetbon/casskop/controllers/cassandracluster/view"
	json "github.com/json-iterator/go"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func prepareStatefulSetSnapshot(livingStatefulSet *appsv1.StatefulSet) (string, error) {
	statefulSetSnapshot := livingStatefulSet.DeepCopy()

	statefulSetSnapshot.GenerateName = ""
	statefulSetSnapshot.SelfLink = ""
	statefulSetSnapshot.UID = ""
	statefulSetSnapshot.ResourceVersion = ""
	statefulSetSnapshot.Generation = 0
	statefulSetSnapshot.CreationTimestamp = metav1.Time{}
	statefulSetSnapshot.DeletionTimestamp = nil
	statefulSetSnapshot.DeletionGracePeriodSeconds = nil
	statefulSetSnapshot.ManagedFields = nil

	statefulSetSnapshot.TypeMeta = metav1.TypeMeta{}
	statefulSetSnapshot.Status = appsv1.StatefulSetStatus{}

	statefulSetSnapshotJson, err := json.ConfigCompatibleWithStandardLibrary.Marshal(statefulSetSnapshot)
	if err != nil {
		return "", err
	}
	return string(statefulSetSnapshotJson), nil
}

func applyPVCModification(newStatefulSet *appsv1.StatefulSet, newDataCapacity resource.Quantity) error {
	err := setNewDataCapacity(newStatefulSet, newDataCapacity)
	if err != nil {
		return err
	}
	return enrichWithCleanLastAppliedAnnotation(newStatefulSet, newDataCapacity)
}

// enrichWithCleanLastAppliedAnnotation
//
// 1. Why not create statefulSet directly from the CR?
//   - because other changes may have been introduced (e.g. scaling or other changes in the pod template)
//     we don't want other actions to interfere with the resize process
//
// 2. Why not create statefulSet from the old CR (last-applied-configuration)?
//   - because other changes may also have been introduced
//     -- someone starts a scale-in
//     -- then changes dataCapacity
//     -- the first rack will still be done correctly,
//     but in the second, at the end of the resize,
//     a StatefulSet with a smaller number of replicas will be applied immediately, without calling decommission
//
// 3. Why do we need to manually handle last-applied annotation on the newStatefulSet?
//   - Banzai stores the original object in annotations and performs a 3-way merge on update
//   - livingStatefulSet is an object fetched from k8s API, so it contains Kubernetes defaults (added by the k8s API server)
//   - newStatefulSet is livingStatefulSet after marshal+unmarshal
//   - if we simply did
//     `patch.DefaultAnnotator.SetLastAppliedAnnotation(newStatefulSet)`
//     we would put into the annotations an object with Kubernetes defaults (added by the k8s API server)
//   - that would force an update during the 3-way merge after the resize
//     (StatefulSet generated from the CR would be clean and would not match the polluted last-applied in the stored StatefulSet)
func enrichWithCleanLastAppliedAnnotation(newStatefulSet *appsv1.StatefulSet, newDataCapacity resource.Quantity) error {

	// best effort = swallow all errors:
	// if we cannot get/edit/encode original sts -> we skip setting last-applied annotation, would lead to extra update after resize (no pod restart)

	originalStatefulSet, err := lastapplied.GetOriginalSts(newStatefulSet)
	if err != nil {
		return nil
	}

	err = setNewDataCapacity(&originalStatefulSet, newDataCapacity)
	if err != nil {
		return nil
	}

	lastApplied, err := lastapplied.EncodeLastAppliedConfigAnnotation(originalStatefulSet)
	if err != nil {
		return nil
	}
	newStatefulSet.Annotations[patch.LastAppliedConfig] = lastApplied

	return nil
}

func removeStatefulSetOrphan(ctx context.Context, cc *api.CassandraCluster, rack view.RackView, stsClient sts.StsClient) actionstep.StepResult {
	if !rack.IsStatefulSetAliveNow() {
		return actionstep.Pass()
	}

	if doesStatefulSetHaveNewCapacity(cc, rack.LivingStatefulSet()) {
		return actionstep.Pass()
	}

	rack.Log().Info("Deleting StatefulSet with orphan option")
	err := stsClient.DeleteStatefulSetWithOrphanOption(ctx, cc.Namespace, rack.LivingStatefulSet().Name)
	if err != nil {
		return actionstep.Error(err)
	}

	return actionstep.Break()
}

func doesStatefulSetHaveNewCapacity(cc *api.CassandraCluster, livingStatefulSet *appsv1.StatefulSet) bool {
	requested := silentParseResourceQuantity(cc.Spec.DataCapacity)
	_, current := findDataCapacity(livingStatefulSet.Spec.VolumeClaimTemplates)
	return requested.Equal(current)
}

func recreateStatefulSetWithNewCapacity(ctx context.Context, rack view.RackView, newDataCapacity resource.Quantity,
	stsClient sts.StsClient) actionstep.StepResult {

	if rack.IsStatefulSetAliveNow() {
		return actionstep.Pass()
	}

	rack.Log().Info("Creating StatefulSet with new capacity")

	newStatefulSet, err := unmarshallSnapshottedStatefulSet(rack)
	if err != nil {
		return actionstep.Error(err)
	}

	err = applyPVCModification(newStatefulSet, newDataCapacity)
	if err != nil {
		return actionstep.Error(err)
	}

	err = stsClient.CreateStatefulSet(ctx, newStatefulSet)
	if err != nil {
		return actionstep.Error(err)
	}

	return actionstep.Break()
}

func waitTillStatefulSetAndAllPodsAreReady(ctx context.Context, cc *api.CassandraCluster, rack view.RackView,
	podsClient pods.PodsClient) actionstep.StepResult {

	if !doesStatefulSetHaveNewCapacity(cc, rack.LivingStatefulSet()) {
		rack.Log().Infof("Resize action is in progress, statefulset need to be re-created with new capacity")
		return actionstep.Break()
	}

	if sts.IsStatefulSetReady(rack.LivingStatefulSet()) {
		podList, err := podsClient.ListPods(ctx, cc.Namespace, rack.GetLabelsForCassandraDCRack(cc))
		if err != nil {
			return actionstep.Error(err)
		}
		expectedNodesPerRacks := *rack.LivingStatefulSet().Spec.Replicas
		if len(podList.Items) != int(expectedNodesPerRacks) {
			errMsg := fmt.Sprintf("Number of pods (%d) different than expected Replicas (%d) for DC-Rack %s",
				len(podList.Items), expectedNodesPerRacks, rack.DcRackName())
			rack.Log().Warn(errMsg)
			return actionstep.Error(errors.New(errMsg))
		}
		if allPodsReady(podList) {
			rack.Log().Info("Resize action finalization, " +
				"all pods are ready with new DataCapacity, we can finalize the action")
			finalizeUpsizeAction(rack.RackStatus())
			return actionstep.Pass()
		}
	}

	rack.Log().Info("Resize action is in progress, " +
		"we wait for all pods to be ready with new DataCapacity before finalizing the action")
	return actionstep.Break()
}

func setNewDataCapacity(statefulSet *appsv1.StatefulSet, dataCapacity resource.Quantity) error {
	for i, template := range statefulSet.Spec.VolumeClaimTemplates {
		if template.Name == consts.DataPVCName {
			template.Spec.Resources.Requests["storage"] = dataCapacity
			statefulSet.Spec.VolumeClaimTemplates[i] = template
			return nil
		}
	}
	return errors.New(fmt.Sprintf("no %s pvc found in statefulSet %s", consts.DataPVCName, statefulSet.Name))
}
