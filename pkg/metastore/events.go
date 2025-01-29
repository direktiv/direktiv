package metastore

import (
	"context"
	"time"
)

type EventsStore interface {
	Append(ctx context.Context, e EventEntry) error
	Get(ctx context.Context, options EventQueryOptions) ([]EventEntry, error)
	Init(ctx context.Context) error
	GetMapping(ctx context.Context) (map[string]interface{}, error)
}

type EventEntry struct {
	ID         string
	ReceivedAt int64
	CloudEvent string // original cloud-event as json-string
	Namespace  string
	Metadata   map[string]string
}

// EventQueryOptions defines filters for querying logs.
type EventQueryOptions struct {
	StartTime time.Time         // Start of the time range for events.
	EndTime   time.Time         // End of the time range for events.
	Metadata  map[string]string // Metadata filters (e.g., service name, environment).
	Keywords  []string          // Keywords to search within cloud-event.
	Limit     int               // Maximum number of events to retrieve.
}
