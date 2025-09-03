package datasql_test

import (
	"context"
	"testing"
	"time"

	"github.com/direktiv/direktiv/internal/datastore"
	"github.com/direktiv/direktiv/internal/datastore/datasql"
	database2 "github.com/direktiv/direktiv/pkg/database"
	"github.com/google/uuid"
)

func TestTraceStoreOperations(t *testing.T) {
	t.Run("testGetNonExistentTrace", testGetNonExistentTrace)
	t.Run("AddDeleteAndGet", testAddDeleteAndGetTrace)
	t.Run("GetByParentSpanID", testGetByParentSpanID)
	t.Run("testBatchInsertTraces", testBatchInsertTraces)
}

func testGetNonExistentTrace(t *testing.T) {
	traceStore, _ := setupTestTrace(t)

	nonExistentTraceID := uuid.NewString()
	_, err := traceStore.GetByTraceID(context.Background(), nonExistentTraceID)
	if err == nil {
		t.Errorf("expected error when retrieving non-existent trace, but got none")
	}
}

func testAddDeleteAndGetTrace(t *testing.T) {
	traceStore, trace := setupTestTrace(t)
	addTestTrace(t, traceStore, trace)
	verifyTraceAdded(t, traceStore, trace)

	retrievedTrace := getTraceByID(t, traceStore, trace.TraceID)
	verifyTraceRetrieved(t, retrievedTrace, trace)
	testDeleteOldTrace(t, traceStore)
	verifyTraceDeleted(t, traceStore, trace.TraceID)
}

func testDeleteOldTrace(t *testing.T, traceStore datastore.TracesStore) {
	// Set a cutoff time in the past to delete the trace
	cutoffTime := time.Now().UTC().Add(time.Hour)
	err := traceStore.DeleteOld(context.Background(), cutoffTime)
	if err != nil {
		t.Fatalf("error deleting old trace: %v", err)
	}
}

func testGetByParentSpanID(t *testing.T) {
	traceStore, trace := setupTestTrace(t)

	// Adding multiple traces with the same parentSpanID
	parentSpanID := uuid.NewString()
	trace.ParentSpanID = &parentSpanID
	addTestTrace(t, traceStore, trace)

	// Create another trace with the same ParentSpanID
	trace2 := createTestTrace()
	trace2.ParentSpanID = &parentSpanID
	addTestTrace(t, traceStore, trace2)

	// Retrieve traces by ParentSpanID
	traces, err := traceStore.GetByParentSpanID(context.Background(), parentSpanID)
	if err != nil {
		t.Fatalf("error getting traces by parent span ID: %v", err)
	}
	if len(traces) != 2 {
		t.Errorf("expected 2 traces, got %d", len(traces))
	}

	// Verify that the traces have the correct parentSpanID
	for _, trace := range traces {
		if *trace.ParentSpanID != parentSpanID {
			t.Errorf("expected ParentSpanID to be %v, got %v", parentSpanID, *trace.ParentSpanID)
		}
	}
}

func setupTestTrace(t *testing.T) (datastore.TracesStore, *datastore.Trace) {
	conn, err := database2.NewTestDB(t)
	if err != nil {
		t.Fatalf("unepxected NewTestDB() error = %v", err)
	}

	traceStore := datasql.NewStore(conn).Traces()

	trace := createTestTrace()

	return traceStore, trace
}

func createTestTrace() *datastore.Trace {
	return &datastore.Trace{
		TraceID:   uuid.NewString(),
		SpanID:    uuid.NewString(),
		StartTime: time.Now().UTC().Add(-10 * time.Hour),
		Metadata:  []byte("{}"), // Test with empty raw trace data
	}
}

func addTestTrace(t *testing.T, traceStore datastore.TracesStore, trace *datastore.Trace) {
	err := traceStore.Append(context.Background(), *trace)
	if err != nil {
		t.Fatalf("error appending test trace: %v", err)
	}
}

func verifyTraceAdded(t *testing.T, store datastore.TracesStore, trace *datastore.Trace) {
	res, err := store.GetByTraceID(context.Background(), trace.TraceID)
	if err != nil {
		t.Errorf("error verifying trace added: %v", err)
	}
	if res.TraceID != trace.TraceID {
		t.Error("ID did not match")
	}
}

func getTraceByID(t *testing.T, store datastore.TracesStore, traceID string) *datastore.Trace {
	retrievedTrace, err := store.GetByTraceID(context.Background(), traceID)
	if err != nil {
		t.Errorf("error retrieving trace by ID: %v", err)
	}
	return &retrievedTrace
}

const timeTolerance = time.Millisecond

func verifyTraceRetrieved(t *testing.T, retrievedTrace, expectedTrace *datastore.Trace) {
	if retrievedTrace.TraceID != expectedTrace.TraceID {
		t.Errorf("retrieved TraceID does not match expected: got %v, want %v", retrievedTrace.TraceID, expectedTrace.TraceID)
	}
	if retrievedTrace.SpanID != expectedTrace.SpanID {
		t.Errorf("retrieved SpanID does not match expected: got %v, want %v", retrievedTrace.SpanID, expectedTrace.SpanID)
	}
	if (retrievedTrace.ParentSpanID == nil && expectedTrace.ParentSpanID != nil) ||
		(retrievedTrace.ParentSpanID != nil && expectedTrace.ParentSpanID == nil) ||
		(retrievedTrace.ParentSpanID != nil && expectedTrace.ParentSpanID != nil && *retrievedTrace.ParentSpanID != *expectedTrace.ParentSpanID) {
		t.Errorf("retrieved ParentSpanID does not match expected: got %v, want %v", retrievedTrace.ParentSpanID, expectedTrace.ParentSpanID)
	}

	// Allow for minor differences in StartTime
	if retrievedTrace.StartTime.Sub(expectedTrace.StartTime) > timeTolerance ||
		expectedTrace.StartTime.Sub(retrievedTrace.StartTime) > timeTolerance {
		t.Errorf("retrieved StartTime does not match expected: got %v, want %v", retrievedTrace.StartTime, expectedTrace.StartTime)
	}

	if retrievedTrace.EndTime.Sub(expectedTrace.EndTime) > timeTolerance ||
		expectedTrace.EndTime.Sub(retrievedTrace.EndTime) > timeTolerance {
		t.Errorf("retrieved EndTime does not match expected: got %v, want %v", retrievedTrace.EndTime, expectedTrace.EndTime)
	}

	if string(retrievedTrace.Metadata) != string(expectedTrace.Metadata) {
		t.Errorf("retrieved RawTrace does not match expected: got %s, want %s", string(retrievedTrace.Metadata), string(expectedTrace.Metadata))
	}
}

func verifyTraceDeleted(t *testing.T, store datastore.TracesStore, traceID string) {
	_, err := store.GetByTraceID(context.Background(), traceID)
	if err == nil {
		t.Error("expected trace to be deleted")
	}
}

func testBatchInsertTraces(t *testing.T) {
	traceStore, _ := setupTestTrace(t)

	traces := []datastore.Trace{
		*createTestTrace(),
		*createTestTrace(),
		*createTestTrace(),
	}

	// Insert batch
	err := traceStore.Append(context.Background(), traces...)
	if err != nil {
		t.Fatalf("error inserting batch of traces: %v", err)
	}

	// Verify each trace exists
	for _, trace := range traces {
		retrievedTrace := getTraceByID(t, traceStore, trace.TraceID)
		verifyTraceRetrieved(t, retrievedTrace, &trace)
	}
}
