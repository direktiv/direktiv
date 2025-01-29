package tracing_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/stretchr/testify/assert"
)

func TestConvertTracesToTimelines(t *testing.T) {
	// Load test data from traces.json
	file, err := os.ReadFile("traces.json")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Parse JSON
	var testData map[string][]map[string]interface{}
	err = json.Unmarshal(file, &testData)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	traces, exists := testData["data"]
	if !exists {
		t.Fatalf("Test data does not contain 'data' key")
	}

	// Call the function
	timeline, err := tracing.ConvertTracesToTimelines(traces)
	if err != nil {
		t.Fatalf("Function returned error: %v", err)
	}

	// Validate metadata
	assert.Equal(t, "5a2515340f1eaaafa3595d3bdd6cbe8d", timeline.Meta.TraceID, "TraceID should match")
	assert.Greater(t, timeline.Meta.TotalDuration, int64(0), "Total duration should be greater than zero")
	assert.Equal(t, "completed", timeline.Meta.Status, "Status should be completed")
	assert.Equal(t, "test", timeline.Meta.Namespace, "Namespace should be 'test'")

	// Validate root span count
	assert.Greater(t, len(timeline.Timeline), 0, "There should be at least one root span")

	// Ensure all spans are accounted for
	spanCount := countSpans(timeline.Timeline)
	expectedSpanCount := len(traces) // Each trace entry should have a corresponding span
	assert.Equal(t, expectedSpanCount, spanCount, "All spans should be included")

	// Validate hierarchy
	checkHierarchy(t, timeline.Timeline)
}

// Recursively count spans
func countSpans(spans []*tracing.SpanNode) int {
	count := len(spans)
	for _, span := range spans {
		count += countSpans(span.Children)
	}
	return count
}

// Recursively check hierarchy correctness
func checkHierarchy(t *testing.T, spans []*tracing.SpanNode) {
	for _, span := range spans {
		for _, child := range span.Children {
			assert.NotEqual(t, span.SpanID, child.SpanID, "Span should not be its own child")
			checkHierarchy(t, child.Children)
		}
	}
}
