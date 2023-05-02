package logstore

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// LogStore manages storing and querying LogEntries.
type LogStore interface {
	// returns a query object to do further queries on the Server Logs
	ForServer() LogQuery
	// returns a query object to do further queries on the Instance Logs
	ForInstance(recipientID uuid.UUID) LogQuery
	// returns a query object to do further queries on the Workspace Logs
	ForWorkspace(recipientID uuid.UUID) LogQuery
	// returns a query object to do further queries on the Namespace Logs
	ForNamespace(recipientID uuid.UUID) LogQuery
}

// Represents an individual log entry for activity.
type LogEntry struct {
	// the timestamp of the log-entry.
	T     time.Time
	Msg   string
	Level Level
	// Tags contains metadata of the log-entry.
	Tags map[string]string
}

// LogQuery generates a query statement to receive Log-Entries.
type LogQuery interface {
	// appends a log-entry to the logs of the corresponding RecipientID, RecipientType pair.
	Append(ctx context.Context, l LogEntry) error
	// returns a list of log-entries using the provided LogQuery.
	Get(ctx context.Context, ls LogSelector) ([]*LogEntry, error)
	// use to build a logSelector for receiving log-entries
	NewFilterBuilder() LogSelectorBuilder
}

// LogSelectorBuilder constructs a LogSelector.
type LogSelectorBuilder interface {
	// select logs with matching level.
	WithLogLevel(level Level)
	// select first x entries off Log-entries matching all other attributes.
	WithLimit(x int)
	// select matching results starting given offset.
	WithOffset(offset int)
	// select logs with matching tag value pair.
	WithTag(tag, value string)
	// get the final LogSelector.
	Get() (LogSelector, error)
}

// LogSelector generates a query statement to receive Log-Entries.
// A LogSelector is created using a LogSelectorBuilder.
type LogSelector interface {
	buildStatement() (string, error)
}

// represents valid log levels.
type Level string

const (
	Debug Level = "debug"
	Info  Level = "info"
	Error Level = "error"
	Panic Level = "panic"
)
