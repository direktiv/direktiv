package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EngineMessage struct {
	Namespace string    `json:"namespace"`
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdAt"`

	Data json.RawMessage `json:"data"`
}

type InstanceMessage struct {
	InstanceID uuid.UUID `json:"instanceId"`
	Namespace  string    `json:"namespace"`
	Script     string    `json:"script"`
	Status     int       `json:"status"`

	Labels map[string]string `json:"labels"`

	Input  json.RawMessage `json:"input,omitempty"`
	Memory json.RawMessage `json:"memory,omitempty"`
	Output json.RawMessage `json:"output,omitempty"`

	Error   string    `json:"error,omitempty"`
	EndedAt time.Time `json:"endedAt,omitempty"`
}

func (m InstanceMessage) StatusString() string {
	return "complete"
}

type Engine interface {
	Start(circuit *Circuit) error
	ExecWorkflow(ctx context.Context, namespace string, script string, fn string, args any, labels map[string]string) (uuid.UUID, error)
	GetInstanceMessages(ctx context.Context, namespace string, instanceID uuid.UUID) ([]EngineMessage, error)
	ListInstances(ctx context.Context, namespace string) ([]uuid.UUID, error)
}
