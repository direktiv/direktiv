package inbound_test

import (
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
		HeadersToAdd: []inbound.NameKeys{
			{
				Name:  "header3",
				Value: "value3",
			},
		},
		HeadersToModify: []inbound.NameKeys{
			{
				Name:  "header1",
				Value: "newvalue",
			},
		},
		HeadersToRemove: []inbound.NameKeys{
			{
				Name: "Header2",
			},
		},
	}
	p2, _ := p.Configure(config, core.SystemNamespace)

	r, _ := http.NewRequest(http.MethodGet, "/dummy", nil)
	r.Header.Add("Header1", "value1")
	r.Header.Add("header2", "value2")

	w := httptest.NewRecorder()
	p2.ExecutePlugin(nil, w, r)

	assert.Empty(t, r.Header.Get("header2"))
	assert.Equal(t, "value3", r.Header.Get("Header3"))
	assert.Equal(t, "newvalue", r.Header.Get("header1"))
}
