package gateway

import (
	"net/http"
)

type tokenAuthenticationPlugin struct {
	conf tokenAuthPluginConfig
}

type tokenAuthPluginConfig struct {
	TokenValue string `json:"token_value"`
}

func (p tokenAuthenticationPlugin) build(c map[string]interface{}) (serve, error) {
	if err := unmarshalConfig(c, &p.conf); err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, r *http.Request) bool {
		token := r.Header.Get("direktiv-token")

		if token == p.conf.TokenValue {
			return true
		}
		w.WriteHeader(http.StatusUnauthorized)

		return false
	}, nil
}

func (p tokenAuthenticationPlugin) getSchema() interface{} {
	return &tokenAuthPluginConfig{}
}

//nolint:gochecknoinits
func init() {
	registry["token_auth_plugin"] = tokenAuthenticationPlugin{}
}
