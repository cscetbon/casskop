package sts

import appsv1 "k8s.io/api/apps/v1"

func IsStatefulSetNotReady(statefulSet *appsv1.StatefulSet) bool {
	return !IsStatefulSetReady(statefulSet)
}

func IsStatefulSetReady(statefulSet *appsv1.StatefulSet) bool {
	return statefulSet.Status.ReadyReplicas == *statefulSet.Spec.Replicas
}
