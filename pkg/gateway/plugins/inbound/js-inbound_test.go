package inbound_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/gateway/plugins/inbound"
	"github.com/stretchr/testify/assert"
)

func TestExecuteJSInboundPlugin(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(inbound.JSInboundPluginName)

	config := &inbound.JSInboundConfig{
		Script: `
		input["Headers"].Del("Header1")
		input["Headers"].Add("Header3", "value3")

		input["Queries"].Del("Query1")
		input["Queries"].Add("Query3", "value3")

		b = JSON.parse(input["Body"])
		b["addquery"] = input["Queries"].Get("Query3")
		b["addheader"] = input["Headers"].Get("Header3")
		input["Body"] = JSON.stringify(b) 
		`,
	}
	p2, _ := p.Configure(config, core.SystemNamespace)

	r, _ := http.NewRequest(http.MethodGet, "/dummy?Query1=value1&Query2=value2", nil)
	r.Header.Add("Header1", "value1")
	r.Header.Add("Header2", "value2")
	r.Body = io.NopCloser(bytes.NewBufferString("{ \"string1\": \"value2\" }"))

	w := httptest.NewRecorder()
	p2.ExecutePlugin(nil, w, r)

	// got deleted
	assert.Empty(t, r.Header.Get("Header1"))
	assert.Empty(t, r.URL.Query().Get("Query1"))

	// newly set
	assert.Equal(t, "value3", r.Header.Get("Header3"))
	assert.Equal(t, "value3", r.URL.Query().Get("Query3"))

	// still available
	assert.NotEmpty(t, r.Header.Get("Header2"))
	assert.NotEmpty(t, r.URL.Query().Get("Query2"))

	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	assert.JSONEq(t, "{\"string1\":\"value2\",\"addheader\":\"value3\", \"addquery\":\"value3\"}", string(b))
}

func TestExecuteJSInboundPluginConsumer(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(inbound.JSInboundPluginName)
	config := &inbound.JSInboundConfig{
		Script: `
		b = JSON.parse(input["Body"])
	    b["user"] = input["Consumer"].Username
		input["Body"] = JSON.stringify(b) 
		`,
	}
	p2, _ := p.Configure(config, core.SystemNamespace)

	r, _ := http.NewRequest(http.MethodGet, "/dummy", nil)
	r.Body = io.NopCloser(bytes.NewBufferString("{ }"))

	w := httptest.NewRecorder()

	p2.ExecutePlugin(&core.ConsumerFile{
		Username: "test",
	}, w, r)

	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	assert.JSONEq(t, "{\"user\":\"test\"}", string(b))
}

func TestExecuteJSInboundPluginURLParam(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(inbound.JSInboundPluginName)
	config := &inbound.JSInboundConfig{
		Script: `
		b = JSON.parse(input["Body"])
	    b["params"] = input["URLParams"].id
		input["Body"] = JSON.stringify(b) 
		`,
	}
	p2, _ := p.Configure(config, core.SystemNamespace)

	urlParams := map[string]string{
		"id": "123",
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, plugins.URLParamCtxKey, urlParams)

	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/dummy/thisismyid", nil)
	r.Body = io.NopCloser(bytes.NewBufferString("{ }"))

	w := httptest.NewRecorder()

	p2.ExecutePlugin(nil, w, r)

	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	assert.JSONEq(t, "{\"params\":\"123\"}", string(b))
}

func TestExecuteJSInboundPluginStatus(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(inbound.JSInboundPluginName)
	config := &inbound.JSInboundConfig{
		Script: `
		b = JSON.parse(input["Body"])
	    b["error"] = "no access" 
		input["Body"] = JSON.stringify(b) 
		input["Headers"].Add("permission", "denied")
		input.Status = 403
		`,
	}
	p2, _ := p.Configure(config, core.SystemNamespace)

	r, _ := http.NewRequest(http.MethodGet, "/dummy", nil)
	r.Body = io.NopCloser(bytes.NewBufferString("{ }"))

	w := httptest.NewRecorder()

	p2.ExecutePlugin(nil, w, r)

	assert.Equal(t, "denied", r.Header.Get("permission"))
	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)

	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	assert.JSONEq(t, "{\"error\":\"no access\"}", string(b))
}
