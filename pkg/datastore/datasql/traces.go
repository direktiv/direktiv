package datasql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/datastore"
	"gorm.io/gorm"
)

var _ datastore.TracesStore = &sqlTracesStore{}

type sqlTracesStore struct {
	db *gorm.DB
}

// Append implements datastore.TracesStore.
func (s *sqlTracesStore) Append(ctx context.Context, traces ...datastore.Trace) error {
	if len(traces) == 0 {
		return nil // No traces to insert
	}

	// The SQL query to insert multiple traces
	q := `INSERT INTO traces (trace_id, span_id, parent_span_id, start_time, end_time, metadata) VALUES %s;`

	var values []interface{}
	var placeholders []string

	for _, trace := range traces {
		trace.StartTime = trace.StartTime.UTC()
		trace.EndTime = trace.EndTime.UTC()
		placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?)")
		values = append(values, trace.TraceID, trace.SpanID, trace.ParentSpanID, trace.StartTime, trace.EndTime, trace.Metadata)
	}

	finalQuery := fmt.Sprintf(q, strings.Join(placeholders, ", "))

	// Executing the batch insert query
	tx := s.db.WithContext(ctx).Exec(finalQuery, values...)

	return tx.Error
}

// DeleteOld implements datastore.TracesStore.
func (s *sqlTracesStore) DeleteOld(ctx context.Context, cutoffTime time.Time) error {
	// SQL query to delete traces older than the cutoff time
	q := `DELETE FROM traces WHERE start_time < $1;`

	tx := s.db.WithContext(ctx).Exec(q, cutoffTime)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

// GetByParentSpanID implements datastore.TracesStore.
func (s *sqlTracesStore) GetByParentSpanID(ctx context.Context, parentSpanID string) ([]datastore.Trace, error) {
	// SQL query to select traces by parent span ID
	q := `SELECT trace_id, span_id, parent_span_id, start_time, end_time, metadata
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
	q := `SELECT trace_id, span_id, parent_span_id, start_time, end_time, metadata
		  FROM traces WHERE trace_id = $1;`

	var trace datastore.Trace
	tx := s.db.WithContext(ctx).Raw(q, traceID).First(&trace)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return datastore.Trace{}, fmt.Errorf("trace not found: %w", tx.Error)
		}

		return datastore.Trace{}, tx.Error
	}

	// Ensure StartTime is in UTC
	trace.StartTime = trace.StartTime.UTC()
	trace.EndTime = trace.EndTime.UTC()

	return trace, nil
}
