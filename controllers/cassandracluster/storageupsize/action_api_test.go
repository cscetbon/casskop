package storageupsize

import (
	"testing"

	v2 "github.com/cscetbon/casskop/api/v2"
	"github.com/cscetbon/casskop/controllers/cassandracluster/consts"
	"github.com/cscetbon/casskop/controllers/cassandracluster/view/stub"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestRevertAnyStorageUpsizeBeyondUpsizeAction(t *testing.T) {

	const InitialCapacity = "10Gi"
	const CapacityAfterUpsize = "999Gi"

	t.Run("upsize action in-progress, do not revert upsize", func(t *testing.T) {
		currentSts := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			pvc(consts.DataPVCName, InitialCapacity),
		}}}
		newSts := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			pvc(consts.DataPVCName, CapacityAfterUpsize),
		}}}
		rack := stub.RackView{
			LivingStatefulSetStub: currentSts.DeepCopy(),
			RackStatusStub: &v2.CassandraRackStatus{
				CassandraLastAction: v2.CassandraLastAction{
					Name:   v2.ActionStorageUpsize.Name,
					Status: v2.StatusOngoing,
				},
			},
		}

		newStsWithRevertApplied := newSts.DeepCopy()
		RevertAnyStorageUpsizeBeyondUpsizeAction(rack, newStsWithRevertApplied)

		assert.Equal(t, resource.MustParse(CapacityAfterUpsize),
			newStsWithRevertApplied.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])
	})

	t.Run("upsize action in-progress, sts resized already, nothing to do", func(t *testing.T) {
		currentSts := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			pvc(consts.DataPVCName, CapacityAfterUpsize),
		}}}
		newSts := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			pvc(consts.DataPVCName, CapacityAfterUpsize),
		}}}
		rack := stub.RackView{
			LivingStatefulSetStub: currentSts.DeepCopy(),
			RackStatusStub: &v2.CassandraRackStatus{
				CassandraLastAction: v2.CassandraLastAction{
					Name:   v2.ActionStorageUpsize.Name,
					Status: v2.StatusOngoing,
				},
			},
		}

		newStsWithRevertApplied := newSts.DeepCopy()
		RevertAnyStorageUpsizeBeyondUpsizeAction(rack, newStsWithRevertApplied)

		assert.Equal(t, resource.MustParse(CapacityAfterUpsize),
			newStsWithRevertApplied.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])
	})

	t.Run("no upsize requested, nothing changes", func(t *testing.T) {
		currentSts := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			pvc(consts.DataPVCName, InitialCapacity),
		}}}
		newSts := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			pvc(consts.DataPVCName, InitialCapacity),
		}}}
		rack := stub.RackView{
			LivingStatefulSetStub: currentSts.DeepCopy(),
			RackStatusStub: &v2.CassandraRackStatus{
				Phase: v2.ClusterPhaseRunning.Name,
			},
		}

		newStsWithRevertApplied := newSts.DeepCopy()
		RevertAnyStorageUpsizeBeyondUpsizeAction(rack, newStsWithRevertApplied)

		assert.Equal(t, resource.MustParse(InitialCapacity),
			newStsWithRevertApplied.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])
	})

	t.Run("upsize requested but action not started yet, revert change", func(t *testing.T) {
		currentSts := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			pvc(consts.DataPVCName, InitialCapacity),
		}}}
		newSts := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			pvc(consts.DataPVCName, CapacityAfterUpsize),
		}}}
		rack := stub.RackView{
			LivingStatefulSetStub: currentSts.DeepCopy(),
			RackStatusStub: &v2.CassandraRackStatus{
				CassandraLastAction: v2.CassandraLastAction{
					Name:   v2.ActionScaleUp.Name,
					Status: v2.StatusOngoing,
				},
			},
		}

		newStsWithRevertApplied := newSts.DeepCopy()
		RevertAnyStorageUpsizeBeyondUpsizeAction(rack, newStsWithRevertApplied)

		assert.Equal(t, resource.MustParse(InitialCapacity),
			newStsWithRevertApplied.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])
	})

	t.Run("upsize requested but action not started yet (previous upsize done), revert change", func(t *testing.T) {
		currentSts := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			pvc(consts.DataPVCName, InitialCapacity),
		}}}
		newSts := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			pvc(consts.DataPVCName, CapacityAfterUpsize),
		}}}
		rack := stub.RackView{
			LivingStatefulSetStub: currentSts.DeepCopy(),
			RackStatusStub: &v2.CassandraRackStatus{
				CassandraLastAction: v2.CassandraLastAction{
					Name:   v2.ActionStorageUpsize.Name,
					Status: v2.StatusDone,
				},
			},
		}

		newStsWithRevertApplied := newSts.DeepCopy()
		RevertAnyStorageUpsizeBeyondUpsizeAction(rack, newStsWithRevertApplied)

		assert.Equal(t, resource.MustParse(InitialCapacity),
			newStsWithRevertApplied.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])
	})

	t.Run("handle unspecified resources", func(t *testing.T) {
		currentSts := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			pvc(consts.DataPVCName, InitialCapacity),
		}}}
		newSts := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			pvcWithoutSpecifiedResources(consts.DataPVCName),
		}}}
		rack := stub.RackView{
			LivingStatefulSetStub: currentSts.DeepCopy(),
			RackStatusStub: &v2.CassandraRackStatus{
				CassandraLastAction: v2.CassandraLastAction{
					Name:   v2.ActionStorageUpsize.Name,
					Status: v2.StatusDone,
				},
			},
		}

		newStsWithRevertApplied := newSts.DeepCopy()
		RevertAnyStorageUpsizeBeyondUpsizeAction(rack, newStsWithRevertApplied)

		assert.Equal(t, resource.MustParse(InitialCapacity),
			newStsWithRevertApplied.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])
	})
}
