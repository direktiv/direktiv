package core

import (
	"context"
	"net/http"
	"time"
)

type Status string

const (
	ErrStatus       Status = "error"
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
	GetOlder(ctx context.Context, params map[string]string) ([]PlattformLogEntry, error)
	Stream(params map[string]string) http.HandlerFunc
}

type PlattformLogEntry struct {
	ID        string      `json:"id"`
	Time      time.Time   `json:"time"`
	Msg       interface{} `json:"msg"`
	Level     interface{} `json:"level"`
	Trace     interface{} `json:"trace"`
	Span      interface{} `json:"span"`
	State     interface{} `json:"state"`
	Branch    interface{} `json:"branch"`
	Workflow  interface{} `json:"workflow"`
	Instance  interface{} `json:"instance"`
	Namespace interface{} `json:"namespace"`
	Activity  interface{} `json:"activity"`
	Route     interface{} `json:"route"`
	Path      interface{} `json:"path"`
	Error     interface{} `json:"error"`
}
