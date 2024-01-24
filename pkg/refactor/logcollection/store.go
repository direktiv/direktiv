package logcollection

import (
	"context"
	"time"
)

type LogEntry struct {
	Time time.Time
	Tag  string
	Data map[string]interface{}
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
