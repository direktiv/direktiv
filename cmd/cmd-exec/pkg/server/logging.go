package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Logger struct {
	// httpLogEnabled, binaryLogEnabled bool
	LogData          bytes.Buffer
	actionID         string
	backendLogServer string

	w io.Writer
}

type httpLogger struct {
	backendLogServer, actionID string
}

const (
	devMode     = "DIREKTIV_DEV_MODE"
	httpBackend = "DIREKTIV_HTTP_BACKEND"
)

func NewLogger(actionID string) *Logger {
	backend := "http://localhost:8889"
	if os.Getenv(httpBackend) != "" {
		backend = os.Getenv(httpBackend)
	}

	l := &Logger{
		actionID:         actionID,
		backendLogServer: backend,
	}

	l.SetWriterState(true)

	return l
}

func (l *Logger) SetWriterState(enable bool) {
	wrs := []io.Writer{}

	if enable {
		wrs = append(wrs, os.Stdout, &l.LogData)
	}

	if enable && os.Getenv(devMode) == "" {
		wrs = append(wrs, &httpLogger{
			actionID:         l.actionID,
			backendLogServer: l.backendLogServer,
		})
	}

	l.w = io.MultiWriter(wrs...)
}

// nolint
func (l *Logger) Log(format string, args ...any) {
	l.Write([]byte(fmt.Sprintf(format+"\n", args...)))
}

func (l *Logger) Write(p []byte) (int, error) {
	return l.w.Write(p)
}

// nolint
func (l *httpLogger) Write(p []byte) (int, error) {
	_, err := http.Post(fmt.Sprintf("%s/log?aid=%s",
		l.backendLogServer, l.actionID), "plain/text", bytes.NewBuffer(p))

	return len(p), err
}
