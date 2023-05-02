package logstore

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// LogStore manages storing and querying LogEntries.
type LogStore interface {
	// appends a log-entry to the logs of the corresponding RecipientID, RecipientType pair.
	Append(recipientID uuid.UUID, recipientType RecipientType, l LogEntry) error
	// returns a list of log-entries using the provided LogQuery.
	// Example usage:
	// ql := GetLogsQuery(recipienttId, recipientType, loglevel)
	// logs, err := logstore.Get(ctx, ql)
	Get(ctx context.Context, ql LogQuery) ([]*LogEntry, error)
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
	BuildStatement() (string, error)
}

// represents valid recipient type for log-messages.
type RecipientType string

const (
	Server    RecipientType = "server"
	Namespace RecipientType = "namespace"
	Workflow  RecipientType = "workflow"
	Instance  RecipientType = "instance"
	Mirror    RecipientType = "mirror"
)

func (rt RecipientType) String() string {
	return string(rt)
}

// represents valid log levels.
type Level string

const (
	Debug Level = "debug"
	Info  Level = "info"
	Error Level = "error"
	Panic Level = "panic"
)
