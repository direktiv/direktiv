package gateway

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type variablePlugin struct {
	conf workflowPluginConfig
}

type variablePluginConfig struct {
	Workflow  string `json:"workflow"`
	Namespace string `json:"namespace"`
	Variable  string `json:"variable"`
	Host      string `json:"host"`
	Scheme    string `json:"scheme"`
	UseTLS    bool   `json:"use_tls"`
}

func (e variablePlugin) build(c map[string]interface{}) (serve, error) {
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
		queryParams.Add("op", "var")
		path := fmt.Sprintf("/%s/%s/tree/%s", baseURL, e.conf.Namespace, e.conf.Workflow)
		targetURL := url.URL{
			Host:     e.conf.Host,
			Path:     path,
			RawQuery: queryParams.Encode(),
			Scheme:   e.conf.Scheme,
		}

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

func (e variablePlugin) getSchema() interface{} {
	return &variablePluginConfig{}
}

//nolint:gochecknoinits
func init() {
	registry["target_workflow_var"] = variablePlugin{}
}
