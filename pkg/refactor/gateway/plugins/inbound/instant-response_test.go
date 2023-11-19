package inbound_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/inbound"
	"github.com/stretchr/testify/assert"
)

func TestExecuteInstantResponsePlugin(t *testing.T) {

	p, _ := plugins.GetPluginFromRegistry(inbound.InstantResponsePluginName)

	// config := &inbound.InstantResponseConfig{
	// }
	p2, _ := p.Configure(nil)

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodGet, "/dummy", nil)
	p2.ExecutePlugin(context.Background(), nil, w, r)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	assert.Equal(t, "This is the end!", w.Body.String())

	config := &inbound.InstantResponseConfig{
		StatusCode:    http.StatusInternalServerError,
		StatusMessage: "HELLO WORLD",
	}
	p2, _ = p.Configure(config)

	w = httptest.NewRecorder()
	p2.ExecutePlugin(context.Background(), nil, w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
	assert.Equal(t, "HELLO WORLD", w.Body.String())

	config = &inbound.InstantResponseConfig{
		StatusCode:    http.StatusOK,
		StatusMessage: "{ \"hello\": \"world\" }",
	}
	p2, _ = p.Configure(config)

	w = httptest.NewRecorder()
	p2.ExecutePlugin(context.Background(), nil, w, r)

	assert.JSONEq(t, "{ \"hello\": \"world\" }", w.Body.String())
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}
