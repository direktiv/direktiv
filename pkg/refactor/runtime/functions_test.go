package runtime_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/compiler"
	"github.com/direktiv/direktiv/pkg/refactor/runtime"
	"github.com/stretchr/testify/assert"
)

func TestFunctionID(t *testing.T) {

	s := startServerHttp(fnResponse)
	defer s.stop()

	var fn = make(map[string]interface{})
	fn["image"] = "localhost:5000/hello"
	fn["envs"] = map[string]string{
		"ENV1": "value",
	}

	// setting the env wit hthe function and its id
	fnID, _ := compiler.GenerateFunctionID(fn)
	// os.Setenv(fnID, fmt.Sprintf("http://127.0.0.1:%d", s.port))

	script := `	
	function start(state) {

		const f = setupFunction(
			{
				image: "localhost:5000/hello"
				envs: {
					"ENV1": "value"
				}
			}
		)

		return f.execute({
			input: ""
		})

	}
	`

	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	rt := createRuntime(t, script, true)
	secrets := make(map[string]string)
	secrets["secret1"] = "value1"
	rt.Secrets = &secrets

	fns := make(map[string]string)
	fns[fnID] = fmt.Sprintf("http://127.0.0.1:%d", s.port)
	rt.Functions = &fns

	w := httptest.NewRecorder()
	v, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, "true", v.(map[string]interface{})["success"])
}

func fnResponse(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("{ \"success\": \"true\" }"))
}

func TestFunctionFiles(t *testing.T) {
	s := startServerHttp(fileResponse)
	defer s.stop()

	script := `	
	function start(state) {

		const f = setupFunction(
			{
				image: "localhost:5000/hello"
				envs: {
					"ENV1": "value"
				}
			}
		)

		var file = getFile({
			name: "send.json",
			scope: "local"
		})
		file.write("test")


		return f.execute({
			file: file
		})

	}
	`

	var fn = make(map[string]interface{})
	fn["image"] = "localhost:5000/hello"
	fn["envs"] = map[string]string{
		"ENV1": "value",
	}

	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	rt := createRuntime(t, script, true)
	fnID, _ := compiler.GenerateFunctionID(fn)
	fns := make(map[string]string)
	fns[fnID] = fmt.Sprintf("http://127.0.0.1:%d", s.port)
	rt.Functions = &fns

	w := httptest.NewRecorder()
	v, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, "test", v)
}

func fileResponse(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	w.Write(b)
}

func TestFunctionDownload(t *testing.T) {

	s := startServerHttp(fileResponse)
	defer s.stop()

	script := `	
	function start(state) {

		const f = setupFunction(
			{
				image: "localhost:5000/hello"
				envs: {
					"ENV1": "value"
				}
			}
		)

		var data = {}
		data["test"] = "me"

		f.execute({
			input: data,
			asFile: true
		})

		var file = getFile({
			name: "result.data"
		})

		return file.data()

	}
	`

	var fn = make(map[string]interface{})
	fn["image"] = "localhost:5000/hello"
	fn["envs"] = map[string]string{
		"ENV1": "value",
	}

	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	rt := createRuntime(t, script, true)
	fnID, _ := compiler.GenerateFunctionID(fn)
	fns := make(map[string]string)
	fns[fnID] = fmt.Sprintf("http://127.0.0.1:%d", s.port)
	rt.Functions = &fns

	w := httptest.NewRecorder()
	v, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, "{\"test\":\"me\"}", v)

}

var countFn int

func retryResponse(w http.ResponseWriter, r *http.Request) {

	if countFn < 2 {
		w.WriteHeader(500)
	}
	countFn = countFn + 1
	w.Write([]byte(fmt.Sprintf("%d", countFn)))
}

func TestFunctionRetry(t *testing.T) {

	s := startServerHttp(retryResponse)
	defer s.stop()

	script := `	
	function start(state) {

		const f = setupFunction(
			{
				image: "localhost:5000/hello"
				envs: {
					"ENV1": "value"
				}
			}
		)

		return f.execute({
			retry: {
				count: 4,
				wait: 1
			}
		})
	}
	`

	var fn = make(map[string]interface{})
	fn["image"] = "localhost:5000/hello"
	fn["envs"] = map[string]string{
		"ENV1": "value",
	}

	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	rt := createRuntime(t, script, true)
	fnID, _ := compiler.GenerateFunctionID(fn)
	fns := make(map[string]string)
	fns[fnID] = fmt.Sprintf("http://127.0.0.1:%d", s.port)
	rt.Functions = &fns

	w := httptest.NewRecorder()
	v, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), v)

}

func erroResp(w http.ResponseWriter, r *http.Request) {

	w.Header().Add(runtime.DirektivErrorCodeHeader, "my.error.code")
	w.Header().Add(runtime.DirektivErrorMessageHeader, "not working")

}

func TestFunctionErrorHandling(t *testing.T) {

	s := startServerHttp(erroResp)
	defer s.stop()

	script := `	
	function start(state) {

		const f = setupFunction(
			{
				image: "localhost:5000/hello"
				envs: {
					"ENV1": "value"
				}
			}
		)
		try {
			return f.execute({})
		} catch (e) {
			return e.name
		}
	}
	`

	var fn = make(map[string]interface{})
	fn["image"] = "localhost:5000/hello"
	fn["envs"] = map[string]string{
		"ENV1": "value",
	}

	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	rt := createRuntime(t, script, true)
	fnID, _ := compiler.GenerateFunctionID(fn)
	fns := make(map[string]string)
	fns[fnID] = fmt.Sprintf("http://127.0.0.1:%d", s.port)
	rt.Functions = &fns

	w := httptest.NewRecorder()
	v, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, "my.error.code", v)
}

func timeoutResp(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
}

func TestFunctionErrorTimeout(t *testing.T) {

	s := startServerHttp(timeoutResp)
	defer s.stop()

	script := `	
	function start(state) {

		const f = setupFunction(
			{
				image: "localhost:5000/hello"
				envs: {
					"ENV1": "value"
				}
			}
		)
		try {
			return f.execute({
				timeout: 1
			})
		} catch (e) {
			return e.name
		}
	}
	`

	var fn = make(map[string]interface{})
	fn["image"] = "localhost:5000/hello"
	fn["envs"] = map[string]string{
		"ENV1": "value",
	}

	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	rt := createRuntime(t, script, true)
	fnID, _ := compiler.GenerateFunctionID(fn)
	fns := make(map[string]string)
	fns[fnID] = fmt.Sprintf("http://127.0.0.1:%d", s.port)
	rt.Functions = &fns

	w := httptest.NewRecorder()
	v, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, "io.direktiv.error.timeout", v)
}
