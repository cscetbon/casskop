package storageupsize

import (
	"context"
	"errors"
	"fmt"
	"strings"

	api "github.com/cscetbon/casskop/api/v2"
	"github.com/cscetbon/casskop/controllers/cassandracluster/consts"
	"github.com/cscetbon/casskop/controllers/cassandracluster/storagestateclient"
	"github.com/cscetbon/casskop/controllers/cassandracluster/storageupsize/actionstep"
	"github.com/cscetbon/casskop/controllers/cassandracluster/view"
	"github.com/sirupsen/logrus"
	"go.uber.org/multierr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func findDataCapacity(pvc []corev1.PersistentVolumeClaim) (int, resource.Quantity) {
	for i, template := range pvc {
		if template.Name == consts.DataPVCName {
			if template.Spec.Resources.Requests == nil {
				return i, resource.Quantity{}
			}
			return i, template.Spec.Resources.Requests[corev1.ResourceStorage]
		}
	}
	return -1, resource.Quantity{}
}

func silentParseResourceQuantity(qs string) resource.Quantity {
	q, _ := resource.ParseQuantity(qs)
	return q
}

func fetchDataPvcs(ctx context.Context, cc *api.CassandraCluster, rack view.RackView,
	storageStateClient storagestateclient.StorageStateClient, outputPVCs *[]corev1.PersistentVolumeClaim) actionstep.StepResult {

	dataPVCs, err := getAllDataPvcs(ctx, cc, rack, storageStateClient)
	if err != nil {
		return actionstep.Error(err)
	}
	*outputPVCs = dataPVCs
	return actionstep.Pass()
}

func getAllDataPvcs(ctx context.Context, cc *api.CassandraCluster, rack view.RackView,
	storageStateClient storagestateclient.StorageStateClient) ([]corev1.PersistentVolumeClaim, error) {

	if rack.LivingStatefulSet() == nil {
		return nil, errors.New(fmt.Sprintf("[%s]: cannot fetch PVC list for storage upsize because"+
			"livingStatefulSet is nil for DC-Rack %s "+
			"(should not see this message, PVC should not be listed before statefulSet is recreated with new capacity)",
			cc.Name, rack.DcRackName()))
	}
	statefulSetName := rack.LivingStatefulSet().Name

	pvcs, err := storageStateClient.ListPVC(ctx, cc.Namespace, rack.GetLabelsForCassandraDCRack(cc))
	if err != nil {
		return nil, err
	}

	dataPVCs := make([]corev1.PersistentVolumeClaim, 0)
	for _, pvc := range pvcs.Items {
		if strings.HasPrefix(pvc.Name, consts.DataPVCName+"-"+statefulSetName) {
			dataPVCs = append(dataPVCs, pvc)
		}
	}

	expectedNodesPerRacks := *rack.LivingStatefulSet().Spec.Replicas

	if len(dataPVCs) != int(expectedNodesPerRacks) {
		errMsg := fmt.Sprintf("[%s]: Number of Data PVCs (%d) different than expected Replicas (%d) for DC-Rack %s",
			cc.Name, len(dataPVCs), expectedNodesPerRacks, rack.DcRackName())
		logrus.Warn(errMsg)
		return nil, errors.New(errMsg)
	}

	return dataPVCs, nil
}

func ensureAllPVCsHaveNewCapacity(ctx context.Context, cc *api.CassandraCluster, dataPVCs []corev1.PersistentVolumeClaim,
	rack view.RackView, storageStateClient storagestateclient.StorageStateClient) actionstep.StepResult {

	requestedCapacity := silentParseResourceQuantity(cc.GetDataCapacityForDCName(rack.DcName()))

	anythingChanged := false
	var multiError error
	for _, pvc := range dataPVCs {
		if pvc.Spec.Resources.Requests["storage"] != requestedCapacity {
			anythingChanged = true
			if pvc.Spec.Resources.Requests == nil {
				pvc.Spec.Resources.Requests = corev1.ResourceList{}
			}
			pvc.Spec.Resources.Requests["storage"] = requestedCapacity
			err := storageStateClient.UpdatePVC(ctx, &pvc)
			if err != nil {
				rack.Log().Errorf("Error updating PVC[%s] capacity to %v", pvc.Name, requestedCapacity)
				multiError = multierr.Append(multiError, err)
			}
			rack.Log().Infof("Update PVC[%s] capacity to %v successful", pvc.Name, requestedCapacity)
		}
	}

	if multiError != nil {
		return actionstep.Error(multiError)
	}

	if anythingChanged {
		return actionstep.Break()
	}
	return actionstep.Pass()
}

func waitTillAllFilesystemsHaveNewCapacity(cc *api.CassandraCluster, dataPVCs []corev1.PersistentVolumeClaim,
	rack view.RackView) actionstep.StepResult {

	requestedCapacity := silentParseResourceQuantity(cc.GetDataCapacityForDCName(rack.DcName()))

	var resized, notResizedYet []string

	for _, pvc := range dataPVCs {
		if pvc.Status.Capacity["storage"].Equal(requestedCapacity) {
			resized = append(resized, pvc.Name)
		} else {
			notResizedYet = append(notResizedYet, pvc.Name)
		}
	}

	if len(notResizedYet) == 0 {
		rack.Log().Infof("All PVs for DC-Rack %s resized to %s", rack.DcRackName(), requestedCapacity.String())
		return actionstep.Pass()
	}

	rack.Log().Infof("Still waiting for PVs to be resized for DC-Rack %s. Resized: [%v], Not resized yet: [%v]",
		rack.DcRackName(), resized, notResizedYet)
	return actionstep.Break()
}
