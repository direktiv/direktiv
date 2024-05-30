package target_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/gateway/plugins/target"
	"github.com/stretchr/testify/assert"
)

func TestConfigInstaneResponsePlugin(t *testing.T) {
	config := target.InstantResponseConfig{
		StatusCode:    205,
		StatusMessage: "test",
	}

	p, _ := plugins.GetPluginFromRegistry(target.InstantResponsePluginName)
	p2, _ := p.Configure(config, core.SystemNamespace)

	configOut := p2.Config().(*target.InstantResponseConfig)
	assert.Equal(t, config.StatusCode, configOut.StatusCode)
	assert.Equal(t, config.StatusMessage, configOut.StatusMessage)
}

func TestExecuteInstantResponsePlugin(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(target.InstantResponsePluginName)

	p2, _ := p.Configure(nil, core.SystemNamespace)

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodGet, "/dummy", nil)
	p2.ExecutePlugin(nil, w, r)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, "This is the end!", w.Body.String())

	config := &target.InstantResponseConfig{
		StatusCode:    http.StatusInternalServerError,
		StatusMessage: "HELLO WORLD",
		ContentType:   "application/demo",
	}
	p2, _ = p.Configure(config, core.SystemNamespace)

	w = httptest.NewRecorder()
	p2.ExecutePlugin(nil, w, r)

	b, _ := httputil.DumpResponse(w.Result(), true)
	fmt.Println(string(b))

	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
	assert.Equal(t, "HELLO WORLD", w.Body.String())
	assert.Equal(t, "application/demo", w.Header().Get("Content-Type"))

	config = &target.InstantResponseConfig{
		StatusCode:    http.StatusOK,
		StatusMessage: "{ \"hello\": \"world\" }",
	}
	p2, _ = p.Configure(config, core.SystemNamespace)

	w = httptest.NewRecorder()
	p2.ExecutePlugin(nil, w, r)

	assert.JSONEq(t, "{ \"hello\": \"world\" }", w.Body.String())
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}
