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
	Trace     string    `json:"trace"`
	Span      string    `json:"span"`
	State     string    `json:"state"`
	Branch    string    `json:"branch"`
	Workflow  string    `json:"workflow"`
	Instance  string    `json:"instance"`
	Namespace string    `json:"namespace"`
	Activity  string    `json:"activity"`
	Route     string    `json:"route"`
	Path      string    `json:"path"`
	Error     string    `json:"error"`
}
