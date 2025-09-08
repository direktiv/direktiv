package engine

import (
	"context"
	"encoding/json"
	"time"

	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/google/uuid"
)

type InstanceStatus struct {
	InstanceID uuid.UUID         `json:"instanceId"`
	Namespace  string            `json:"namespace"`
	Metadata   map[string]string `json:"metadata"`
	Script     string            `json:"script,omitempty"`
	Input      json.RawMessage   `json:"input,omitempty"`
	Memory     json.RawMessage   `json:"memory,omitempty"`
	Output     json.RawMessage   `json:"output,omitempty"`
	Error      string            `json:"error,omitempty"`
	Status     string            `json:"status"`
	CreatedAt  time.Time         `json:"createdAt"`
	EndedAt    time.Time         `json:"endedAt"`
	// history stream sequence this status came from
	HistorySequence uint64 `json:"historySequence"`
	Sequence        uint64 `json:"sequence"`
}

func (i *InstanceStatus) StatusString() string {
	switch i.Status {
	case "started":
		return "pending"
	case "failed":
		return "failed"
	case "succeeded":
		return "complete"
	}

	return i.Status
}

type InstanceEvent struct {
	EventID    uuid.UUID         `json:"eventId"`
	InstanceID uuid.UUID         `json:"instanceId"`
	Namespace  string            `json:"namespace"`
	Metadata   map[string]string `json:"metadata"`
	Type       string            `json:"type"`
	Time       time.Time         `json:"time"`

	Script string          `json:"script,omitempty"`
	Input  json.RawMessage `json:"input,omitempty"`
	Memory json.RawMessage `json:"memory,omitempty"`
	Output json.RawMessage `json:"output,omitempty"`
	Error  string          `json:"error,omitempty"`

	// history stream sequence
	Sequence uint64 `json:"sequence"`
}

type Projector interface {
	Start(lc *lifecycle.Manager) error
}

type WorkflowRunner interface {
	Execute(ctx context.Context, namespace string, scrip string, fn string, args any, labels map[string]string) (uuid.UUID, error)
}

type DataBus interface {
	Start(lc *lifecycle.Manager) error
	PushInstanceEvent(ctx context.Context, event *InstanceEvent) error
	QueryInstanceStatus(ctx context.Context, filterNamespace string, filterInstanceID uuid.UUID) []InstanceStatus
}
