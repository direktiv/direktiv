package plattformlogs

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
		ID:    e.ID,
		Time:  e.Time,
		Msg:   m["msg"],
		Level: m["level"],
	}
	featureLogEntry.Error = m["error"]
	featureLogEntry.Trace = m["trace"]
	featureLogEntry.Span = m["span"]
	featureLogEntry.Namespace = m["namespace"]
	wfLogCtx := core.WorkflowEntryContext{}
	wfLogCtx.State = m["state"]
	wfLogCtx.Workflow = m["workflow"]
	wfLogCtx.Instance = m["instance"]
	wfLogCtx.CalledAs = m["calledAs"]
	wfLogCtx.Status = m["status"]
	wfLogCtx.Branch = m["branch"]
	featureLogEntry.Workflow = &wfLogCtx
	if wfLogCtx.Workflow == nil && wfLogCtx.Instance == nil {
		featureLogEntry.Workflow = nil
	}
	if id, ok := m["activity"]; ok && id != nil {
		featureLogEntry.Activity = &core.ActivityEntryContext{ID: id}
	}
	if path, ok := m["route"]; ok && path != nil {
		featureLogEntry.Route = &core.RouteEntryContext{Path: path}
	}
	// TODO Remove path log-key
	// if s, ok := m["path"]; ok {
	// 	featureLogEntry.Path = fmt.Sprint(s)
	// }

	return featureLogEntry, nil
}
