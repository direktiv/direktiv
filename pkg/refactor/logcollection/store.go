package logcollection

import (
	"context"
	"time"
)

// Represents an individual log entry.
type LogEntry struct {
	// the timestamp of the log-entry.
	Time time.Time
	Tag  string
	// Fields contains metadata of the log-entry.
	data map[string]interface{}
}

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
)

// LogStore manages storing and querying LogEntries.
type LogStore interface {
	Get(ctx context.Context, stream string, offset int) ([]LogEntry, error)
	GetInstanceLogs(ctx context.Context, stream string, instanceID string, offset int) ([]LogEntry, error)
	DeleteOldLogs(ctx context.Context, t time.Time) error
}
