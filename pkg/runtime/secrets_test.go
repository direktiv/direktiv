package runtime_test

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/runtime"
	"github.com/stretchr/testify/assert"
)

func TestSecretsBasic(t *testing.T) {

	script := `	
	function start(state) {
		const s = getSecret({
			name: "secret1"
		})

		const ret = {}
		ret["plain"] = s.string()
		ret["base64"] = s.base64()
		return ret
	}
	`
	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	rt := createRuntime(t, script, true)

	secrets := make(map[string]string)
	secrets["secret1"] = "value1"
	rt.Secrets = &secrets

	w := httptest.NewRecorder()
	v, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)

	m := v.(map[string]interface{})
	assert.Equal(t, "value1", m["plain"])
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("value1")), m["base64"])

}

func TestSecretsBasicFile(t *testing.T) {
	script := `	
	function start(state) {
		const s = getSecret({ name: "secret1"})
		s.file("test.sec", 0400)

		const f = getFile({
			name: "test.sec"
		})

		return f.base64()
		
	}
	`
	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	rt := createRuntime(t, script, true)

	secrets := make(map[string]string)
	secrets["secret1"] = "value1"
	rt.Secrets = &secrets

	w := httptest.NewRecorder()
	v, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("value1")), v)
}

func TestSecretsErrorName(t *testing.T) {

	script := `	
	function start(state) {
		try {
			const s = getSecret({})
		} catch (e) {
			return e.name
		}
	}
	`
	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	rt := createRuntime(t, script, true)

	secrets := make(map[string]string)
	secrets["secret1"] = "value1"
	rt.Secrets = &secrets

	w := httptest.NewRecorder()
	v, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, runtime.DirektivSecretsErrorCode, v)
}

func TestSecretsErrorNotExist(t *testing.T) {

	script := `	
	function start(state) {
		try {
			const s = getSecret({
				name: "doesnotexist"
			})
		} catch (e) {
			return e.msg
		}
	}
	`
	req := &http.Request{
		Body: io.NopCloser(strings.NewReader("")),
	}

	rt := createRuntime(t, script, true)

	secrets := make(map[string]string)
	secrets["secret1"] = "value1"
	rt.Secrets = &secrets

	w := httptest.NewRecorder()
	v, _, err := rt.Execute("start", req, w)
	assert.NoError(t, err)
	assert.Equal(t, "secret doesnotexist does not exist", v)
}
