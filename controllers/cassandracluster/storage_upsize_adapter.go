package cassandracluster

import (
	"context"

	api "github.com/cscetbon/casskop/api/v2"
	"github.com/cscetbon/casskop/controllers/cassandracluster/pods"
	"github.com/cscetbon/casskop/controllers/cassandracluster/storagestateclient"
	"github.com/cscetbon/casskop/controllers/cassandracluster/storageupsize"
	"github.com/cscetbon/casskop/controllers/cassandracluster/sts"
	"github.com/cscetbon/casskop/controllers/cassandracluster/view"
	"github.com/cscetbon/casskop/pkg/k8s"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
)

func (rcc *CassandraClusterReconciler) RevertAnyStorageUpsizeBeyondUpsizeAction(completeDcRackName api.CompleteRackName,
	dcRackStatus *api.CassandraRackStatus, statefulSet *appsv1.StatefulSet) {

	storageupsize.RevertAnyStorageUpsizeBeyondUpsizeAction(rcc.newRackView(completeDcRackName, dcRackStatus), statefulSet)
}

func (rcc *CassandraClusterReconciler) UpdateStatusIfStorageUpsize(completeDcRackName api.CompleteRackName,
	status *api.CassandraClusterStatus) bool {

	rackView := rcc.newRackView(completeDcRackName, status.GetCassandraRackStatus(completeDcRackName.DcRackName))
	if rcc.shouldStorageUpsizeBeStarted(rackView) {
		rcc.startStorageUpsize(rackView)
		return true
	}
	return false
}

func (rcc *CassandraClusterReconciler) shouldStorageUpsizeBeStarted(rackView view.RackView) bool {
	return storageupsize.ShouldBeStarted(rackView, rcc.cc.GetDataCapacityForDCName(rackView.DcName()))
}

func (rcc *CassandraClusterReconciler) startStorageUpsize(rackView view.RackView) {
	storageupsize.Start(rackView)
	ClusterPhaseMetric.set(api.ClusterPhasePending, rcc.cc.Name)
	ClusterActionMetric.set(api.ActionStorageUpsize, rcc.cc.Name)
}

func (rcc *CassandraClusterReconciler) IsStorageUpsizeStarted(dcRackStatus *api.CassandraRackStatus) bool {
	return storageupsize.IsStarted(dcRackStatus)
}

func (rcc *CassandraClusterReconciler) ReconcileStorageUpsize(ctx context.Context, cc *api.CassandraCluster,
	status *api.CassandraClusterStatus, completeDcRackName api.CompleteRackName) error {

	status.GetCassandraRackStatus(completeDcRackName.DcRackName).Phase = api.ClusterPhasePending.Name
	ClusterPhaseMetric.set(api.ClusterPhasePending, cc.Name)

	newDataCapacity := generateResourceQuantity(cc.GetDataCapacityForDCName(completeDcRackName.DcName))

	rackView := rcc.newRackView(completeDcRackName, status.GetCassandraRackStatus(completeDcRackName.DcRackName))
	var storageStateClient storagestateclient.StorageStateClient = rcc
	var stsClient sts.StsClient = rcc
	var podsClient pods.PodsClient = rcc

	return storageupsize.Reconcile(ctx, cc, rackView, newDataCapacity, storageStateClient, stsClient, podsClient)
}

func (rcc *CassandraClusterReconciler) newRackView(completeRackName api.CompleteRackName,
	dcRackStatus *api.CassandraRackStatus) view.RackView {

	return &rccRackView{
		rcc:              rcc,
		completeRackName: completeRackName,
		dcRackStatus:     dcRackStatus,
	}
}

type rccRackView struct {
	rcc              *CassandraClusterReconciler
	completeRackName api.CompleteRackName
	dcRackStatus     *api.CassandraRackStatus
}

var _ view.RackView = (*rccRackView)(nil)

func (v *rccRackView) ClusterName() string {
	return v.rcc.cc.Name
}

func (v *rccRackView) DcName() api.DcName {
	return v.completeRackName.DcName
}

func (v *rccRackView) RackName() api.RackName {
	return v.completeRackName.RackName
}

func (v *rccRackView) DcRackName() api.DcRackName {
	return v.completeRackName.DcRackName
}

func (v *rccRackView) RackStatus() *api.CassandraRackStatus {
	return v.dcRackStatus
}

func (v *rccRackView) LivingStatefulSet() *appsv1.StatefulSet {
	return v.rcc.storedStatefulSet
}

func (v *rccRackView) IsStatefulSetAliveNow() bool {
	return v.rcc.storedStatefulSet != nil
}

func (v *rccRackView) GetLabelsForCassandraDCRack(cc *api.CassandraCluster) map[string]string {
	return k8s.LabelsForCassandraDCRackStrongTypes(cc, v.DcName(), v.RackName())
}

func (v *rccRackView) Log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{"cluster": v.ClusterName(), "dc-rack": v.DcRackName()})
}
