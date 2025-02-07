package datasql

import (
	"context"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/datastore"
	"gorm.io/gorm"
)

var _ datastore.TracesStore = &sqlTracesStore{}

type sqlTracesStore struct {
	db *gorm.DB
}

// Append implements datastore.TracesStore.
func (s *sqlTracesStore) Append(ctx context.Context, trace datastore.Trace) error {
	// Ensure StartTime is stored in UTC
	trace.Starttime = trace.Starttime.UTC()
	if trace.Endtime != nil {
		t := trace.Endtime.UTC()
		trace.Endtime = &t
	}
	// The SQL query to insert a new trace
	q := `INSERT INTO traces (trace_id, span_id, parent_span_id, starttime, endtime, raw_trace)
		  VALUES ($1, $2, $3, $4, $5, $6);`

	// Executing the query using the provided trace data
	tx := s.db.WithContext(ctx).Exec(
		q,
		trace.TraceID,
		trace.SpanID,
		trace.ParentSpanID,
		trace.Starttime,
		trace.Endtime,
		trace.RawTrace,
	)

	return tx.Error
}

// DeleteOld implements datastore.TracesStore.
func (s *sqlTracesStore) DeleteOld(ctx context.Context, cutoffTime time.Time) error {
	// SQL query to delete traces older than the cutoff time
	q := `DELETE FROM traces WHERE starttime < $1;`

	tx := s.db.WithContext(ctx).Exec(q, cutoffTime)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

// GetByParentSpanID implements datastore.TracesStore.
func (s *sqlTracesStore) GetByParentSpanID(ctx context.Context, parentSpanID string) ([]datastore.Trace, error) {
	// SQL query to select traces by parent span ID
	q := `SELECT trace_id, span_id, parent_span_id, starttime, endtime, raw_trace
		  FROM traces WHERE parent_span_id = $1;`

	var res []datastore.Trace
	tx := s.db.WithContext(ctx).Raw(q, parentSpanID).Scan(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return res, nil
}

// GetByTraceID implements datastore.TracesStore.
func (s *sqlTracesStore) GetByTraceID(ctx context.Context, traceID string) (datastore.Trace, error) {
	q := `SELECT trace_id, span_id, parent_span_id, starttime, endtime, raw_trace
		  FROM traces WHERE trace_id = $1;`

	var trace datastore.Trace
	tx := s.db.WithContext(ctx).Raw(q, traceID).First(&trace)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return datastore.Trace{}, fmt.Errorf("trace not found: %w", tx.Error)
		}

		return datastore.Trace{}, tx.Error
	}

	// Ensure StartTime is in UTC
	trace.Starttime = trace.Starttime.UTC()
	if trace.Endtime != nil {
		t := trace.Endtime.UTC()
		trace.Endtime = &t
	}

	return trace, nil
}
