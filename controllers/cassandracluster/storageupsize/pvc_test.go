package storageupsize

import (
	"context"
	"errors"
	"testing"

	api "github.com/cscetbon/casskop/api/v2"
	"github.com/cscetbon/casskop/controllers/cassandracluster/storagestateclient"
	storagestateclientmock "github.com/cscetbon/casskop/controllers/cassandracluster/storagestateclient/mock"
	rackviewstub "github.com/cscetbon/casskop/controllers/cassandracluster/view/stub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var testCtx = context.Background()

func Test_ensureAllPVCsHaveNewCapacity_happyPath(t *testing.T) {
	const expectedCapacity = "25Gi"

	dataPvc0 := pvc("data-dc-rack1-0", "10Gi")
	dataPvc1 := pvc("data-dc-rack1-1", "25Gi")
	dataPvc2 := pvc("data-dc-rack1-2", "")
	pvcs := []corev1.PersistentVolumeClaim{dataPvc0, dataPvc1, dataPvc2}
	cc := &api.CassandraCluster{
		Spec: api.CassandraClusterSpec{
			DataCapacity: expectedCapacity,
		},
	}
	rackView := rackviewstub.RackView{CompleteRackNameStub: api.CompleteRackName{DcName: "dc"}}
	fakeClientScheme := scheme.Scheme
	fakeClientScheme.AddKnownTypes(api.GroupVersion, cc)
	cl := fake.NewClientBuilder().
		WithScheme(fakeClientScheme).
		WithRuntimeObjects(&dataPvc0, &dataPvc1, &dataPvc2).
		WithStatusSubresource(cc).Build()
	storageStateClient := storagestateclient.New(cl)

	t.Run("should update all PVCs to new capacity", func(t *testing.T) {
		res := ensureAllPVCsHaveNewCapacity(testCtx, cc, pvcs, rackView, storageStateClient)

		assert.False(t, res.HasError())
		assert.NoError(t, res.Error())
		assertPvcCapacity(t, cl, "data-dc-rack1-0", expectedCapacity)
		assertPvcCapacity(t, cl, "data-dc-rack1-1", expectedCapacity)
		assertPvcCapacity(t, cl, "data-dc-rack1-2", expectedCapacity)
	})

	t.Run("should handle PVC already at expected capacity", func(t *testing.T) {
		dataPvc0 = pvc("data-dc-rack1-0", expectedCapacity)
		dataPvc1 = pvc("data-dc-rack1-1", expectedCapacity)
		dataPvc2 = pvc("data-dc-rack1-2", expectedCapacity)
		assert.NoError(t, cl.Update(testCtx, &dataPvc0))
		assert.NoError(t, cl.Update(testCtx, &dataPvc1))
		assert.NoError(t, cl.Update(testCtx, &dataPvc2))
		pvcs = []corev1.PersistentVolumeClaim{dataPvc0, dataPvc1, dataPvc2}

		res := ensureAllPVCsHaveNewCapacity(testCtx, cc, pvcs, rackView, storageStateClient)

		assert.False(t, res.HasError())
		assert.NoError(t, res.Error())
		assertPvcCapacity(t, cl, "data-dc-rack1-0", expectedCapacity)
		assertPvcCapacity(t, cl, "data-dc-rack1-1", expectedCapacity)
		assertPvcCapacity(t, cl, "data-dc-rack1-2", expectedCapacity)
	})
}

func Test_ensureAllPVCsHaveNewCapacity_errors(t *testing.T) {
	const expectedCapacity = "25Gi"

	dataPvc0 := pvc("data-dc-rack1-0", "10Gi")
	dataPvc1 := pvc("data-dc-rack1-1", "25Gi")
	dataPvc2 := pvc("data-dc-rack1-2", "")
	pvcs := []corev1.PersistentVolumeClaim{dataPvc0, dataPvc1, dataPvc2}
	cc := &api.CassandraCluster{
		Spec: api.CassandraClusterSpec{
			DataCapacity: expectedCapacity,
		},
	}
	rackView := rackviewstub.RackView{CompleteRackNameStub: api.CompleteRackName{DcName: "dc"}}

	storageStateClient := &storagestateclientmock.StorageStateClient{}
	storageStateClient.On("UpdatePVC", mock.Anything, mock.Anything).
		Return(errors.New("update failed")).Times(3)

	res := ensureAllPVCsHaveNewCapacity(testCtx, cc, pvcs, rackView, storageStateClient)

	assert.True(t, res.HasError())
	assert.EqualError(t, res.Error(), "update failed; update failed")
}

func pvc(name, capacity string) corev1.PersistentVolumeClaim {
	var resources corev1.ResourceList
	if capacity != "" {
		resources = corev1.ResourceList{
			corev1.ResourceStorage: resource.MustParse(capacity),
		}
	}
	return corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.VolumeResourceRequirements{
				Requests: resources,
			},
		},
	}
}

func assertPvcCapacity(t *testing.T, cl client.WithWatch, pvcName, expectedCapacity string) {
	updatedPvc := corev1.PersistentVolumeClaim{}
	err := cl.Get(testCtx, types.NamespacedName{Namespace: "default", Name: pvcName}, &updatedPvc)
	assert.NoError(t, err)
	finalDataCapacity := updatedPvc.Spec.Resources.Requests[corev1.ResourceStorage]
	assert.Equal(t, expectedCapacity, finalDataCapacity.String())
}
