package logstore

import (
	"context"
	"time"
)

// LogStore manages storing and querying LogEntries.
type LogStore interface {
	// appends a log-entry to the logs.
	Append(ctx context.Context, timestamp time.Time, msg string, keysAndValues ...interface{}) error
	// returns a list of log-entries which have matching fields with the provides keysAndValues pairs.
	Get(ctx context.Context, keysAndValues ...interface{}) ([]*LogEntry, error)
}

// Represents an individual log entry for activity.
type LogEntry struct {
	// the timestamp of the log-entry.
	T   time.Time
	Msg string
	// Fields contains metadata of the log-entry.
	Fields map[string]interface{}
}
