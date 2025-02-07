package datastore

import (
	"context"
	"time"
)

// Trace represents the structure of a trace in the database.
type Trace struct {
	TraceID      string     `json:"traceID"`
	SpanID       string     `json:"spanID"`
	ParentSpanID *string    `json:"parentSpanID,omitempty"`
	Starttime    time.Time  `json:"starttime"`
	Endtime      *time.Time `json:"endtime,omitempty"`
	RawTrace     []byte     `json:"rawTrace"`
}

// TracesStore defines the interface for interacting with trace data.
type TracesStore interface {
	// Append adds a new trace entry to the store.
	Append(ctx context.Context, trace Trace) error

	// DeleteOld deletes traces older than the specified cutoff time.
	DeleteOld(ctx context.Context, cutoffTime time.Time) error

	// GetByParentSpanID retrieves traces by their parent span ID.
	GetByParentSpanID(ctx context.Context, parentSpanID string) ([]Trace, error)

	// GetByTraceID retrieves a trace by its trace ID.
	GetByTraceID(ctx context.Context, traceID string) (Trace, error)
}
