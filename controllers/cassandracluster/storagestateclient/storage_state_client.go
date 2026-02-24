package storagestateclient

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func New(client client.Client) StorageStateClient {
	return &storageStateClient{
		k8sClient: client,
	}
}

type StorageStateClient interface {
	GetPVC(ctx context.Context, namespace, name string) (*v1.PersistentVolumeClaim, error)
	ListPVC(ctx context.Context, namespace string, selector map[string]string) (*v1.PersistentVolumeClaimList, error)
	UpdatePVC(ctx context.Context, pvc *v1.PersistentVolumeClaim) error
}

var _ StorageStateClient = (*storageStateClient)(nil)

type storageStateClient struct {
	k8sClient client.Client
}

func (c *storageStateClient) GetPVC(ctx context.Context, namespace, name string) (*v1.PersistentVolumeClaim, error) {
	o := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return o, c.k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, o)
}

func (c *storageStateClient) ListPVC(ctx context.Context, namespace string,
	selector map[string]string) (*v1.PersistentVolumeClaimList, error) {

	clientOpt := &client.ListOptions{Namespace: namespace, LabelSelector: labels.SelectorFromSet(selector)}
	opt := []client.ListOption{
		clientOpt,
	}

	o := &v1.PersistentVolumeClaimList{}

	return o, c.k8sClient.List(ctx, o, opt...)
}

func (c *storageStateClient) UpdatePVC(ctx context.Context, pvc *v1.PersistentVolumeClaim) error {
	return c.k8sClient.Update(ctx, pvc)
}
