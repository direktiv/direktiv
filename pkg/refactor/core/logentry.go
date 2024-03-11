package core

import (
	"context"
	"net/http"
	"time"
)

type Status string

const (
	ErrStatus       Status = "error"
	UnknownStatus   Status = "unknown"
	RunningStatus   Status = "running"
	FailedStatus    Status = "failed"
	CompletedStatus Status = "completed"
)

type ContextKey string

const (
	TrackKey ContextKey = "track"
	TagsKey  ContextKey = "tags"
)

type LogCollectionManager interface {
	GetNewer(ctx context.Context, t time.Time, params map[string]string) ([]PlattformLogEntry, error)
	GetOlder(ctx context.Context, params map[string]string) ([]PlattformLogEntry, time.Time, error)
	Stream(params map[string]string) http.HandlerFunc
}

type PlattformLogEntry struct {
	ID        int                   `json:"id"`
	Time      time.Time             `json:"time"`
	Msg       interface{}           `json:"msg"`
	Level     interface{}           `json:"level"`
	Namespace interface{}           `json:"namespace"`
	Trace     interface{}           `json:"trace"`
	Span      interface{}           `json:"span"`
	Workflow  *WorkflowEntryContext `json:"workflow,omitempty"`
	Activity  *ActivityEntryContext `json:"activity,omitempty"`
	Route     *RouteEntryContext    `json:"route,omitempty"`
	Error     interface{}           `json:"error"`
}

type WorkflowEntryContext struct {
	Status interface{} `json:"status"`

	State    interface{} `json:"state"`
	Branch   interface{} `json:"branch"`
	Path     interface{} `json:"workflow"`
	CalledAs interface{} `json:"calledAs"`
	Instance interface{} `json:"instance"`
}

type ActivityEntryContext struct {
	ID interface{} `json:"id,omitempty"`
}
type RouteEntryContext struct {
	Path interface{} `json:"path,omitempty"`
}
