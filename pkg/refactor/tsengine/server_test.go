package tsengine_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/compiler"
	"github.com/direktiv/direktiv/pkg/refactor/tsengine"
	"github.com/stretchr/testify/assert"
)

func TestServerMuxGetter(t *testing.T) {

	srv := startServerHttp(action)
	defer srv.stop()

	script := `
	var secret = getSecret({ name: "secret1" })

	var fn = setupFunction({
		image: "localhost:5000/hello"
	})

	function start(state) {

		var f = getFile({
			name: "file1",
			scope: "shared"
		})

		var result =  fn.execute({
			input: state.data()["data"]
		})

		var r = {}
		r["content"] = f.data()
		r["secret"] = secret.string()
		r["value"] = result["return"]

		return r
	}`

	// flowPath := "/path"
	f, _ := os.MkdirTemp("", "test")
	os.Setenv("DIREKTIV_JSENGINE_FLOWPATH", "/mypath")
	os.Setenv("DIREKTIV_JSENGINE_BASEDIR", f)

	s, _ := tsengine.NewServer()

	// function
	var fn = make(map[string]interface{})
	fn["image"] = "localhost:5000/hello"
	fnID, _ := compiler.GenerateFunctionID(fn)
	os.Setenv(fnID, fmt.Sprintf("http://127.0.0.1:%d", srv.port))

	// secrets
	secrets := make(map[string]string)
	secrets["secret1"] = "secret1"
	secrets["secret2"] = "secret2"

	// files
	files := make(map[string]io.Reader)
	files["file1"] = strings.NewReader("file1")
	files["file2"] = strings.NewReader("file2")

	pr, wr, _ := tsengine.CreateMultiPartForm(s.Prefix(), script, "/mypath", secrets, files)

	// request should not be ready
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status", nil)
	s.HandleStatusRequest(w, req)
	sr, _ := io.ReadAll(w.Result().Body)
	var status tsengine.Status
	json.Unmarshal(sr, &status)
	assert.False(t, status.Initialized)

	// run init request manually
	req, _ = http.NewRequest("POST", "/init", pr)
	w = httptest.NewRecorder()
	req.Header.Set("Content-Type", wr.FormDataContentType())
	mi := s.Initializer().(*tsengine.MuxInitializer)
	mi.HandleInitRequest(w, req)

	// init is true now
	req, _ = http.NewRequest("GET", "/status", nil)
	w = httptest.NewRecorder()
	s.HandleStatusRequest(w, req)
	sr, _ = io.ReadAll(w.Result().Body)
	json.Unmarshal(sr, &status)
	assert.True(t, status.Initialized)

	// run flow
	req, _ = http.NewRequest("POST", "/", strings.NewReader(string("{ \"data\": \"coming-in\"}")))
	w = httptest.NewRecorder()

	s.Engine.RunRequest(req, w)
	r, _ := io.ReadAll(w.Result().Body)

	var m map[string]string
	json.Unmarshal(r, &m)

	assert.Equal(t, "secret1", m["secret"])
	assert.Equal(t, "coming-in", m["value"])
	assert.Equal(t, "file1", m["content"])
}

func TestServerFileGetter(t *testing.T) {

	srv := startServerHttp(action)
	defer srv.stop()

	script := `
	var fn = setupFunction({
		image: "localhost:5000/hello"
	})

	const secret = getSecret({
		name: "mysecret"
	})

	function start(state) {
		var result =  fn.execute({
			input: state.data()["data"]
		})

		var r = {}
		r["secret"] = secret.string()
		r["value"] = result["return"]

		return r
	}
	`

	f, _ := os.MkdirTemp("", "test")
	scriptPath := filepath.Join(f, "myflow.ts")
	os.WriteFile(scriptPath, []byte(script), 0755)

	// add secrets
	os.Mkdir(filepath.Join(f, "secrets"), 0755)
	os.WriteFile(filepath.Join(f, "secrets", "mysecret"), []byte("secretvalue"), 0755)
	os.Setenv("DIREKTIV_JSENGINE_INIT", "file")
	os.Setenv("DIREKTIV_JSENGINE_FLOWPATH", scriptPath)
	os.Setenv("DIREKTIV_JSENGINE_BASEDIR", f)

	var fn = make(map[string]interface{})
	fn["image"] = "localhost:5000/hello"
	fnID, _ := compiler.GenerateFunctionID(fn)
	os.Setenv(fnID, fmt.Sprintf("http://127.0.0.1:%d", srv.port))

	s, _ := tsengine.NewServer()
	s.Initializer().Init()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dummy", strings.NewReader(string("{ \"data\": \"coming-in\"}")))

	s.Engine.RunRequest(req, w)

	r, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, "{\"secret\":\"secretvalue\",\"value\":\"coming-in\"}", string(r))

}

func action(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	w.Write([]byte(fmt.Sprintf("{ \"return\": \"%s\" }", string(b))))
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
