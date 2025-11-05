package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/direktiv/direktiv/internal/telemetry"
)

type Logger struct {
	io.Writer

	LogData          bytes.Buffer
	actionID         string
	backendLogServer string

	lo telemetry.LogObject
}

var _ io.Writer = (*Logger)(nil)

const (
	devMode = "DIREKTIV_DEV_MODE"
)

// NewLogger creates a new Logger instance.
func NewLogger(logObject telemetry.LogObject, actionID string) *Logger {
	l := &Logger{
		actionID: actionID,
		lo:       logObject,
	}

	l.SetWriterState(true)

	return l
}

// SetWriterState configures the Logger's writer state, enabling or disabling certain outputs.
func (l *Logger) SetWriterState(enable bool) {
	writers := []io.Writer{}

	if enable {
		writers = append(writers, os.Stdout, &l.LogData)
	} else {
		writers = append(writers, &l.LogData)
	}

	if enable && os.Getenv(devMode) == "" {
		writers = append(writers, NewCtxLogger(
			l.lo,
			l.actionID,
		))
	}

	l.Writer = io.MultiWriter(writers...)
}

// Logf logs a formatted message, appending a newline, and writes it to the configured writers.
func (l *Logger) Logf(format string, args ...interface{}) {
	message := fmt.Sprintf(format+"\n", args...)
	if message == "\n" {
		message = ""
	}
	_, err := l.Write([]byte(message))
	if err != nil {
		slog.Error("failed to log message", slog.String("error", err.Error()))
	}
}

// Write writes the provided byte slice to the configured writers.
func (l *Logger) Write(p []byte) (int, error) {
	if l.Writer == nil {
		return 0, fmt.Errorf("no writer set")
	}

	return l.Writer.Write(p)
}

var _ io.Writer = (*ctxLogger)(nil)

func NewCtxLogger(
	logObject telemetry.LogObject,
	actionID string,
) io.Writer {
	return ctxLogger{
		actionID: actionID,
		lo:       logObject,
	}
}

type ctxLogger struct {
	actionID string
	lo       telemetry.LogObject
}

// Write sends the provided byte slice as a log message to the backend server.
func (l ctxLogger) Write(p []byte) (int, error) {
	ctx := telemetry.LogInitCtx(context.Background(), l.lo)
	telemetry.LogInstance(ctx, telemetry.LogLevelInfo, string(p))

	return len(p), nil
}
