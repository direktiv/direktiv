package runtime_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/runtime"
	"github.com/stretchr/testify/assert"
)

func TestHttpClientFail(t *testing.T) {

	// s := startServerHttp(failTwice)
	// defer s.stop()

	script := `
	function start() {
		httpRequest(
			{}
		)
	}
	`

	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}
	fns := make(map[string]string)
	sec := make(map[string]string)
	rt := createRuntime(t, sec, fns, script, true)

	w := httptest.NewRecorder()
	_, _, err := rt.Execute("start", req, w)
	assert.Error(t, err)
}

func TestHttpClientSimple(t *testing.T) {

	s := startServerHttp(serveSimple)
	defer s.stop()

	script := fmt.Sprintf(`
	function start() {
		var r = httpRequest(
			{
				url: "http://127.0.0.1:%d"
			}
		)
		return r["success"]
	}
	`, s.port)

	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	fns := make(map[string]string)
	sec := make(map[string]string)
	rt := createRuntime(t, sec, fns, script, true)

	w := httptest.NewRecorder()
	r, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, "true", r)
}

func TestHttpClientRetry(t *testing.T) {

	s := startServerHttp(failTwice)
	defer s.stop()

	script := fmt.Sprintf(`
	function start() {
		var r = httpRequest(
			{
				url: "http://127.0.0.1:%d",
				retry: {
					count: 5,
					wait: 1
				}
			}
		)
		return r["success"]
	}
	`, s.port)

	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	fns := make(map[string]string)
	sec := make(map[string]string)
	rt := createRuntime(t, sec, fns, script, true)

	w := httptest.NewRecorder()
	r, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, "true", r)
}

func TestHttpClientFileUpload(t *testing.T) {

	s := startServerHttp(serveFile)
	defer s.stop()

	script := fmt.Sprintf(`
	function start() {
		var f = getFile({
			name: "test.txt"
		})
		f.write("testme")

		var r = httpRequest(
			{
				url: "http://127.0.0.1:%d",
				retry: {
					count: 5,
					wait: 1
				},
				file: f
			}
		)
		return r
	}
	`, s.port)

	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	fns := make(map[string]string)
	sec := make(map[string]string)
	rt := createRuntime(t, sec, fns, script, true)

	w := httptest.NewRecorder()
	r, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, "testme", r)
}

func TestHttpClientFileDownload(t *testing.T) {

	s := startServerHttp(serveFile)
	defer s.stop()

	script := fmt.Sprintf(`
	function start() {
		var r = httpRequest(
			{
				url: "http://127.0.0.1:%d",
				retry: {
					count: 5,
					wait: 1
				},
				result: "file",
				input: "response"
			}
		)
		var f = getFile({
			name: "result.data"
		})
		return f.data()
	}
	`, s.port)

	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	fns := make(map[string]string)
	sec := make(map[string]string)
	rt := createRuntime(t, sec, fns, script, true)

	w := httptest.NewRecorder()
	r, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, "response", r)
}

func TestHttpClientTimeout(t *testing.T) {

	s := startServerHttp(timeout)
	defer s.stop()

	script := fmt.Sprintf(`
	function start() {
		var r = httpRequest(
			{
				url: "http://127.0.0.1:%d",
				timeout: 1
			}
		)
		return r
	}
	`, s.port)

	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	fns := make(map[string]string)
	sec := make(map[string]string)
	rt := createRuntime(t, sec, fns, script, true)

	w := httptest.NewRecorder()
	_, _, err := rt.Execute("start", req, w)

	e := err.(*runtime.DirektivError).Code()
	assert.Equal(t, runtime.DirektivTimeoutErrorCode, e)
}

type httpServer struct {
	srv  *http.Server
	port int
}

func startServerHttp(f func(http.ResponseWriter, *http.Request)) *httpServer {

	listener, _ := net.Listen("tcp", ":0")
	port := listener.Addr().(*net.TCPAddr).Port

	l := fmt.Sprintf(":%d", port)
	listener.Close()

	mux := &http.ServeMux{}

	srv := &http.Server{Addr: l, Handler: mux}

	mux.HandleFunc("/", f)

	s := &httpServer{
		port: port,
		srv:  srv,
	}
	fmt.Printf("serving at %d\n", s.port)

	go func() {
		srv.ListenAndServe()
	}()

	return s

}

func (srv *httpServer) stop() {
	srv.srv.Shutdown(context.Background())
}

func timeout(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	w.Write([]byte("wait"))
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	w.Write(b)
}

func serveSimple(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{ \"success\": \"true\" }"))
}

var counter int

func failTwice(w http.ResponseWriter, r *http.Request) {

	if counter < 1 {
		w.WriteHeader(500)
	}

	w.Write([]byte("{ \"success\": \"true\" }"))

	counter = counter + 1

}
