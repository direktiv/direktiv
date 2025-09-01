package engine

import (
	"context"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/google/uuid"
)

type Store interface {
	PushInstanceMessage(ctx context.Context, namespace string, instanceID uuid.UUID, typ string, payload any) (uuid.UUID, error)
	PullInstanceMessages(ctx context.Context, namespace string, instanceID uuid.UUID, typ string) ([]core.EngineMessage, error)
}
