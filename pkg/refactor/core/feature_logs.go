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
	ID     string    `json:"id"`
	Time   time.Time `json:"time"`
	Msg    string    `json:"msg"`
	Level  string    `json:"level"`
	Trace  string    `json:"trace,omitempty"`
	State  string    `json:"state,omitempty"`
	Branch string    `json:"branch,omitempty"`
}
