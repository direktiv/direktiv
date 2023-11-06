package gateway

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type workflowPlugin struct {
	conf workflowPluginConfig
}

type workflowPluginConfig struct {
	Workflow  string `json:"workflow"`
	Namespace string `json:"namespace"`
	Host      string `json:"host"`
	Scheme    string `json:"scheme"`
	UseTLS    bool   `json:"use_tls"`
}

func (e workflowPlugin) build(c map[string]interface{}) (serve, error) {
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
		queryParams.Add("op", "wait")
		queryParams.Add("ref", "latest")

		path := fmt.Sprintf("/%s/%s/tree/%s", baseURL, e.conf.Namespace, e.conf.Workflow)
		targetURL := url.URL{
			Host:     e.conf.Host,
			Path:     path,
			RawQuery: queryParams.Encode(),
			Scheme:   e.conf.Scheme,
		}

		fmt.Printf("target url:%v\n", targetURL)

		proxy := httputil.NewSingleHostReverseProxy(&targetURL)
		proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !e.conf.UseTLS}, //nolint:gosec
		}

		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.URL.Path = targetURL.Path
			req.Host = targetURL.Host
			req.URL.RawQuery = queryParams.Encode()
		}

		proxy.ServeHTTP(w, r)

		return false
	}, nil
}

func (e workflowPlugin) getSchema() interface{} {
	return &e.conf
}

//nolint:gochecknoinits
func init() {
	registry["target_workflow"] = workflowPlugin{}
}
