package metastore

import (
	"context"
	"time"
)

// LogStore defines a high-level interface for managing logs in the Metastore.
type LogStore interface {
	Init(ctx context.Context) error
	Get(ctx context.Context, options LogQueryOptions) ([]LogEntry, error)
}

// LogEntry represents a log entry aligned with the OpenSearch index mapping.
type LogEntry struct {
	Time      time.Time `json:"time"`      // The timestamp of the log entry
	Callpath  string    `json:"callpath"`  // The call path of the instance
	Instance  string    `json:"instance"`  // Instance identifier
	Invoker   string    `json:"invoker"`   // Who invoked the instance/workflow
	Level     string    `json:"level"`     // Log severity level
	Msg       string    `json:"msg"`       // Log message
	Namespace string    `json:"namespace"` // Namespace of the log entry
	Span      string    `json:"span"`      // Trace span ID
	State     string    `json:"state"`     // Workflow state
	Status    string    `json:"status"`    // Status of the workflow/process
	Stream    string    `json:"stream"`    // Log stream
	Trace     string    `json:"trace"`     // Trace ID
	Track     string    `json:"track"`     // Tracking ID
	Workflow  string    `json:"workflow"`  // Workflow name
}

// LogQueryOptions defines filters for querying logs.
type LogQueryOptions struct {
	StartTime time.Time         // Start of time range filter
	EndTime   time.Time         // End of time range filter
	Level     string            // Log level (e.g., "INFO", "ERROR")
	Keywords  []string          // Keywords to search within the message field
	Metadata  map[string]string // Additional metadata filters (e.g., namespace, instance, workflow)
	Limit     int               // Maximum number of logs to return
}
