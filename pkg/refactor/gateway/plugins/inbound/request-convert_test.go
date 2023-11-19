package inbound_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/inbound"
	"github.com/stretchr/testify/assert"
)

func TestExecuteRequestConvertPlugin(t *testing.T) {

	r, _ := http.NewRequest(http.MethodGet, "/dummy?key=val&key=val2&hello=world",
		strings.NewReader("{ \"content\": \"value\" }"))
	r.Header.Add("header1", "value1")

	urlParams := make(map[string]string)
	urlParams["test"] = "value"
	ctx := context.WithValue(r.Context(), plugins.URLParamCtxKey, urlParams)

	p, _ := plugins.GetPluginFromRegistry(inbound.RequestConvertPluginName)
	p2, _ := p.Configure(&inbound.RequestConvertConfig{})

	w := httptest.NewRecorder()
	p2.ExecutePlugin(ctx, nil, w, r)

	b, _ := io.ReadAll(r.Body)

	var response inbound.RequestConvertResponse
	json.Unmarshal(b, &response)

	assert.ElementsMatch(t, []string{"val", "val2"}, response.QueryParams["key"])
	assert.ElementsMatch(t, []string{"world"}, response.QueryParams["hello"])
	assert.Equal(t, "value", response.URLParams["test"])
	assert.Equal(t, "value1", response.Headers.Get("header1"))

}

func TestExecuteRequestConvertPluginNoContent(t *testing.T) {

	r, _ := http.NewRequest(http.MethodGet, "/dummy", nil)

	w := httptest.NewRecorder()
	p, _ := plugins.GetPluginFromRegistry(inbound.RequestConvertPluginName)
	p2, _ := p.Configure(&inbound.RequestConvertConfig{})

	p2.ExecutePlugin(r.Context(), nil, w, r)

	b, _ := io.ReadAll(r.Body)

	var response inbound.RequestConvertResponse
	json.Unmarshal(b, &response)

	assert.Empty(t, response.URLParams)
	assert.Empty(t, response.QueryParams)
	assert.Equal(t, len(response.Body), 2)
	assert.Empty(t, response.Headers)
}

func TestExecuteRequestConvertPluginBinaryContent(t *testing.T) {

	r, _ := http.NewRequest(http.MethodGet, "/dummy",
		strings.NewReader("NONJSON"))

	p, _ := plugins.GetPluginFromRegistry(inbound.RequestConvertPluginName)
	p2, _ := p.Configure(&inbound.RequestConvertConfig{})

	w := httptest.NewRecorder()
	p2.ExecutePlugin(r.Context(), nil, w, r)

	b, _ := io.ReadAll(r.Body)

	var response inbound.RequestConvertResponse
	json.Unmarshal(b, &response)

	nj := base64.StdEncoding.EncodeToString([]byte("NONJSON"))
	assert.Equal(t, fmt.Sprintf("{\"data\":\"%s\"}", nj), string(response.Body))
}

func TestExecuteRequestConvertPluginSkip(t *testing.T) {

	r, _ := http.NewRequest(http.MethodGet, "/dummy?key=val&key=val2&hello=world",
		strings.NewReader("{ \"content\": \"value\" }"))
	r.Header.Add("header1", "value1")

	p, _ := plugins.GetPluginFromRegistry(inbound.RequestConvertPluginName)

	config := &inbound.RequestConvertConfig{
		OmitHeaders: true,
		OmitQueries: true,
		OmitBody:    true,
	}
	p2, _ := p.Configure(config)

	w := httptest.NewRecorder()
	p2.ExecutePlugin(r.Context(), nil, w, r)

	b, _ := io.ReadAll(r.Body)

	var response inbound.RequestConvertResponse
	json.Unmarshal(b, &response)

	fmt.Println(string(b))

	assert.Empty(t, response.Headers)
	assert.Empty(t, response.QueryParams)
	assert.Equal(t, json.RawMessage("{}"), response.Body)

}
