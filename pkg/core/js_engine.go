package core

import (
	"context"

	"github.com/google/uuid"
)

type JSEngine interface {
	Run(circuit *Circuit) error
	ExecWorkflow(ctx context.Context, namespace string, path string, input string) (uuid.UUID, error)
}
