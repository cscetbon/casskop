package storageupsize

import (
	"github.com/cscetbon/casskop/controllers/cassandracluster/cassandrapod"
	v1 "k8s.io/api/core/v1"
)

func allPodsReady(podList *v1.PodList) bool {
	for _, pod := range podList.Items {
		if !cassandrapod.IsReady(&pod) {
			return false
		}
	}
	return true
}
