package metastore

import (
	"context"
	"time"
)

type TimelineStore interface {
	Get(ctx context.Context, traceID string, options TimelineQueryOptions) ([]map[string]any, error)
}

type TimelineQueryOptions struct {
	StartTime time.Time // Start of time range filter
	EndTime   time.Time // End of time range filter
	Severity  string    // Log level (e.g., "INFO", "ERROR")
	SpanID    string
	Metadata  map[string]string // Additional metadata filters (e.g., namespace, instance, workflow)
	Limit     int               // Maximum number of logs to return
}
