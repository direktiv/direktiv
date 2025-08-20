package engine

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	Put(ctx context.Context, namespace, instanceID, typ string, payload any) (uuid.UUID, error)
	QueryByInstance(ctx context.Context, namespace, instanceID uuid.UUID, typ string) ([]Message, error)
}

type Message struct {
	Namespace string    `json:"namespace"`
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`

	Data json.RawMessage `json:"data"`
}
