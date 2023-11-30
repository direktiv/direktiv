package inbound_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/inbound"
	"github.com/stretchr/testify/assert"
)

func TestExecuteJSInboundPlugin(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(inbound.JSInboundPluginName)

	config := &inbound.JSInboundConfig{
		Script: `
		log("JENS")
		input["Headers"].Delete("Header1")
		input["Headers"].Add("demo", "value")
		input["Queries"].Add("new", "param")
		b = JSON.parse(input["Body"])
		b["newvalue"] = 200
		input["Body"] = JSON.stringify(b) 
		`,
	}
	p2, _ := p.Configure(config, core.MagicalGatewayNamespace)

	r, _ := http.NewRequest(http.MethodGet, "/dummy?test=me", nil)
	r.Header.Add("Header1", "value1")
	r.Header.Add("header2", "value2")
	r.Body = io.NopCloser(bytes.NewBufferString("{ \"string1\": \"value2\" }"))

	w := httptest.NewRecorder()
	p2.ExecutePlugin(nil, w, r)

	fmt.Println(r.URL)
	assert.Equal(t, "me", r.URL.Query().Get("test"))
	assert.Equal(t, "param", r.URL.Query().Get("new"))

	assert.Equal(t, "value", r.Header.Get("demo"))
	assert.Empty(t, r.Header.Get("Header1"))

	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	assert.JSONEq(t, "{\"string1\":\"value2\",\"newvalue\":200}", string(b))
}
