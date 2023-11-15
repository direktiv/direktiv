package gateway

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type nsVariablePlugin struct {
	conf nsVariablePluginConfig
}

type nsVariablePluginConfig struct {
	Namespace string `json:"namespace" jsonschema:"required"`
	Variable  string `json:"variable"  jsonschema:"required"`
	Host      string `json:"host"`
	Scheme    string `json:"scheme"`
	UseTLS    bool   `json:"use_tls"`
}

func (e nsVariablePlugin) build(c map[string]interface{}) (serve, error) {
	if err := unmarshalConfig(c, &e.conf); err != nil {
		return nil, err
	}
	if e.conf.Host == "" {
		e.conf.Host = "localhost:6665"
	}

	if e.conf.Scheme == "" {
		e.conf.Scheme = "http"
	}

	return func(w http.ResponseWriter, r *http.Request) bool {
		baseURL := "api/namespaces"
		queryParams := url.Values{}
		path := fmt.Sprintf("/%s/%s/vars/%s", baseURL, e.conf.Namespace, e.conf.Variable)
		targetURL := url.URL{
			Host:     e.conf.Host,
			Path:     path,
			RawQuery: queryParams.Encode(),
			Scheme:   e.conf.Scheme,
		}
		slog.Error(targetURL.String())
		proxy := httputil.NewSingleHostReverseProxy(&targetURL)
		proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !e.conf.UseTLS}, //nolint:gosec
		}

		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.URL.Path = targetURL.Path
			req.Host = targetURL.Host
		}

		proxy.ServeHTTP(w, r)

		return false
	}, nil
}

func (e nsVariablePlugin) getSchema() interface{} {
	return &nsVariablePluginConfig{}
}

//nolint:gochecknoinits
func init() {
	registry["target_namespace_var"] = nsVariablePlugin{}
}
