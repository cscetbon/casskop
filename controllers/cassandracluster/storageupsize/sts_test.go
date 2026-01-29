package storageupsize

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	v2 "github.com/cscetbon/casskop/api/v2"
	"github.com/cscetbon/casskop/controllers/cassandracluster/consts"
	"github.com/cscetbon/casskop/controllers/cassandracluster/sts"
	"github.com/cscetbon/casskop/controllers/cassandracluster/view/stub"
	json "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func Test_applyPVCModification(t *testing.T) {

	const OldStsKey = "banzaicloud.com/last-applied"
	const InitialCapacity = "5Gi"
	const CapacityAfterUpsize = "10Gi"
	getStsBeforeChange := func() *appsv1.StatefulSet {
		return &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			pvc(consts.DataPVCName, InitialCapacity),
		}}}
	}

	t.Run("no data pvc - error expected", func(t *testing.T) {
		statefulSet := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "dc1-rack1"}}

		err := applyPVCModification(statefulSet, resource.MustParse(CapacityAfterUpsize))

		assert.EqualError(t, err, "no data pvc found in statefulSet dc1-rack1")
	})

	t.Run("no old sts - should just apply new capacity", func(t *testing.T) {
		statefulSet := getStsBeforeChange()

		err := applyPVCModification(statefulSet, resource.MustParse(CapacityAfterUpsize))

		assert.NoError(t, err)
		assert.Equal(t, resource.MustParse(CapacityAfterUpsize),
			statefulSet.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])
		assert.Empty(t, statefulSet.Annotations[OldStsKey])
	})

	t.Run("old sts malformed - should just apply new capacity", func(t *testing.T) {
		statefulSet := getStsBeforeChange()
		statefulSet.Annotations = map[string]string{
			OldStsKey: "malformed-annotation",
		}

		err := applyPVCModification(statefulSet, resource.MustParse(CapacityAfterUpsize))

		assert.NoError(t, err)
		assert.Equal(t, resource.MustParse(CapacityAfterUpsize),
			statefulSet.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])
		assert.Equal(t, "malformed-annotation", statefulSet.Annotations[OldStsKey])
	})

	t.Run("old sts without data pvc - should just apply new capacity", func(t *testing.T) {
		oldStsWithoutDataPvc := string(toJson(t, &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: "dc1-rack1"},
			Spec: appsv1.StatefulSetSpec{
				VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
					pvc("some-other-pvc1", "1Gi"),
					pvc("some-other-pvc2", "1Gi"),
				},
			},
		}))
		statefulSet := getStsBeforeChange()
		statefulSet.Annotations = map[string]string{
			OldStsKey: oldStsWithoutDataPvc,
		}

		err := applyPVCModification(statefulSet, resource.MustParse(CapacityAfterUpsize))

		assert.NoError(t, err)
		assert.Equal(t, resource.MustParse(CapacityAfterUpsize),
			statefulSet.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])
		assert.Equal(t, oldStsWithoutDataPvc, unzip(statefulSet.Annotations[OldStsKey]))
	})

	t.Run("old sts exists - should apply new capacity to current AND old spec", func(t *testing.T) {
		statefulSet := getStsBeforeChange()
		statefulSet.Annotations = map[string]string{
			OldStsKey: string(toJson(t, statefulSet)),
		}

		err := applyPVCModification(statefulSet, resource.MustParse(CapacityAfterUpsize))

		assert.NoError(t, err)
		assert.Equal(t, resource.MustParse(CapacityAfterUpsize),
			statefulSet.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])
		assert.Equal(t, string(toJson(t, removeAnnotations(statefulSet.DeepCopy()))), unzip(statefulSet.Annotations[OldStsKey]))
	})
}

func Test_recreateStatefulSetWithNewCapacity(t *testing.T) {

	t.Run("statefulSet already exists - no op", func(t *testing.T) {
		rack := stub.RackView{
			LivingStatefulSetStub: &appsv1.StatefulSet{},
		}

		result := recreateStatefulSetWithNewCapacity(testCtx, rack, resource.MustParse("15Gi"), nil)

		assert.False(t, result.HasError())
		assert.NoError(t, result.Error())
		assert.False(t, result.ShouldBreakReconcileLoop())
	})

	t.Run("statefulSet snapshot not exist - error", func(t *testing.T) {
		rack := stub.RackView{
			RackStatusStub: &v2.CassandraRackStatus{
				StatefulSetSnapshotBeforeStorageResize: "",
			},
		}

		result := recreateStatefulSetWithNewCapacity(testCtx, rack, resource.MustParse("15Gi"), nil)

		assert.True(t, result.HasError())
		assert.Contains(t, result.Error().Error(), "cannot unmarshall snapshotted statefulSet for storage upsize")
		assert.True(t, result.ShouldBreakReconcileLoop())
	})

	t.Run("cannot find data PVC - error", func(t *testing.T) {
		stsSnapshot := string(toJson(t, &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: "dc1-rack1", Namespace: "default"},
			Spec: appsv1.StatefulSetSpec{
				VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
					pvc("some other pvc", "9Gi"),
				},
			},
		}))
		rack := stub.RackView{
			RackStatusStub: &v2.CassandraRackStatus{
				StatefulSetSnapshotBeforeStorageResize: stsSnapshot,
			},
		}

		result := recreateStatefulSetWithNewCapacity(testCtx, rack, resource.MustParse("15Gi"), nil)

		assert.True(t, result.HasError())
		assert.Contains(t, result.Error().Error(), "no data pvc found in statefulSet dc1-rack1")
		assert.True(t, result.ShouldBreakReconcileLoop())
	})

	t.Run("statefulSet created successfully - break loop, sts should be created with proper capacity", func(t *testing.T) {
		stsSnapshot := string(toJson(t, &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: "dc1-rack1", Namespace: "default"},
			Spec: appsv1.StatefulSetSpec{
				VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
					pvc(consts.DataPVCName, "9Gi"),
				},
			},
		}))
		rack := stub.RackView{
			RackStatusStub: &v2.CassandraRackStatus{
				StatefulSetSnapshotBeforeStorageResize: stsSnapshot,
			},
		}
		cl := fake.NewClientBuilder().Build()

		result := recreateStatefulSetWithNewCapacity(testCtx, rack, resource.MustParse("15Gi"), sts.NewClient(cl))

		assert.False(t, result.HasError())
		assert.NoError(t, result.Error())
		assert.True(t, result.ShouldBreakReconcileLoop())
		createdSts := &appsv1.StatefulSet{}
		assert.NoError(t, cl.Get(testCtx, types.NamespacedName{Namespace: "default", Name: "dc1-rack1"}, createdSts))
		assert.Equal(t, resource.MustParse("15Gi"),
			createdSts.Spec.VolumeClaimTemplates[0].Spec.Resources.Requests[corev1.ResourceStorage])
	})

	t.Run("statefulSet creation fails - error", func(t *testing.T) {
		stsSnapshot := string(toJson(t, &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: "dc1-rack1", Namespace: "default"},
			Spec: appsv1.StatefulSetSpec{
				VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
					pvc(consts.DataPVCName, "9Gi"),
				},
			},
		}))
		rack := stub.RackView{
			RackStatusStub: &v2.CassandraRackStatus{
				StatefulSetSnapshotBeforeStorageResize: stsSnapshot,
			},
		}
		cl := interceptor.NewClient(fake.NewClientBuilder().Build(), interceptor.Funcs{
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				return errors.New("creation failed")
			},
		})

		result := recreateStatefulSetWithNewCapacity(testCtx, rack, resource.MustParse("15Gi"), sts.NewClient(cl))

		assert.True(t, result.HasError())
		assert.Contains(t, result.Error().Error(), "creation failed")
		assert.True(t, result.ShouldBreakReconcileLoop())
		assert.True(t, apierrors.IsNotFound(cl.Get(
			testCtx,
			types.NamespacedName{Namespace: "default", Name: "dc1-rack1"},
			&appsv1.StatefulSet{},
		)))
	})
}

func removeAnnotations(sts *appsv1.StatefulSet) *appsv1.StatefulSet {
	sts.Annotations = map[string]string{}
	return sts
}

func toJson(t *testing.T, sts *appsv1.StatefulSet) []byte {
	marshalled, err := json.ConfigCompatibleWithStandardLibrary.Marshal(sts)
	require.NoError(t, err)
	out, _, err := patch.DeleteNullInJson(marshalled)
	require.NoError(t, err)
	return out
}

func unzip(in string) string {
	decoded, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return in
	}

	reader, err := zip.NewReader(bytes.NewReader(decoded), int64(len(decoded)))
	if err != nil {
		return in
	}

	if len(reader.File) == 0 {
		return in
	}

	file := reader.File[0]
	rc, err := file.Open()
	if err != nil {
		return in
	}
	defer rc.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(rc)
	if err != nil {
		return in
	}

	return buf.String()
}
