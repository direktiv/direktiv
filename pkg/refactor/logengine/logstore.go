package logengine

import (
	"context"
	"time"
)

// LogStore manages storing and querying LogEntries.
type LogStore interface {
	// appends a log entry to the logs. Passed keysAnValues will be associated with the log entry.
	// - the primaryKey MUST not be empty
	// - Instance logs SHOULD contain a log_instance_call_path is present in the keysAnValues it will be stored a secondary key
	Append(ctx context.Context, timestamp time.Time, level LogLevel, msg string, primaryKey string, keysAndValues map[string]interface{}) error
	// returns a limited number of log-entries that have matching associated fields with the provided keysAndValues pairs
	// starting a given offset. For no offset or unlimited log-entries in the result set the value to 0.
	// passing a level in keysAndValues returns log-entries will have same or higher level as the passed one.
	// passing a log_instance_call_path in keysAndValues will return all logs which have a callpath with the prefix as the passed log_instance_call_path value.
	Get(ctx context.Context, limit, offset int, primaryKey string, keysAndValues map[string]interface{}) ([]*LogEntry, error)
}

// Represents an individual log entry.
type LogEntry struct {
	// the timestamp of the log-entry.
	T   time.Time
	Msg string
	// Fields contains metadata of the log-entry.
	Fields map[string]interface{}
}

type LogLevel int

const (
	Debug LogLevel = iota
	Info           = iota
	Error          = iota
)
