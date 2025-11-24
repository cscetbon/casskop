package sts

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewClient(client client.Client) StsClient {
	return &stsClient{client: client}
}

type StsClient interface {
	CreateStatefulSet(ctx context.Context, statefulSet *appsv1.StatefulSet) error
	DeleteStatefulSetWithOrphanOption(ctx context.Context, namespace, name string) error
}

var _ StsClient = (*stsClient)(nil)

type stsClient struct {
	client client.Client
}

func (c *stsClient) CreateStatefulSet(ctx context.Context, statefulSet *appsv1.StatefulSet) error {
	err := c.client.Create(ctx, statefulSet)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("statefulset already exists: %v", err)
		}
		return fmt.Errorf("failed to create cassandra statefulset: %v", err)
	}
	return nil
}

func (c *stsClient) DeleteStatefulSetWithOrphanOption(ctx context.Context, namespace, name string) error {
	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return c.client.Delete(ctx, ss, client.PropagationPolicy(metav1.DeletePropagationOrphan))
}
