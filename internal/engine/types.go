package engine

import (
	"context"
	"encoding/json"
	"time"

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
	StateCodeCrashed   StateCode = "crashed"
)

var AllStateCodes = []StateCode{
	StateCodePending,
	StateCodeFailed,
	StateCodeComplete,
	StateCodeCancelled,
	StateCodeCrashed,
}

type InstanceStatus struct {
	InstanceID uuid.UUID
	Namespace  string
	Metadata   map[string]string
	Script     string
	Mappings   string
	Fn         string
	Input      json.RawMessage `json:",omitempty"`
	Output     json.RawMessage `json:",omitempty"`
	Error      string
	State      StateCode
	CreatedAt  time.Time
	StartedAt  time.Time
	EndedAt    time.Time
	// history stream sequence this status came from
	HistorySequence uint64
	Sequence        uint64
}

func (i *InstanceStatus) StatusString() string {
	switch i.State {
	case StateCodeRunning:
		return string(StateCodePending)
	}

	return string(i.State)
}

func (i *InstanceStatus) Clone() *InstanceStatus {
	// start with a shallow copy
	clone := *i

	// deep copy the Metadata map
	if i.Metadata != nil {
		clone.Metadata = make(map[string]string, len(i.Metadata))
		for k, v := range i.Metadata {
			clone.Metadata[k] = v
		}
	}
	// deep copy the buffers
	if i.Input != nil {
		clone.Input = make(json.RawMessage, len(i.Input))
		copy(clone.Input, i.Input)
	}
	if i.Output != nil {
		clone.Output = make(json.RawMessage, len(i.Output))
		copy(clone.Output, i.Output)
	}

	return &clone
}

func (i *InstanceStatus) IsEndStatus() bool {
	return i.State == StateCodeComplete || i.State == StateCodeFailed || i.State == StateCodeCancelled
}

type InstanceEvent struct {
	EventID    uuid.UUID
	InstanceID uuid.UUID
	Namespace  string
	Metadata   map[string]string
	Type       StateCode
	Time       time.Time

	Script   string
	Mappings string
	Fn       string
	Memory   json.RawMessage `json:",omitempty"`
	Error    string

	// history stream sequence
	Sequence uint64
}

func (e *InstanceEvent) Clone() *InstanceEvent {
	// start with a shallow copy
	clone := *e

	// deep copy Metadata
	if e.Metadata != nil {
		clone.Metadata = make(map[string]string, len(e.Metadata))
		for k, v := range e.Metadata {
			clone.Metadata[k] = v
		}
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

	clone.Memory = copyRaw(e.Memory)

	return &clone
}

func ApplyInstanceEvent(st *InstanceStatus, ev *InstanceEvent) {
	st.State = ev.Type
	st.HistorySequence = ev.Sequence //

	switch ev.Type {
	case StateCodePending:
		st.InstanceID = ev.InstanceID
		st.Namespace = ev.Namespace
		st.Metadata = ev.Metadata
		st.Script = ev.Script
		st.Mappings = ev.Mappings
		st.Fn = ev.Fn
		st.Input = ev.Memory
		st.CreatedAt = ev.Time
	case StateCodeRunning:
		st.StartedAt = ev.Time
		st.Fn = ev.Fn
	case StateCodeFailed:
		st.EndedAt = ev.Time
		st.Error = ev.Error
	case StateCodeComplete:
		st.EndedAt = ev.Time
		st.Output = ev.Memory
		st.Error = ev.Error
	}
}

type WorkflowRunner interface {
	Execute(ctx context.Context, namespace string, scrip string, fn string, args any, labels map[string]string) (uuid.UUID, error)
}

type DataBus interface {
	Start(lc *lifecycle.Manager) error

	PublishInstanceHistoryEvent(ctx context.Context, event *InstanceEvent) error
	PublishInstanceQueueEvent(ctx context.Context, event *InstanceEvent) error

	ListInstanceStatuses(ctx context.Context, filterNamespace string, filterInstanceID uuid.UUID, limit int, offset int) ([]*InstanceStatus, int)
	GetInstanceHistory(ctx context.Context, namespace string, instanceID uuid.UUID) []*InstanceEvent

	NotifyInstanceStatus(ctx context.Context, instanceID uuid.UUID, done chan<- *InstanceStatus)

	DeleteNamespace(ctx context.Context, namespace string) error
}
