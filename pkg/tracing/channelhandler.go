package tracing

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/google/uuid"
)

type chanHandler struct {
	attrs    []slog.Attr
	ch       chan metastore.LogEntry
	topic    string
	minLevel slog.Level
}

// NewChannelHandler creates a new chanHandler that writes log records to a Go channel
// chanHandler is a custom slog.Handler that writes log records to a Go channel.
// It holds a set of log attributes, a log topic, and a minimum logging level.
// The log records are sent as metastore.LogEntry through the provided channel.
func NewChannelHandler(ch chan metastore.LogEntry, attrs []slog.Attr, group string, minLevel slog.Level) slog.Handler {
	return &chanHandler{
		ch:       ch,
		attrs:    attrs,
		topic:    group,
		minLevel: minLevel,
	}
}

// Enabled implements slog.Handler.
func (h *chanHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Assume all levels are enabled; modify as needed
	return level >= h.minLevel
}

// Handle implements slog.Handler.
func (h *chanHandler) Handle(ctx context.Context, rec slog.Record) error {
	attr := GetAttributes(ctx)
	attrConv := map[string]string{}
	for k, v := range attr {
		attrConv[k] = fmt.Sprint(v)
	}
	// Create a LogRecord and send it to the channel
	record := metastore.LogEntry{
		Metadata:  attrConv,
		ID:        uuid.NewString(),
		Level:     int(rec.Level),
		Message:   rec.Message,
		Timestamp: rec.Time.UnixMilli(),
	}

	select {
	case h.ch <- record:
		// Successfully sent
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second):
		return fmt.Errorf("failed to send log record")
	}

	return nil
}

// WithAttrs implements slog.Handler.
func (h *chanHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make(map[string]slog.Attr, len(h.attrs))
	for _, v := range h.attrs {
		newAttrs[v.Key] = v
	}
	for _, a := range attrs {
		newAttrs[a.Key] = slog.Attr{
			Key:   a.Key,
			Value: a.Value,
		}
	}
	convAttr := make([]slog.Attr, 0, len(newAttrs))
	for _, v := range newAttrs {
		convAttr = append(convAttr, v)
	}

	return NewChannelHandler(h.ch, convAttr, h.topic, h.minLevel)
}

// WithGroup implements slog.Handler.
func (h *chanHandler) WithGroup(name string) slog.Handler {
	return NewChannelHandler(h.ch, h.attrs, name, h.minLevel)
}

var _ slog.Handler = &chanHandler{}
