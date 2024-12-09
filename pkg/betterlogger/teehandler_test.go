package betterlogger_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/direktiv/direktiv/pkg/betterlogger"
)

// mockHandler is a mock implementation of slog.Handler for testing purposes.
type mockHandler struct {
	enabled   bool
	handled   int
	withAttrs []slog.Attr
	withGroup string
	returnErr error
}

func (m *mockHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return m.enabled
}

func (m *mockHandler) Handle(ctx context.Context, rec slog.Record) error {
	m.handled++
	return m.returnErr
}

func (m *mockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	m.withAttrs = append(m.withAttrs, attrs...)
	return m
}

func (m *mockHandler) WithGroup(name string) slog.Handler {
	m.withGroup = name
	return m
}

// TestTeeHandler tests the behavior of TeeHandler with multiple handlers.
func TestTeeHandler(t *testing.T) {
	ctx := context.Background()
	rec := slog.Record{
		Message: "test message",
		Level:   slog.LevelInfo,
	}

	// Create two mock handlers
	mock1 := &mockHandler{enabled: true}
	mock2 := &mockHandler{enabled: true}

	// Create TeeHandler with the mock handlers
	tee := betterlogger.TeeHandler{mock1, mock2}

	// Test Enabled - should return true if first handler is enabled
	if !tee.Enabled(ctx, slog.LevelInfo) {
		t.Error("Expected TeeHandler to be enabled")
	}

	// Test Handle - both handlers should process the log record
	err := tee.Handle(ctx, rec)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if mock1.handled != 1 || mock2.handled != 1 {
		t.Errorf("Expected both handlers to process the log record, got mock1: %d, mock2: %d", mock1.handled, mock2.handled)
	}

	// Test error handling - should aggregate errors from both handlers
	mock1.returnErr = errors.New("error from mock1")
	mock2.returnErr = errors.New("error from mock2")

	err = tee.Handle(ctx, rec)
	if err == nil {
		t.Error("Expected error, but got nil")
	} else if !errors.Is(err, mock1.returnErr) || !errors.Is(err, mock2.returnErr) {
		t.Errorf("Expected aggregated errors, got: %v", err)
	}

	// Test WithAttrs - should propagate attrs to all handlers
	attrs := []slog.Attr{{Key: "key", Value: slog.StringValue("value")}}
	teeWithAttrs := tee.WithAttrs(attrs).(betterlogger.TeeHandler)

	for i, logger := range teeWithAttrs {
		if l, _ := logger.(*mockHandler); !l.withAttrs[0].Equal(attrs[0]) {
			t.Errorf("Handler %d did not receive correct attributes", i)
		}
	}

	// Test WithGroup - should propagate group name to all handlers
	groupName := "group"
	teeWithGroup, _ := tee.WithGroup(groupName).(betterlogger.TeeHandler)

	for i, logger := range teeWithGroup {
		if logger.(*mockHandler).withGroup != groupName {
			t.Errorf("Handler %d did not receive correct group name", i)
		}
	}
}
