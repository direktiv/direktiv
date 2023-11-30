package endpoints_test

import (
	"net/http"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/endpoints"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/auth"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/inbound"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/outbound"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/target"
	"github.com/stretchr/testify/assert"
)

func TestSetEndpointsFileEmpty(t *testing.T) {
	ep := &endpoints.Endpoint{}

	epl := endpoints.NewEndpointList()
	epl.SetEndpoints([]*endpoints.Endpoint{ep})

	assert.Len(t, epl.Routes(), 0)
}

func TestSetEndpointsWarnings(t *testing.T) {
	ep := &endpoints.Endpoint{
		EndpointBase: &core.EndpointBase{
			Methods:        []string{http.MethodPost},
			AllowAnonymous: true,
			Plugins: core.Plugins{
				Auth: []core.PluginConfig{
					{
						Type: auth.BasicAuthPluginName,
					},
				},
			},
		},
		FilePath: "/route.yaml",
	}

	epl := endpoints.NewEndpointList()
	epl.SetEndpoints([]*endpoints.Endpoint{ep})

	assert.Len(t, epl.Routes(), 1)

	// get route
	r := epl.Routes()[0].Handlers[http.MethodPost]

	// should have a warning
	assert.Len(t, r.Warnings, 1)
}

func TestSetEndpointsErrors(t *testing.T) {
	ep := &endpoints.Endpoint{
		EndpointBase: &core.EndpointBase{
			Methods:        []string{http.MethodPost},
			AllowAnonymous: true,
			Plugins: core.Plugins{
				Auth: []core.PluginConfig{
					{
						Type: auth.BasicAuthPluginName,
						Configuration: map[string]interface{}{
							"add_username_header": 1000,
						},
					},
				},
			},
		},

		FilePath: "/route.yaml",
	}

	epl := endpoints.NewEndpointList()
	epl.SetEndpoints([]*endpoints.Endpoint{ep})

	assert.Len(t, epl.Routes(), 1)

	// get route
	r := epl.Routes()[0].Handlers[http.MethodPost]

	// should have an error
	assert.Len(t, r.Errors, 1)
}

func TestSetEndpoints(t *testing.T) {
	ep := &endpoints.Endpoint{
		EndpointBase: &core.EndpointBase{
			Methods:        []string{http.MethodPost},
			AllowAnonymous: true,
			Plugins: core.Plugins{
				Auth: []core.PluginConfig{
					{
						Type: auth.BasicAuthPluginName,
					},
					{
						Type: auth.KeyAuthPluginName,
						Configuration: map[string]interface{}{
							"key_name": "demo",
						},
					},
				},
				Inbound: []core.PluginConfig{
					{
						Type: inbound.ACLPluginName,
					},
				},
				Target: core.PluginConfig{
					Type: target.InstantResponsePluginName,
					Configuration: map[string]interface{}{
						"status_code":    201,
						"status_message": "demo",
					},
				},
				Outbound: []core.PluginConfig{
					{
						Type: outbound.JSOutboundPluginName,
					},
				},
			},
		},
		FilePath: "/route.yaml",
	}

	epl := endpoints.NewEndpointList()
	epl.SetEndpoints([]*endpoints.Endpoint{ep})

	assert.Len(t, epl.Routes(), 1)

	// // get route
	r := epl.Routes()[0].Handlers[http.MethodPost]

	assert.Len(t, r.AuthPluginInstances, 2)
	c := r.AuthPluginInstances[1].Config().(*auth.KeyAuthConfig)
	assert.Equal(t, "demo", c.KeyName)

	assert.Len(t, r.InboundPluginInstances, 1)

	ci := r.TargetPluginInstance.Config().(*target.InstantResponseConfig)
	assert.Equal(t, 201, ci.StatusCode)
	assert.Equal(t, "demo", ci.StatusMessage)

	assert.Len(t, r.OutboundPluginInstances, 1)
}

func TestSetEndpointsFullError(t *testing.T) {
	ep := &endpoints.Endpoint{
		EndpointBase: &core.EndpointBase{
			Methods:        []string{http.MethodPost},
			AllowAnonymous: true,
			Plugins: core.Plugins{
				Auth: []core.PluginConfig{
					{
						Type: auth.BasicAuthPluginName,
					},
					{
						Type: auth.KeyAuthPluginName,
						Configuration: map[string]interface{}{
							"key_name": "demo",
						},
					},
				},
				Inbound: []core.PluginConfig{
					{
						Type: inbound.ACLPluginName,
					},
				},
				Target: core.PluginConfig{
					Type: target.InstantResponsePluginName,
					Configuration: map[string]interface{}{
						"status_code":    "textnotallowed",
						"status_message": "demo",
					},
				},
				Outbound: []core.PluginConfig{
					{
						Type: outbound.JSOutboundPluginName,
					},
				},
			},
		},
		FilePath: "/route.yaml",
	}

	epl := endpoints.NewEndpointList()
	epl.SetEndpoints([]*endpoints.Endpoint{ep})

	assert.Len(t, epl.Routes(), 1)

	// get route
	r := epl.Routes()[0].Handlers[http.MethodPost]

	// should have an error but route still in there
	assert.Len(t, r.Errors, 1)
}

func TestSetEndpointsFind(t *testing.T) {
	ep := &endpoints.Endpoint{
		EndpointBase: &core.EndpointBase{
			Methods:        []string{http.MethodPost, http.MethodGet},
			AllowAnonymous: true,
			Plugins: core.Plugins{
				Auth: []core.PluginConfig{
					{
						Type: auth.BasicAuthPluginName,
					},
				},
			},
		},
		FilePath: "/route.yaml",
	}

	ep1 := &endpoints.Endpoint{
		EndpointBase: &core.EndpointBase{
			Methods:        []string{http.MethodGet},
			AllowAnonymous: true,
			Plugins: core.Plugins{
				Auth: []core.PluginConfig{
					{
						Type: auth.BasicAuthPluginName,
					},
				},
			},
		},
		FilePath: "/path/to/route.yaml",
	}

	ep2 := &endpoints.Endpoint{
		EndpointBase: &core.EndpointBase{
			Methods:        []string{http.MethodGet},
			AllowAnonymous: true,
			Plugins: core.Plugins{
				Auth: []core.PluginConfig{
					{
						Type: auth.BasicAuthPluginName,
					},
				},
			},
			PathExtension: "/{id}",
		},
		FilePath: "/path/to/route.yaml",
	}

	epl := endpoints.NewEndpointList()
	epl.SetEndpoints([]*endpoints.Endpoint{ep, ep1, ep2})

	assert.Len(t, epl.Routes(), 3)

	r1, _ := epl.FindRoute("/route", http.MethodPost)
	assert.NotNil(t, r1)
	assert.Equal(t, "/route.yaml", r1.FilePath)

	r1, _ = epl.FindRoute("/path/to/route", http.MethodGet)
	assert.NotNil(t, r1)
	assert.Equal(t, "/path/to/route.yaml", r1.FilePath)

	r1, _ = epl.FindRoute("/path/to/route", http.MethodPost)
	assert.Nil(t, r1)

	r2, m2 := epl.FindRoute("/path/to/route/200", http.MethodGet)
	assert.NotNil(t, r2)
	assert.Equal(t, "/path/to/route.yaml", r2.FilePath)

	// path args
	assert.Equal(t, "200", m2["id"])
}

func TestSetEndpointsWrongMethod(t *testing.T) {
	ep := &endpoints.Endpoint{
		EndpointBase: &core.EndpointBase{
			Methods:        []string{http.MethodPost, "DOESNOTEXIST"},
			AllowAnonymous: true,
			Plugins: core.Plugins{
				Auth: []core.PluginConfig{
					{
						Type: auth.BasicAuthPluginName,
					},
				},
			},
		},
		FilePath: "/route.yaml",
	}

	epl := endpoints.NewEndpointList()
	epl.SetEndpoints([]*endpoints.Endpoint{ep})

	// should have only one route for post
	assert.Len(t, epl.Routes()[0].Handlers, 1)
}

func TestSetEndpointsTypeErrors(t *testing.T) {
	ep := &endpoints.Endpoint{
		EndpointBase: &core.EndpointBase{
			Methods:        []string{http.MethodPost},
			AllowAnonymous: true,
			Plugins: core.Plugins{
				Auth: []core.PluginConfig{
					{
						Type: inbound.RequestConvertPluginName,
					},
				},
			},
		},
		FilePath: "/route.yaml",
	}

	epl := endpoints.NewEndpointList()
	epl.SetEndpoints([]*endpoints.Endpoint{ep})

	assert.Len(t, epl.Routes(), 1)

	// get route
	r := epl.Routes()[0].Handlers[http.MethodPost]

	// should have an error because wrong type
	assert.Len(t, r.Errors, 1)
}
