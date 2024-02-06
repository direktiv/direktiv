package plattform_logs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
)

// LogStore manages storing and querying LogEntries.
type LogStore interface {
	Get(ctx context.Context, stream string, cursorTime time.Time) ([]LogEntry, error)
	GetInstanceLogs(ctx context.Context, stream string, cursorTime time.Time) ([]LogEntry, error)
	DeleteOldLogs(ctx context.Context, t time.Time) error
}

type LogEntry struct {
	Time time.Time
	Tag  string
	Data map[string]interface{}
}

func (e LogEntry) ToFeatureLogEntry() (core.FeatureLogEntry, error) {
	entry, ok := e.Data["entry"].(string)
	if !ok {
		return core.FeatureLogEntry{}, fmt.Errorf("log-entry format is corrupt")
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(entry), &m); err != nil {
		return core.FeatureLogEntry{}, fmt.Errorf("failed to unmarshal log entry: %w", err)
	}

	return core.FeatureLogEntry{
		Time:     e.Time,
		Msg:      fmt.Sprint(m["msg"]),
		Level:    fmt.Sprint(m["level"]),
		Trace:    fmt.Sprint(m["trace"]),
		State:    fmt.Sprint(m["state"]),
		Branch:   fmt.Sprint(m["branch"]),
		Metadata: map[string]string{},
	}, nil
}
