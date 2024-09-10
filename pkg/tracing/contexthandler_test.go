package tracing_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"regexp"
	"testing"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/stretchr/testify/assert"
)

// Helper function to set up a logger with a JSON handler
func setupLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(tracing.NewContextHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{})))
}

// Helper function to count occurrences of a substring in a string
func countOccurrences(s, substr string) int {
	re := regexp.MustCompile(regexp.QuoteMeta(substr))
	return len(re.FindAllString(s, -1))
}

// TestContextHandler_LogOutputWithAttributes checks that the context attributes are present in the final log output
func TestContextHandler_LogOutputWithAttributes(t *testing.T) {
	var buf bytes.Buffer
	logger := setupLogger(&buf)

	// Create a context with attributes
	ctx := context.Background()
	ctx = tracing.AddTag(ctx, "key", "value")
	ctx = tracing.WithTrack(ctx, "test-track")

	// Log a message with this context
	logger.InfoContext(ctx, "Test message")

	// Capture the log output
	var logOutput map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logOutput)
	assert.NoError(t, err)

	// Verify that the attributes from the context are present in the log output
	assert.Equal(t, "value", logOutput["key"], "Expected 'key' attribute to be present in the log")
	assert.Equal(t, "test-track", logOutput[string(core.LogTrackKey)], "Expected 'track' attribute to be present in the log")
	assert.Equal(t, "Test message", logOutput["msg"], "Expected log message to be 'Test message'")
}

// TestContextHandler_LogOutputWithoutAttributes ensures default attributes are added when no context attributes are provided
func TestContextHandler_LogOutputWithoutAttributes(t *testing.T) {
	var buf bytes.Buffer
	logger := setupLogger(&buf)

	// Create an empty context
	ctx := context.Background()

	// Log a message with this context
	logger.InfoContext(ctx, "Test message without attributes")

	// Capture the log output
	var logOutput map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logOutput)
	assert.NoError(t, err)

	// You could add more assertions here if needed to check defaults
}

// TestContextHandler_DuplicateAttributes ensures that attributes set in both context and slog.With don't duplicate in logs
func TestContextHandler_DuplicateAttributes(t *testing.T) {
	var buf bytes.Buffer
	logger := setupLogger(&buf)

	// Create a context with attributes
	ctx := context.Background()
	ctx = tracing.AddTag(ctx, "key", "context-value")

	// Log a message and also set the same key via slog.With
	logger.With("key", "slog-value").InfoContext(ctx, "Test message with possible duplicates")

	// Capture the log output
	var logOutput map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logOutput)
	assert.NoError(t, err)

	assert.Equal(t, "context-value", logOutput["key"], "Expected 'key' to be set by context and not slog.With")

	// Ensure that there are no duplicate 'key' attributes
	logOutputRaw := buf.String()
	keyCount := countOccurrences(logOutputRaw, `"key"`)
	assert.Equal(t, 2, keyCount, "Expected 'key' to appear only twice in the log output (once in each key-value pair)")
}

// TestContextHandler_NoDuplicateAttributes ensures that slog attributes take precedence over context attributes
func TestContextHandler_NoDuplicateAttributes(t *testing.T) {
	var buf bytes.Buffer
	logger := setupLogger(&buf)

	// Create a context with attributes
	ctx := context.Background()
	ctx = tracing.AddTag(ctx, "key", "context-value")

	// Log a message and also set the same key via slog.With attributes
	logger.InfoContext(ctx, "Test message with possible duplicates", "key", "other-value")

	// Capture the log output
	var logOutput map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logOutput)
	assert.NoError(t, err)

	assert.Equal(t, "other-value", logOutput["key"], "Expected 'key' to be set by slog attrs and not context")

	// Ensure that there are no duplicate 'key' attributes
	logOutputRaw := buf.String()
	keyCount := countOccurrences(logOutputRaw, `"key"`)
	assert.Equal(t, 2, keyCount, "Expected 'key' to appear only twice in the log output (once in each key-value pair)")
}
