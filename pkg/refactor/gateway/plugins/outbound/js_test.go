package outbound_test

import (
	"bytes"
	"context"
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
		input["Headers"].Add("jens", "hallo2")
		input["Code"] = 204
		`,
	}
	p2, _ := p.Configure(config, core.MagicalGatewayNamespace)

	r, _ := http.NewRequest(http.MethodGet, "/dummy", nil)
	r.Header.Add("header1", "value1")
	r.Header.Add("header2", "value2")
	r.Body = io.NopCloser(bytes.NewBufferString("{ \"string1\": \"value2\" }"))

	w := gateway.NewDummyWriter()
	p2.ExecutePlugin(context.Background(), nil, w, r)

	assert.Equal(t, 204, w.Code)

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
	p2.ExecutePlugin(context.Background(), nil, w, r)

	assert.Equal(t, 500, w.Code)
}
