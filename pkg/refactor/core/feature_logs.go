package core

import (
	"context"
	"net/http"
	"time"
)

type LogCollectionManager interface {
	GetAfter(ctx context.Context, after time.Time, params map[string]string) ([]FeatureLogEntry, error)
	GetFirst(ctx context.Context, params map[string]string) ([]FeatureLogEntry, error)
	Stream(params map[string]string) http.HandlerFunc
}

type FeatureLogEntry struct {
	ID       string            `json:"id"`
	Time     time.Time         `json:"time"`
	Msg      string            `json:"msg"`
	Level    string            `json:"level"`
	Trace    string            `json:"trace,omitempty"`
	State    string            `json:"state,omitempty"`
	Branch   string            `json:"branch,omitempty"`
	Metadata map[string]string `json:"metadata"`
}
