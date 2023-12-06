package inbound_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/inbound"
	"github.com/stretchr/testify/assert"
)

func TestExecuteHeaderManipulationPlugin(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(inbound.HeaderManipulation)

	config := &inbound.HeaderManipulationConfig{
		HeadersToAdd: map[string]string{
			"test/header": "added",
		},
		HeadersToModify: map[string]string{
			"header1": "value was changed",
		},
		HeadersToRemove: []string{
			"header2",
		},
	}
	p2, _ := p.Configure(config, core.MagicalGatewayNamespace)

	r, _ := http.NewRequest(http.MethodGet, "/dummy?test=me", nil)
	r.Header.Add("Header1", "value1")
	r.Header.Add("header2", "value2")

	w := httptest.NewRecorder()
	p2.ExecutePlugin(nil, w, r)

	fmt.Println(r.URL)
	assert.Equal(t, "value was changed", r.Header.Get("header1"))
	assert.Equal(t, "added", r.Header.Get("test/header"))
	assert.Equal(t, "", r.Header.Get("header2"))
}
