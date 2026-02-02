package testfixtures

import api "github.com/cscetbon/casskop/api/v2"

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
	InitializingSubPhase: nil,
}
