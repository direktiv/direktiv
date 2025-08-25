package core

import (
	"context"

	"github.com/direktiv/direktiv/pkg/engine"
	"github.com/google/uuid"
)

type Engine interface {
	Start(circuit *Circuit) error
	ExecWorkflow(ctx context.Context, namespace string, path string, input string) (uuid.UUID, error)
	GetInstance(ctx context.Context, namespace, instanceID uuid.UUID) ([]engine.Message, error)
}
