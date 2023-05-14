package logengine

import (
	"context"
	"time"
)

// LogStore manages storing and querying LogEntries.
type LogStore interface {
	// appends a log entry to the logs. Passed keysAnValues will be associated with the log entry.
	Append(ctx context.Context, level string, msg string, keysAndValues map[string]interface{}) error
	// returns a limited number of log-entries that have matching associated fields with the provided keysAndValues pairs
	// starting a given offset. For no offset or unlimited log-entries in the result set the value to -1.
	Get(ctx context.Context, keysAndValues map[string]interface{}, limit, offset int) ([]*LogEntry, error)
}

// Represents an individual log entry.
type LogEntry struct {
	// the timestamp of the log-entry.
	T   time.Time
	Msg string
	// Fields contains metadata of the log-entry.
	Fields map[string]interface{}
}
