package core_test

import (
	"reflect"
	"testing"

	"github.com/direktiv/direktiv/pkg/core"
)

func TestParseOpenAPIPathFile(t *testing.T) {
	tests := []struct {
		name     string
		ns       string
		filePath string
		data     []byte
		expected core.Endpoint
	}{
		{
			name:     "Valid input with all fields",
			ns:       "namespace1",
			filePath: "path1.yaml",
			data: []byte(`
get:
    summary: "Fetch endpoint details"
    description: "Retrieves example details."
    responses:
        "200":
            description: "Successful response"
x-extensions:
    direktiv: "api_path/v1"
    path: "endpoint1"
    allow-anonymous: true
    timeout: 30
    plugins:
        auth: []
        inbound: []
        target:
            type: "instant-response"
            configuration:
                status_code: 201
                status_message: "TEST1"
        outbound: []
`),
			expected: core.Endpoint{
				Namespace: "namespace1",
				FilePath:  "path1.yaml",
				EndpointFile: core.EndpointFile{
					DirektivAPI: "api_path/v1",
					Path:        "/endpoint1",
					PluginsConfig: core.PluginsConfig{
						Auth: []core.PluginConfig{},
						Target: core.PluginConfig{
							Typ: "instant-response",
							Config: map[string]any{
								"status_code":    201,
								"status_message": "TEST1",
							},
						},
					},
					AllowAnonymous: true,
					Methods:        []string{"GET"},
					Timeout:        30,
				},
				Errors: []string{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := core.ParseOpenAPIPathFile(test.ns, test.filePath, test.data)
			if !areEndpointsEqual(result, test.expected) {
				t.Errorf("Test %s failed.\nExpected: %+v\nGot: %+v", test.name, test.expected, result)
			}
		})
	}
}

func areEndpointsEqual(a, b core.Endpoint) bool {
	if a.Namespace != b.Namespace || a.FilePath != b.FilePath {
		return false
	}
	if len(a.Errors) != len(b.Errors) {
		return false
	}
	for i := range a.Errors {
		if a.Errors[i] != b.Errors[i] {
			return false
		}
	}
	if a.DirektivAPI != b.DirektivAPI {
		return false
	}
	if len(a.Methods) != len(b.Methods) {
		return false
	}
	for i := range a.Methods {
		if a.Methods[i] != b.Methods[i] {
			return false
		}
	}
	if a.Path != b.Path {
		return false
	}
	if a.AllowAnonymous != b.AllowAnonymous {
		return false
	}
	if !arePluginsConfigsEqual(a.PluginsConfig, b.PluginsConfig) {
		return false
	}
	if a.Timeout != b.Timeout {
		return false
	}

	return true
}

func arePluginsConfigsEqual(a, b core.PluginsConfig) bool {
	if len(a.Auth) != len(b.Auth) {
		return false
	}
	for i := range a.Auth {
		if !reflect.DeepEqual(a.Auth[i], b.Auth[i]) {
			return false
		}
	}
	if a.Target.Typ != b.Target.Typ {
		return false
	}
	if !reflect.DeepEqual(a.Target.Config, b.Target.Config) {
		return false
	}
	if len(a.Inbound) != len(b.Inbound) {
		return false
	}
	for i := range a.Inbound {
		if !reflect.DeepEqual(a.Inbound[i], b.Inbound[i]) {
			return false
		}
	}
	if len(a.Outbound) != len(b.Outbound) {
		return false
	}
	for i := range a.Outbound {
		if !reflect.DeepEqual(a.Outbound[i], b.Outbound[i]) {
			return false
		}
	}

	return true
}
