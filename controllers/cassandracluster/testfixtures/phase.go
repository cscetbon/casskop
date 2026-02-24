package testfixtures

import (
	api "github.com/cscetbon/casskop/api/v2"
	"k8s.io/utils/ptr"
)

var RunningPhase = api.CassandraPhase{
	Phase:                api.ClusterPhaseRunning.Name,
	InitializingSubPhase: nil,
}

var PendingPhase = api.CassandraPhase{
	Phase:                api.ClusterPhasePending.Name,
	InitializingSubPhase: nil,
}

var InitialPhase = api.CassandraPhase{
	Phase:                api.ClusterPhaseInitial.Name,
	InitializingSubPhase: ptr.To(api.ClusterPhaseInitialSubPhaseFirstPodPerRack),
}

var FirstPodPerRackReadyPhase = api.CassandraPhase{
	Phase:                api.ClusterPhaseInitial.Name,
	InitializingSubPhase: ptr.To(api.ClusterPhaseInitialSubPhaseNextPodPerRack),
}
