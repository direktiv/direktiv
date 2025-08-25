package core

import (
	"context"

	"github.com/google/uuid"
)

type Engine interface {
	Start(circuit *Circuit) error
	ExecWorkflow(ctx context.Context, namespace string, path string, input string) (uuid.UUID, error)
	GetInstanceMessages(ctx context.Context, namespace string, instanceID uuid.UUID) (any, error)
}
