package stub

import (
	api "github.com/cscetbon/casskop/api/v2"
	"github.com/cscetbon/casskop/controllers/cassandracluster/view"
	"github.com/cscetbon/casskop/pkg/k8s"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
)

type RackView struct {
	ClusterNameStub       string
	CompleteRackNameStub  api.CompleteRackName
	RackStatusStub        *api.CassandraRackStatus
	LivingStatefulSetStub *appsv1.StatefulSet
}

var _ view.RackView = RackView{}

func (v RackView) ClusterName() string {
	return v.ClusterNameStub
}

func (v RackView) DcName() api.DcName {
	return v.CompleteRackNameStub.DcName
}

func (v RackView) RackName() api.RackName {
	return v.CompleteRackNameStub.RackName
}

func (v RackView) DcRackName() api.DcRackName {
	return v.CompleteRackNameStub.DcRackName
}

func (v RackView) RackStatus() *api.CassandraRackStatus {
	return v.RackStatusStub
}

func (v RackView) LivingStatefulSet() *appsv1.StatefulSet {
	return v.LivingStatefulSetStub
}

func (v RackView) IsStatefulSetAliveNow() bool {
	return v.LivingStatefulSetStub != nil
}

func (v RackView) GetLabelsForCassandraDCRack(cc *api.CassandraCluster) map[string]string {
	return k8s.LabelsForCassandraDCRackStrongTypes(cc, v.DcName(), v.RackName())
}

func (v RackView) Log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{})
}
