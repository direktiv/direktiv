package tsengine_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/tsengine"
	"github.com/stretchr/testify/assert"
)

func setupEngineenv(t *testing.T, flow string) *tsengine.RuntimeHandler {

	f, err := os.MkdirTemp("", "test")
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	_, _ = tsengine.New(f)

	// create flow
	f3, _ := os.Create(filepath.Join(f, "flow.ts"))
	f3.WriteString(flow)

	// create secrets
	os.MkdirAll(filepath.Join(f, "secrets"), 0777)
	s1, _ := os.Create(filepath.Join(f, "secrets", "secret1"))
	s1.WriteString("mysecret1")
	s2, _ := os.Create(filepath.Join(f, "secrets", "secret2"))
	s2.WriteString("mysecret2")

	// create files
	f1, _ := os.Create(filepath.Join(f, "file1.txt"))
	f1.WriteString("myfile1")
	f2, _ := os.Create(filepath.Join(f, "file2"))
	f2.WriteString("myfile2")

	h, _ := tsengine.CreateRuntimeHandler(tsengine.Config{
		BaseDir:  f,
		FlowPath: f3.Name()}, nil)
	return h

}

func TestBaseEngine(t *testing.T) {

	def := `
	function hello(state) {
		return state.data()["test"]
	}
	`

	e := setupEngineenv(t, def)

	req, _ := http.NewRequest("POST", "/dummy", strings.NewReader("{ \"test\": \"engine\"}"))
	resp := httptest.NewRecorder()

	e.ServeHTTP(resp, req)
	b, _ := io.ReadAll(resp.Result().Body)
	assert.Equal(t, "\"engine\"", string(b))
}

func TestEngineResponse(t *testing.T) {

	def := `
	function hello(state) {
		state.response().addHeader("Content-Type", "application/text")
		state.response().write("hello")
	}
	`

	e := setupEngineenv(t, def)

	req, _ := http.NewRequest("POST", "/dummy", strings.NewReader("{ \"test\": \"engine\"}"))
	resp := httptest.NewRecorder()

	e.ServeHTTP(resp, req)

	b, _ := io.ReadAll(resp.Result().Body)
	assert.Equal(t, "hello", string(b))
	assert.Equal(t, "application/text", resp.Header().Get("Content-type"))
}

func TestEngineStateURL(t *testing.T) {

	def := `
	function hello(state) {
		var r = {}

		r["h"] = state.getHeader("myheader")
		r["q1"] = state.getParam("query1")
		r["q2"] = state.getParam("query2")
		
		return r
	}
	`

	e := setupEngineenv(t, def)

	req, _ := http.NewRequest("POST", "/dummy?query1=one&query2=two", strings.NewReader("{ \"test\": \"engine\"}"))
	req.Header.Set("myheader", "myheadervalue")
	resp := httptest.NewRecorder()

	e.ServeHTTP(resp, req)

	b, _ := io.ReadAll(resp.Result().Body)
	var out map[string]string
	json.Unmarshal(b, &out)
	assert.Equal(t, "myheadervalue", out["h"])
	assert.Equal(t, "one", out["q1"])
	assert.Equal(t, "two", out["q2"])
}

func TestEngineStateFile(t *testing.T) {

	def := `

	const flow : FlowDefintion = {
		json: false
	}

	function hello(state) {
		var f = getFile({ 
			name: "input.data" 
		})
		state.response().writeFile(f)
	}
	`

	e := setupEngineenv(t, def)

	req, _ := http.NewRequest("POST", "/dummy?query1=one&query2=two", strings.NewReader("{ \"test\": \"engine\" }"))
	req.Header.Set("myheader", "myheadervalue")
	resp := httptest.NewRecorder()

	e.ServeHTTP(resp, req)

	b, _ := io.ReadAll(resp.Result().Body)
	var out map[string]string
	json.Unmarshal(b, &out)
	assert.Equal(t, "engine", out["test"])
}

func TestEngineNoJSON(t *testing.T) {

	def := `
	function hello(state) {
		return state.data()
	}
	`

	e := setupEngineenv(t, def)

	req, _ := http.NewRequest("POST", "/dummy?query1=one&query2=two", strings.NewReader("{{{"))
	req.Header.Set("myheader", "myheadervalue")
	resp := httptest.NewRecorder()

	e.ServeHTTP(resp, req)

	b, _ := io.ReadAll(resp.Result().Body)
	assert.Equal(t, "\"{{{\"", string(b))
}
