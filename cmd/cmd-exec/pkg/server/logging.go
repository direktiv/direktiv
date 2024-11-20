package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Logger struct {
	LogData          bytes.Buffer
	actionID         string
	backendLogServer string
	io.Writer
}

var _ io.Writer = (*Logger)(nil)

const (
	devMode        = "DIREKTIV_DEV_MODE"
	httpBackend    = "DIREKTIV_HTTP_BACKEND"
	requestTimeout = 10 * time.Second
)

// NewLogger creates a new Logger instance.
func NewLogger(httpBackend, actionID string) *Logger {
	l := &Logger{
		actionID:         actionID,
		backendLogServer: httpBackend,
	}

	l.SetWriterState(true)

	return l
}

// SetWriterState configures the Logger's writer state, enabling or disabling certain outputs.
func (l *Logger) SetWriterState(enable bool) {
	writers := []io.Writer{}

	if enable {
		writers = append(writers, os.Stdout, &l.LogData)
	}

	if enable && os.Getenv(devMode) == "" {
		writers = append(writers, NewHTTPLogger(
			l.backendLogServer,
			l.actionID,
			http.DefaultClient.Post,
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

var _ io.Writer = (*httpLogger)(nil)

// NewHTTPLogger creates a new HTTP logger for sending log messages to a backend server.
func NewHTTPLogger(
	backendLogServer string,
	actionID string,
	post func(url, contentType string, body io.Reader) (*http.Response, error),
) io.Writer {
	return httpLogger{
		backendLogServer: backendLogServer,
		actionID:         actionID,
		post:             post,
	}
}

type httpLogger struct {
	backendLogServer string
	actionID         string
	post             func(url, contentType string, body io.Reader) (*http.Response, error)
}

// Write sends the provided byte slice as a log message to the backend server.
func (l httpLogger) Write(p []byte) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	reqBody := bytes.NewBuffer(p)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/log?aid=%s", l.backendLogServer, l.actionID),
		reqBody)
	if err != nil {
		slog.Error("failed to create HTTP request for logging", slog.String("error", err.Error()))
		return 0, err
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to send log message to backend", slog.String("error", err.Error()))
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("backend log server returned non-OK status",
			slog.Int("status_code", resp.StatusCode))

		return 0, fmt.Errorf("non-ok status from log server: %d", resp.StatusCode)
	}

	return len(p), nil
}
