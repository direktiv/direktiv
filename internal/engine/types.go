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
	Mappings   string            `json:"mappings,omitempty"`
	Fn         string            `json:"fn,omitempty"`
	Input      json.RawMessage   `json:"input,omitempty"`
	Memory     json.RawMessage   `json:"memory,omitempty"`
	Output     json.RawMessage   `json:"output,omitempty"`
	Error      string            `json:"error,omitempty"`
	Status     string            `json:"status"`
	CreatedAt  time.Time         `json:"createdAt"`
	StartedAt  time.Time         `json:"StartedAt"`
	EndedAt    time.Time         `json:"endedAt"`
	// history stream sequence this status came from
	HistorySequence uint64 `json:"historySequence"`
	Sequence        uint64 `json:"sequence"`
}

func (i *InstanceStatus) StatusString() string {
	switch i.Status {
	case "running":
		return "pending"
	case "failed":
		return "failed"
	case "succeeded":
		return "complete"
	}

	return i.Status
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
	if i.Memory != nil {
		clone.Memory = make(json.RawMessage, len(i.Memory))
		copy(clone.Memory, i.Memory)
	}
	if i.Output != nil {
		clone.Output = make(json.RawMessage, len(i.Output))
		copy(clone.Output, i.Output)
	}

	return &clone
}

func (i *InstanceStatus) IsEndStatus() bool {
	return i.Status == "succeeded" || i.Status == "failed"
}

type InstanceEvent struct {
	EventID    uuid.UUID         `json:"eventId"`
	InstanceID uuid.UUID         `json:"instanceId"`
	Namespace  string            `json:"namespace"`
	Metadata   map[string]string `json:"metadata"`
	Type       string            `json:"type"`
	Time       time.Time         `json:"time"`

	Script   string          `json:"script,omitempty"`
	Mappings string          `json:"mappings,omitempty"`
	Fn       string          `json:"fn,omitempty"`
	Input    json.RawMessage `json:"input,omitempty"`
	Memory   json.RawMessage `json:"memory,omitempty"`
	Output   json.RawMessage `json:"output,omitempty"`
	Error    string          `json:"error,omitempty"`

	// history stream sequence
	Sequence uint64 `json:"sequence"`
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

	clone.Input = copyRaw(e.Input)
	clone.Memory = copyRaw(e.Memory)
	clone.Output = copyRaw(e.Output)

	return &clone
}

func ApplyInstanceEvent(st *InstanceStatus, ev *InstanceEvent) {
	st.Status = ev.Type
	st.HistorySequence = ev.Sequence //

	switch ev.Type {
	case "pending":
		st.InstanceID = ev.InstanceID
		st.Namespace = ev.Namespace
		st.Metadata = ev.Metadata
		st.Script = ev.Script
		st.Mappings = ev.Mappings
		st.Fn = ev.Fn
		st.Input = ev.Input
		st.CreatedAt = ev.Time
	case "running":
		st.StartedAt = ev.Time
		st.Memory = ev.Memory
		st.Fn = ev.Fn
	case "failed":
		st.EndedAt = ev.Time
		st.Memory = ev.Memory
		st.Output = ev.Output
		st.Error = ev.Error
	case "succeeded":
		st.EndedAt = ev.Time
		st.Memory = ev.Memory
		st.Output = ev.Output
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
