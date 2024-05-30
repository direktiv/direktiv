package runtime_test

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilesIllegal(t *testing.T) {

	script := `	
	function start(state) {
		const f = getFile(
		{
			name: "../../../../../../whatever.txt",
			permission: 666,
		}
		)
	}
	`
	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	rt := createRuntime(t, script, true)
	w := httptest.NewRecorder()

	_, _, err := rt.Execute("start", req, w)
	assert.Error(t, err)

	script = `	
	function start(state) {
		const f = getFile(
		{
			name: "../../../../../../whatever.txt",
			permission: 666,
			scope: shared,
		}
		)
	}
	`
	req = &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	rt = createRuntime(t, script, true)
	w = httptest.NewRecorder()

	_, _, err = rt.Execute("start", req, w)
	assert.Error(t, err)

}

func TestFilesShared(t *testing.T) {

	script := `	
	function start(state) {
		try {
		const f = getFile(
		{
			name: "whatever.txt",
			permission: 666,
			scope: "shared"
		}
		)
		f.write("hello")
	} catch (e) {
		log(e)
	}
		return f.base64()
	}
	`
	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	rt := createRuntime(t, script, true)
	w := httptest.NewRecorder()

	v, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("hello")), v)
}

func TestFilesLocal(t *testing.T) {

	script := `	
	function start(state) {
		try {
		const f = getFile(
		{
			name: "whatever.txt",
			permission: 666,
		}
		)
		f.write("hello")
	} catch (e) {
		log(e)
	}
		return f.base64()
	}
	`
	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	rt := createRuntime(t, script, true)
	w := httptest.NewRecorder()

	v, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("hello")), v)
}
