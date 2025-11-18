package engine

import (
	"context"
	"encoding/json"
	"maps"
	"time"

	"github.com/direktiv/direktiv/internal/api/filter"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/google/uuid"
)

type StateCode string

const (
	StateCodePending   StateCode = "pending"
	StateCodeRunning   StateCode = "running"
	StateCodeComplete  StateCode = "complete"
	StateCodeFailed    StateCode = "failed"
	StateCodeCancelled StateCode = "cancelled"
)

var AllStateCodes = []StateCode{
	StateCodePending,
	StateCodeRunning,
	StateCodeComplete,
	StateCodeFailed,
	StateCodeCancelled,
}

type InstanceEvent struct {
	State      StateCode
	InstanceID uuid.UUID
	Namespace  string
	Metadata   map[string]string
	Script     string
	Fn         string
	Mappings   string

	Input  json.RawMessage `json:",omitempty"`
	Output json.RawMessage `json:",omitempty"`
	Error  string

	CreatedAt time.Time
	StartedAt time.Time
	EndedAt   time.Time

	EventID  uuid.UUID
	Sequence uint64
}

func (e *InstanceEvent) IsEndStatus() bool {
	return e.State == StateCodeComplete || e.State == StateCodeFailed || e.State == StateCodeCancelled
}

func (e *InstanceEvent) Clone() *InstanceEvent {
	// start with a shallow copy
	clone := *e

	// deep copy Metadata
	if e.Metadata != nil {
		clone.Metadata = make(map[string]string, len(e.Metadata))
		maps.Copy(clone.Metadata, e.Metadata)
	}

	// deep copy json.RawMessage fields
	copyRaw := func(src json.RawMessage) json.RawMessage {
		if src == nil {
			return nil
		}
		dst := make(json.RawMessage, len(src))
		copy(dst, src)

		return dst
	}

	clone.Input = copyRaw(e.Input)
	clone.Output = copyRaw(e.Output)

	return &clone
}

type WorkflowRunner interface {
	Execute(ctx context.Context, namespace string, scrip string, fn string, args any, labels map[string]string) (uuid.UUID, error)
}

type DataBus interface {
	Start(lc *lifecycle.Manager) error

	PublishInstanceHistoryEvent(ctx context.Context, event *InstanceEvent) error
	PublishInstanceQueueEvent(ctx context.Context, event *InstanceEvent) error

	ListInstanceStatuses(ctx context.Context, limit int, offset int, filters filter.Values) ([]*InstanceEvent, int)
	GetInstanceHistory(ctx context.Context, namespace string, instanceID uuid.UUID) []*InstanceEvent

	DeleteNamespace(ctx context.Context, namespace string) error

	PublishIgniteAction(ctx context.Context, svcID string) error
}
