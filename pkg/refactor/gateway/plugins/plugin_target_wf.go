package plugins

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"golang.org/x/exp/slog"
)

type executeWF struct {
	conf  targetWFSpec
	proxy *httputil.ReverseProxy
}

func (e *executeWF) buildPlugin(conf interface{}) (Execute, error) {
	cfg, ok := conf.(targetWFSpec)
	if !ok {
		return nil, fmt.Errorf("invalid configuration type")
	}
	e.conf = cfg

	targetURL := url.URL{
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Scheme: cfg.Scheme,
	}

	e.proxy = httputil.NewSingleHostReverseProxy(&targetURL)
	e.proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.Insecure}, //nolint:gosec
	}
	e.proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host

		req.Header.Set("X-Forwarded-For", req.RemoteAddr)

		if req.TLS != nil {
			req.Header.Set("X-Forwarded-Proto", "https")
		} else {
			req.Header.Set("X-Forwarded-Proto", "http")
		}
	}

	return e.safeProcess, nil
}

func (e *executeWF) safeProcess(w http.ResponseWriter, r *http.Request) Result {
	proxyURL := fmt.Sprintf("/api/namespaces/%s/tree/%s?op=execute&ref=%s", e.conf.Namespace, e.conf.Workflow, e.conf.Revision)
	slog.Debug("Proxing to:", proxyURL)

	// Update the path for this specific request.
	r.URL.Path = proxyURL

	e.proxy.ServeHTTP(w, r)

	slog.Debug("Executed")

	return Result{Status: http.StatusOK}
}

//nolint:gochecknoinits
func init() {
	register(formPluginKey("v1", "execute_workflow"), &executeWF{})
}

type targetWFSpec struct {
	Revision  string `json:"revision"`
	Namespace string `json:"namespace"`
	Workflow  string `json:"workflow"`
	Host      string `json:"host"`
	Port      string `json:"port"`
	Insecure  bool   `json:"insecure"`
	Scheme    string `json:"scheme"`
}

func (e *executeWF) GetConfigStruct() interface{} {
	return targetWFSpec{}
}
