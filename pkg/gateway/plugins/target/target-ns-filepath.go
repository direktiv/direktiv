package target

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
)

// NamespaceFilepathPlugin returns files in the explorer tree.
type NamespaceFilepathPlugin struct {
	Namespace string `mapstructure:"namespace"`
	Filepath  string `mapstructure:"filepath"`
}

func (tnf *NamespaceFilepathPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &NamespaceFilepathPlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	if pl.Filepath == "" {
		return nil, fmt.Errorf("filepath is required")
	}

	if !strings.HasPrefix(pl.Filepath, "/") {
		pl.Filepath = "/" + pl.Filepath
	}

	return pl, nil
}

func (tnf *NamespaceFilepathPlugin) Type() string {
	return "target-namespace-filepath"
}

func (tnf *NamespaceFilepathPlugin) Execute(w http.ResponseWriter, r *http.Request) *http.Request {
	currentNS := gateway.ExtractContextEndpoint(r).Namespace
	if tnf.Namespace == "" {
		tnf.Namespace = currentNS
	}
	if tnf.Namespace != currentNS && currentNS != core.SystemNamespace {
		gateway.WriteForbiddenError(r, w, nil, "plugin can not target different namespace")
		return nil
	}

	pattern := gateway.ExtractContextRouterPattern(r)
	relativePath := strings.TrimPrefix(r.URL.Path, pattern)
	target := tnf.Filepath
	if relativePath != "" {
		target = filepath.Join(target, relativePath)
	}

	url := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/files%s?raw=true",
		os.Getenv("DIREKTIV_API_PORT"), tnf.Namespace, target)

	// request failed if nil and response already written
	resp, err := doRequest(r, http.MethodGet, url, nil)
	if err != nil {
		gateway.WriteInternalError(r, w, err, "couldn't execute downstream request")

		return nil
	}
	defer resp.Body.Close()

	// copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// copy the status code
	w.WriteHeader(resp.StatusCode)

	// copy the response body
	if _, err := io.Copy(w, resp.Body); err != nil {
		gateway.WriteInternalError(r, w, nil, "couldn't write downstream response")

		return nil
	}

	return r
}

func init() {
	gateway.RegisterPlugin(&NamespaceFilepathPlugin{})
}
