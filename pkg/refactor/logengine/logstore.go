package logengine

import (
	"context"
	"time"
)

// LogStore manages storing and querying LogEntries.
type LogStore interface {
	// ! Do not use Append directly use the BetterLogger Interface instead.
	// Appends a log entry to the logs. Passed keysAnValues will be associated with the log entry.
	// - keysAndValues will be attached to the log-entry.
	// - keysAndValues SHOULD contain contextual information for the log message.
	// - source:uuid, type:string SHOULD be present in keyValues. This makes it possible to process the logEntry as an log-event.
	// - for keysAndValues["type"]="instance" the following also SHOULD be present: log_instance_call_path:string, root_instance_id:uuid.
	Append(ctx context.Context, timestamp time.Time, level LogLevel, msg string, keysAndValues map[string]interface{}) error
	// Returns a limited number of log-entries (and the total count) that have matching associated fields with the provided keysAndValues pairs
	// starting a given offset. For no offset or unlimited log-entries in the result set the value to 0.
	// - level SHOULD be passed as a string. Possible values are debug, info, error.
	// - This method will search for any of followings keys and query all matching logs:
	// level, log_instance_call_path, root_instance_id, type, source
	// Any other not mentioned passed key value pair will be ignored.
	// Returned log-entries will have same or higher level as the passed one.
	// - Passing a log_instance_call_path will return all logs which have a callpath with the prefix as the passed log_instance_call_path value.
	// when passing log_instance_call_path the root_instance_id SHOULD be passed to optimize the performance of the query.
	Get(ctx context.Context, keysAndValues map[string]interface{}, limit, offset int) ([]*LogEntry, int, error)
	DeleteOldLogs(ctx context.Context, t time.Time) error
}

// Represents an individual log entry.
type LogEntry struct {
	ID int
	// the timestamp of the log-entry.
	T   time.Time
	Msg string
	// Fields contains metadata of the log-entry.
	Fields map[string]interface{}
}

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
)
