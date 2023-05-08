package logengine

import (
	"context"
	"time"
)

// LogStore manages storing and querying LogEntries.
type LogStore interface {
	// appends a log entry to the logs. Passed keyAnValues will be associated with the log entry.
	// Example: store.Append(ctx, time.Time, "message", "key", "value", "other-key", "value")
	Append(ctx context.Context, timestamp time.Time, msg string, keysAndValues ...interface{}) error
	// returns a list of log-entries that have matching associated fields with the provided keysAndValues pairs.
	// Example: store.Get(ctx, "key1", value1, "key2", value2)
	Get(ctx context.Context, keysAndValues ...interface{}) ([]*LogEntry, error)
}

// Represents an individual log entry.
type LogEntry struct {
	// the timestamp of the log-entry.
	T   time.Time
	Msg string
	// Fields contains metadata of the log-entry.
	Fields map[string]interface{}
}
