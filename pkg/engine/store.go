package engine

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	PushInstanceMessage(ctx context.Context, namespace string, instanceID uuid.UUID, typ string, payload any) (uuid.UUID, error)
	PullInstanceMessages(ctx context.Context, namespace string, instanceID uuid.UUID, typ string) ([]Message, error)
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
	Status       int       `json:"status"`

	WorkflowText string `json:"workflowText,omitempty"`

	Input  json.RawMessage `json:"input,omitempty"`
	Memory json.RawMessage `json:"memory,omitempty"`
	Output json.RawMessage `json:"output,omitempty"`

	Error   error     `json:"error,omitempty"`
	EndedAt time.Time `json:"endedAt,omitempty"`
}
