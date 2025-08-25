package engine

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	PutInstanceMessage(ctx context.Context, namespace string, instanceID uuid.UUID, typ string, payload any) (uuid.UUID, error)
	GetInstanceMessages(ctx context.Context, namespace string, instanceID uuid.UUID, typ string) ([]Message, error)
}

type Message struct {
	Namespace string    `json:"namespace"`
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdAt"`

	Data json.RawMessage `json:"data"`
}

type InstanceMessage struct {
	InstanceID   uuid.UUID `json:"instanceId"`
	Namespace    string    `json:"namespace"`
	WorkflowPath string    `json:"workflowPath"`
	WorkflowText string    `json:"workflowText"`

	Status int `json:"status"`

	Input  sql.NullString `json:"input"`
	Memory sql.NullString `json:"memory"`
	Output sql.NullString `json:"output"`
	Error  sql.NullString `json:"error"`

	EndedAt time.Time `json:"endedAt"`
}
