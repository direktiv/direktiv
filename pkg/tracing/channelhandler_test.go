package tracing_test

import (
	"context"
	"testing"
	"time"

	"log/slog"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/direktiv/direktiv/pkg/tracing"
)

// TestEnabled checks if logging is enabled based on the minimum log level
func TestEnabled(t *testing.T) {
	ch := make(chan metastore.LogEntry, 1)
	handler := tracing.NewChannelHandler(ch, nil, "testGroup", slog.LevelInfo)

	tests := []struct {
		level    slog.Level
		expected bool
	}{
		{slog.LevelDebug, false},
		{slog.LevelInfo, true},
		{slog.LevelError, true},
	}

	for _, tt := range tests {
		if got := handler.Enabled(context.Background(), tt.level); got != tt.expected {
			t.Errorf("Enabled(%v) = %v, want %v", tt.level, got, tt.expected)
		}
	}
}

// TestHandle checks if log records are handled correctly
func TestHandle(t *testing.T) {
	ch := make(chan metastore.LogEntry, 1)
	handler := tracing.NewChannelHandler(ch, nil, "testGroup", slog.LevelInfo)

	ctx := context.WithValue(context.Background(), tracing.LogTrackKey, "customTopic")

	record := slog.Record{
		Level:   slog.LevelInfo,
		Message: "Test log message",
		Time:    time.Now(),
	}

	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("Handle() returned an unexpected error: %v", err)
	}

	select {
	case logRecord := <-ch:
		if logRecord.Message != "Test log message" {
			t.Errorf("expected message %v, got %v", "Test log message", logRecord.Message)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Handle() did not send log record to the channel")
	}
}

// TestHandle_Timeout checks if log handling times out after 1 second
func TestHandle_Timeout(t *testing.T) {
	ch := make(chan metastore.LogEntry)
	handler := tracing.NewChannelHandler(ch, nil, "testGroup", slog.LevelInfo)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	record := slog.Record{
		Level:   slog.LevelInfo,
		Message: "Test log message",
		Time:    time.Now(),
	}

	err := handler.Handle(ctx, record)
	if err == nil {
		t.Fatal("expected a timeout error, got none")
	}
}

// TestWithAttrs ensures that the handler includes new attributes correctly
func TestWithAttrs(t *testing.T) {
	ch := make(chan metastore.LogEntry, 1)
	handler := tracing.NewChannelHandler(ch, []slog.Attr{{Key: "attr1", Value: slog.StringValue("test")}}, "testGroup", slog.LevelInfo)

	newHandler := handler.WithAttrs([]slog.Attr{{Key: "attr1", Value: slog.StringValue("test")}})

	// Use reflection or manually inspect new attributes if needed
	if newHandler == handler {
		t.Errorf("WithAttrs() should return a new handler, got same instance")
	}
}

// TestWithGroup checks if a new handler with the new group is created
func TestWithGroup(t *testing.T) {
	ch := make(chan metastore.LogEntry, 1)
	handler := tracing.NewChannelHandler(ch, nil, "testGroup", slog.LevelInfo)

	newHandler := handler.WithGroup("newGroup")

	if newHandler == handler {
		t.Errorf("WithGroup() should return a new handler, got same instance")
	}
}
