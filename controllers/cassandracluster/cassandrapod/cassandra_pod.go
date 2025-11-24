package cassandrapod

import (
	"github.com/cscetbon/casskop/controllers/cassandracluster/consts"
	v1 "k8s.io/api/core/v1"
)

func IsReady(pod *v1.Pod) bool {
	cassandraContainerStatus := getContainerStatus(pod)
	if cassandraContainerStatus != nil && pod.Status.Phase == v1.PodRunning && cassandraContainerStatus.Ready {
		return true
	}
	return false
}

func getContainerStatus(pod *v1.Pod) *v1.ContainerStatus {
	for i := range pod.Status.ContainerStatuses {
		if pod.Status.ContainerStatuses[i].Name == consts.CassandraContainerName {
			return &pod.Status.ContainerStatuses[i]
		}
	}
	return nil
}

func RestartCount(pod *v1.Pod) int32 {
	for idx := range pod.Status.ContainerStatuses {
		if pod.Status.ContainerStatuses[idx].Name == consts.CassandraContainerName {
			return pod.Status.ContainerStatuses[idx].RestartCount
		}
	}
	return 0
}
