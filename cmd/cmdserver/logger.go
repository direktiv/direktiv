package cmdserver

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
)

type Logger struct {
	LogData          bytes.Buffer
	actionID         string
	backendLogServer string
	io.Writer
}

var _ io.Writer = (*Logger)(nil)

func newLogger(httpBackend, actionID string) *Logger {
	l := &Logger{
		actionID:         actionID,
		backendLogServer: httpBackend,
	}

	writers := []io.Writer{}

	writers = append(writers, os.Stdout, &l.LogData)
	writers = append(writers, newHTTPLogger(
		l.backendLogServer,
		l.actionID,
		http.DefaultClient.Post,
	))

	l.Writer = io.MultiWriter(writers...)

	return l
}

func (l *Logger) Logf(format string, args ...interface{}) {
	message := fmt.Sprintf(format+"\n", args...)
	if message == "\n" {
		message = ""
	}
	_, err := l.Write([]byte(message))
	if err != nil {
		slog.Error("failed to log in cmd-exec", "error", err)
	}
}

func (l *Logger) Write(p []byte) (int, error) {
	if l.Writer == nil {
		return 0, fmt.Errorf("no writer set")
	}

	return l.Writer.Write(p)
}

var _ io.Writer = (*httpLogger)(nil)

func newHTTPLogger(
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

func (l httpLogger) Write(p []byte) (int, error) {
	//nolint:noctx
	resp, err := l.post(fmt.Sprintf("%s/log?aid=%s",
		l.backendLogServer, l.actionID), "plain/text", bytes.NewBuffer(p))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return len(p), nil
}
