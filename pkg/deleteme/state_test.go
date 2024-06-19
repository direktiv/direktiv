package deleteme

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testDummyErrorWriter struct {
}

func (tr *testDummyErrorWriter) Header() http.Header {
	return http.Header{}
}

func (tr *testDummyErrorWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("this failed")
}

func (tr *testDummyErrorWriter) WriteHeader(statusCode int) {
}

func TestStateWriteError(t *testing.T) {

	script := `
	function start(state) {
		try {
		 	state.response().write("hello")
		} catch (e) {
			return e
		}
	}
	`
	sec := make(map[string]string)
	fns := make(map[string]string)

	rt := createRuntime(t, sec, fns, script, true)
	w := &testDummyErrorWriter{}
	req, _ := http.NewRequest("POST", "/", strings.NewReader(""))

	r, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)

	ret := r.(map[string]interface{})
	assert.Equal(t, ret["code"], "io.direktiv.error.file")
}

func TestStateInputHeaders(t *testing.T) {

	script := `
	function start(state) {
		return state.getHeader("test")
	}
	`

	h := http.Header{}
	h.Add("test", "value1")
	h.Add("test", "value2")
	req := &http.Request{
		Body:   io.NopCloser(strings.NewReader("")),
		Header: h,
	}

	sec := make(map[string]string)
	fns := make(map[string]string)

	rt := createRuntime(t, sec, fns, script, true)
	w := httptest.NewRecorder()

	r, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, "value1", r)

	script = `
	function start(state) {
		return state.getHeaderValues("test")
	}
	`

	sec = make(map[string]string)
	fns = make(map[string]string)

	rt = createRuntime(t, sec, fns, script, true)
	r, _, err = rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, h.Values("test"), r)

}

func TestStateInputParams(t *testing.T) {

	script := `
	function start(state) {
		return state.getParam("test")
	}
	`

	u, _ := url.ParseRequestURI("http://dummy/url?test=value1&test=value2")
	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
		URL:  u,
	}
	sec := make(map[string]string)
	fns := make(map[string]string)
	rt := createRuntime(t, sec, fns, script, true)
	w := httptest.NewRecorder()

	r, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, "value1", r)

	script = `
	function start(state) {
		return state.getParamValues("test")
	}
	`

	sec = make(map[string]string)
	fns = make(map[string]string)

	rt = createRuntime(t, sec, fns, script, true)
	r, _, err = rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, u.Query()["test"], r)
}

func TestStateInputData(t *testing.T) {

	script := `
	const flow : FlowDefintion = {
		json: true
	}
	
	function start(state) {
		return state.data()["my"]
	}
	`
	content := "{ \"my\": \"data\"}"
	req := &http.Request{
		Body: io.NopCloser(strings.NewReader(content)),
	}

	sec := make(map[string]string)
	fns := make(map[string]string)

	rt := createRuntime(t, sec, fns, script, true)
	w := httptest.NewRecorder()

	r, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, "data", r)
}

func TestStateInputFile(t *testing.T) {

	script := `
	const flow : FlowDefintion = {
		json: false
	}
	
	function start(state) {
		const f = getFile({
			name: "input.data"
		})
		return f.base64()
	}
	`
	content := "data going in"
	req := &http.Request{
		Body: io.NopCloser(strings.NewReader(content)),
	}
	sec := make(map[string]string)
	fns := make(map[string]string)

	rt := createRuntime(t, sec, fns, script, false)
	w := httptest.NewRecorder()

	r, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t,
		base64.StdEncoding.EncodeToString([]byte(content)), r)

}

func TestStateResponseWrite(t *testing.T) {

	script := `
	const hello = "world"
	
	function start(state) {
		state.response().addHeader("hello", "world1")
		state.response().addHeader("hello", "world2")
		state.response().setHeader("second", "value")
		state.response().setStatus(201)
		state.response().write("this is a test value")

		return "testme"
	}
	`
	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	sec := make(map[string]string)
	fns := make(map[string]string)

	rt := createRuntime(t, sec, fns, script, true)
	w := httptest.NewRecorder()

	r, _, err := rt.Execute("start", req, w)

	assert.NoError(t, err)

	// response written, so no return value
	assert.Empty(t, r)

	assert.Equal(t, 201, w.Result().StatusCode)
	assert.Equal(t, w.Header().Get("second"), "value")
	assert.Len(t, w.Header().Values("hello"), 2)

	b, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, "this is a test value", string(b))
}
