package core

import (
	"context"
	"net/http"
	"time"
)

type LogCollectionManager interface {
	Get(ctx context.Context, cursorTime time.Time, params map[string]string) ([]FeatureLogEntry, error)
	Stream(params map[string]string) http.HandlerFunc
}

type FeatureLogEntry struct {
	Time     time.Time         `json:"time"`
	Msg      string            `json:"msg"`
	Level    string            `json:"level"`
	Trace    string            `json:"trace"`
	State    string            `json:"state"`
	Branch   string            `json:"branch"`
	Metadata map[string]string `json:"metadata"`
}
