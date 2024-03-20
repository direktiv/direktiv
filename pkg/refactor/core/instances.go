package core

import (
	"context"

	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
)

type InstanceManager interface {
	CancelInstance(ctx context.Context, namespace, instanceID string) error
	StartInstance(ctx context.Context, namespace, path string, input []byte) (*instancestore.InstanceData, error)
}

func NewInstanceManager(starter func(ctx context.Context, namespace, path string, input []byte) (*instancestore.InstanceData, error), canceller func(ctx context.Context, namespace, instanceID string) error) InstanceManager {
	return &instanceManager{
		canceller: canceller,
		start:     starter,
	}
}

type instanceManager struct {
	canceller func(ctx context.Context, namespace, instanceID string) error
	start     func(ctx context.Context, namespace, path string, input []byte) (*instancestore.InstanceData, error)
}

func (mgr *instanceManager) StartInstance(ctx context.Context, namespace, path string, input []byte) (*instancestore.InstanceData, error) {
	return mgr.start(ctx, namespace, path, input)
}

func (mgr *instanceManager) CancelInstance(ctx context.Context, namespace, instanceID string) error {
	return mgr.canceller(ctx, namespace, instanceID)
}
