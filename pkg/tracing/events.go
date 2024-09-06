package tracing

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(err error, format string, args ...interface{})
}

type TelemetryEvents struct {
	attr []interface{}
	ctx  context.Context //nolint: containedctx
}

// NewTelemetryEvents initializes a trace and warning/error metrics for log events.
func NewTelemetryEvents(ctx context.Context, name string) (context.Context, *TelemetryEvents, func(), error) {
	// Create a new tracing span
	ctx, end, err := NewSpan(ctx, name)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create span: %w", err)
	}

	logger := &TelemetryEvents{
		attr: getSlogAttributes(ctx),
		ctx:  ctx,
	}

	return ctx, logger, end, nil
}

func (l *TelemetryEvents) addEvent(msg string) {
	span := trace.SpanFromContext(l.ctx)
	if span.IsRecording() {
		span.AddEvent(msg)
	}
}

func (l *TelemetryEvents) Debugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.addEvent(msg)
	slog.DebugContext(l.ctx, msg, l.attr...)
}

func (l *TelemetryEvents) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.addEvent(msg)
	slog.InfoContext(l.ctx, msg, l.attr...)
}

func (l *TelemetryEvents) Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.addEvent(msg)
	slog.WarnContext(l.ctx, msg, l.attr...)
	// l.warnCounter.Add(ctx, 1)
}

func (l *TelemetryEvents) Errorf(err error, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	tags := append(l.attr, "error", err, "status", "error")
	l.addEvent(msg)
	slog.ErrorContext(l.ctx, msg, tags...)
	// l.errorCounter.Add(ctx, 1)
}
