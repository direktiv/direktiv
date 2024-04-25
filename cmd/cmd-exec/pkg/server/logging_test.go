package server_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/cmd/cmd-exec/pkg/server"
	"github.com/stretchr/testify/assert"
)

func TestLogging(t *testing.T) {
	var b bytes.Buffer

	ts := startMockServer(&b)

	loggingText := "Hello World"

	l := server.NewLogger(ts.URL, "123")
	l.Logf(loggingText)

	// Log adds \n to log text
	assert.Equal(t, loggingText+"\n", b.String())

	// reset buffer
	b.Reset()

	l.Write([]byte(loggingText))
	assert.Equal(t, loggingText, b.String())

	ts.Close()
}

func TestMultipleLogEntries(t *testing.T) {
	var b bytes.Buffer
	ts := startMockServer(&b)

	loggingText1 := "First Log Entry"
	loggingText2 := "Second Log Entry"

	l := server.NewLogger(ts.URL, "123")
	l.Logf(loggingText1)
	l.Logf(loggingText2)

	expectedOutput := loggingText1 + "\n" + loggingText2 + "\n"
	assert.Equal(t, expectedOutput, b.String())

	ts.Close()
}

func TestLargeLogData(t *testing.T) {
	var b bytes.Buffer
	ts := startMockServer(&b)

	largeLoggingText := "A" + strings.Repeat("B", 1024*1024) // 1MB log data

	l := server.NewLogger(ts.URL, "123")
	l.Write([]byte(largeLoggingText))

	assert.Equal(t, largeLoggingText, b.String())

	ts.Close()
}

func TestEmptyLogMessage(t *testing.T) {
	var b bytes.Buffer
	ts := startMockServer(&b)

	l := server.NewLogger(ts.URL, "123")
	l.Logf("")

	assert.Equal(t, "", b.String())

	ts.Close()
}

func startMockServer(b *bytes.Buffer) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "got data\n")
		br, _ := io.ReadAll(r.Body)
		b.Write(br)
		defer r.Body.Close()
	})

	return httptest.NewServer(handler)
}
