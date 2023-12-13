package outbound_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/outbound"
	"github.com/stretchr/testify/assert"
)

func TestJSOutboundPlugin(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(outbound.JSOutboundPluginName)

	config := &outbound.JSOutboundConfig{
		Script: `
		input["Headers"].Delete("Header1")
		input["Headers"].Add("demo", "value")
		input["Headers"].Add("demo2", "value2")
		input["Code"] = 204
		`,
	}
	p2, _ := p.Configure(config, core.MagicalGatewayNamespace)

	r, _ := http.NewRequest(http.MethodGet, "/dummy", nil)
	r.Header.Add("header1", "value1")
	r.Header.Add("header2", "value2")
	r.Body = io.NopCloser(bytes.NewBufferString("{ \"string1\": \"value2\" }"))

	w := gateway.NewDummyWriter()
	p2.ExecutePlugin(nil, w, r)

	assert.Equal(t, 204, w.Code)
}

func TestJsonModJSOutboundPlugin(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(outbound.JSOutboundPluginName)

	config := &outbound.JSOutboundConfig{
		Script: `
        input["Body"] = JSON.parse(input["Body"]).csv
		`,
	}
	p2, _ := p.Configure(config, core.MagicalGatewayNamespace)

	r, _ := http.NewRequest(http.MethodGet, "/dummy", nil)
	r.Body = io.NopCloser(bytes.NewBufferString("{ \"csv\": \"text\" }"))

	w := gateway.NewDummyWriter()
	p2.ExecutePlugin(nil, w, r)

	assert.Equal(t, "text", w.Body.String())
}

func TestJSOutboundPluginBroken(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(outbound.JSOutboundPluginName)

	config := &outbound.JSOutboundConfig{
		Script: `
		random stuff / 2
		`,
	}
	p2, _ := p.Configure(config, core.MagicalGatewayNamespace)

	r, _ := http.NewRequest(http.MethodGet, "/dummy", nil)

	w := gateway.NewDummyWriter()
	p2.ExecutePlugin(nil, w, r)

	assert.Equal(t, 500, w.Code)
}
