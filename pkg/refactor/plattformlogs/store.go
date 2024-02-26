package plattformlogs

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
	GetNewer(ctx context.Context, stream string, t time.Time) ([]LogEntry, error)
	GetOlder(ctx context.Context, stream string, t time.Time) ([]LogEntry, error)
	GetStartingIDUntilTime(ctx context.Context, stream string, lastID int, t time.Time) ([]LogEntry, error)
	GetNewerInstance(ctx context.Context, stream string, t time.Time) ([]LogEntry, error)
	GetOlderInstance(ctx context.Context, stream string, t time.Time) ([]LogEntry, error)
	GetStartingIDUntilTimeInstance(ctx context.Context, stream string, lastID int, t time.Time) ([]LogEntry, error)
	DeleteOldLogs(ctx context.Context, t time.Time) error
}

type LogEntry struct {
	ID   int
	Time time.Time
	Tag  string
	Data map[string]interface{}
}

func ToFeatureLogEntry(e LogEntry) (core.PlattformLogEntry, error) {
	entry, ok := e.Data["log"].(string)
	if !ok {
		return core.PlattformLogEntry{}, fmt.Errorf("log-entry format is corrupt")
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(entry), &m); err != nil {
		return core.PlattformLogEntry{}, fmt.Errorf("failed to unmarshal log entry: %w", err)
	}

	featureLogEntry := core.PlattformLogEntry{
		ID:    strconv.Itoa(e.ID),
		Time:  e.Time,
		Msg:   fmt.Sprint(m["msg"]),
		Level: fmt.Sprint(m["level"]),
	}
	if s, ok := m["trace"]; ok {
		featureLogEntry.Trace = fmt.Sprint(s)
	}
	if s, ok := m["span"]; ok {
		featureLogEntry.Span = fmt.Sprint(s)
	}
	if s, ok := m["state"]; ok {
		featureLogEntry.State = fmt.Sprint(s)
	}
	if s, ok := m["namespace"]; ok {
		featureLogEntry.Namespace = fmt.Sprint(s)
	}
	if s, ok := m["workflow"]; ok {
		featureLogEntry.Workflow = fmt.Sprint(s)
	}
	if s, ok := m["instance"]; ok {
		featureLogEntry.Instance = fmt.Sprint(s)
	}
	if s, ok := m["route"]; ok {
		featureLogEntry.Route = fmt.Sprint(s)
	}
	if s, ok := m["activity"]; ok {
		featureLogEntry.Activity = fmt.Sprint(s)
	}
	if s, ok := m["branch"]; ok {
		featureLogEntry.Branch = fmt.Sprint(s)
	}
	if s, ok := m["error"]; ok {
		featureLogEntry.Error = fmt.Sprint(s)
	}
	if s, ok := m["path"]; ok {
		featureLogEntry.Path = fmt.Sprint(s)
	}

	return featureLogEntry, nil
}
