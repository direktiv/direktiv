package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/cmd/cmd-exec/pkg/server"
	"github.com/stretchr/testify/assert"
)

type testIn struct {
	Dummy   string
	Integer int
}

var inData = `
{
	"Dummy": "Data",
	"Integer": 1
}
`

var inDataErr = `
{
	"Dummy": "Data",
	"Integer": 500
}
`

func TestNewServer(t *testing.T) {
	h := server.Handler[testIn](handleit)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(inData))

	h.ServeHTTP(w, r)

	ehc := w.Header().Get(server.DirektivErrorCodeHeader)
	ehm := w.Header().Get(server.DirektivErrorMessageHeader)

	assert.Equal(t, "io.direktiv.error.execution", ehc)
	assert.Equal(t, "no temp directory provided", ehm)

	r.Header.Add(server.DirektivTempDir, os.TempDir())

	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)

	ehc = w.Header().Get(server.DirektivErrorCodeHeader)
	ehm = w.Header().Get(server.DirektivErrorMessageHeader)

	assert.Equal(t, "io.direktiv.error.execution", ehc)
	assert.Equal(t, "no action id provided", ehm)

	w = httptest.NewRecorder()
	r.Header.Add(server.DirektivActionIDHeader, "123")
	h.ServeHTTP(w, r)

	var outData testIn
	json.Unmarshal(w.Body.Bytes(), &outData)

	assert.Equal(t, 200, outData.Integer)
}

func TestNewServerErrors(t *testing.T) {
	h := server.Handler[testIn](handleit)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("random data\nhello world"))
	r.Header.Add(server.DirektivActionIDHeader, "123")
	r.Header.Add(server.DirektivTempDir, os.TempDir())

	h.ServeHTTP(w, r)

	// fails
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)

	w = httptest.NewRecorder()
	r, _ = http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add(server.DirektivActionIDHeader, "123")
	r.Header.Add(server.DirektivTempDir, os.TempDir())

	h.ServeHTTP(w, r)

	// ok with empty
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	w = httptest.NewRecorder()
	r, _ = http.NewRequest(http.MethodPost, "/", strings.NewReader(inDataErr))
	r.Header.Add(server.DirektivActionIDHeader, "123")
	r.Header.Add(server.DirektivTempDir, os.TempDir())

	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func handleit(ctx context.Context, in testIn, ei *server.ExecutionInfo) (interface{}, error) {
	if in.Integer == 500 {
		return nil, fmt.Errorf("that does not work")
	}

	in.Dummy = "response"
	in.Integer = 200

	return in, nil
}

func TestNewServerFiles(t *testing.T) {
	payload := `
	{
		"files": [
			{
				"name": "hello.sh",
				"permission": 493,
				"content": "#!/bin/sh\necho -n hello"
			}
		]
	}
	`

	h := server.Handler[testIn](handleit)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
	r.Header.Add(server.DirektivActionIDHeader, "123")
	r.Header.Add(server.DirektivTempDir, os.TempDir())

	fp := filepath.Join(os.TempDir(), "hello.sh")

	h.ServeHTTP(w, r)

	cmd := exec.Command(fp)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	assert.NoError(t, err)

	h.ServeHTTP(w, r)
}
