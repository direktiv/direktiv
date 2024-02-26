package core

import (
	"context"
	"net/http"
	"time"
)

type LogCollectionManager interface {
	GetNewer(ctx context.Context, t time.Time, params map[string]string) ([]PlattformLogEntry, error)
	GetOlder(ctx context.Context, params map[string]string) ([]PlattformLogEntry, error)
	Stream(params map[string]string) http.HandlerFunc
}

type PlattformLogEntry struct {
	ID        string    `json:"id"`
	Time      time.Time `json:"time"`
	Msg       string    `json:"msg"`
	Level     string    `json:"level"`
	Trace     string    `json:"trace,omitempty"`
	Span      string    `json:"span,omitempty"`
	State     string    `json:"state,omitempty"`
	Branch    string    `json:"branch,omitempty"`
	Workflow  string    `json:"workflow,omitempty"`
	Instance  string    `json:"instance,omitempty"`
	Namespace string    `json:"namespace,omitempty"`
	Activity  string    `json:"activity,omitempty"`
	Route     string    `json:"route,omitempty"`
	Path      string    `json:"path,omitempty"`
	Error     string    `json:"error,omitempty"`
}
