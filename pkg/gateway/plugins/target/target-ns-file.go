package target

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
)

// NamespaceFilePlugin returns a files in the explorer tree.
type NamespaceFilePlugin struct {
	Namespace   string `mapstructure:"namespace"`
	File        string `mapstructure:"file"`
	ContentType string `mapstructure:"content_type"`
}

func (tnf *NamespaceFilePlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &NamespaceFilePlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	if pl.File == "" {
		return nil, fmt.Errorf("file is required")
	}

	if !strings.HasPrefix(pl.File, "/") {
		pl.File = "/" + pl.File
	}

	return pl, nil
}

func (tnf *NamespaceFilePlugin) Type() string {
	return "target-namespace-file"
}

func (tnf *NamespaceFilePlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	currentNS := gateway.ExtractContextEndpoint(r).Namespace
	if tnf.Namespace == "" {
		tnf.Namespace = currentNS
	}
	if tnf.Namespace != currentNS && currentNS != core.SystemNamespace {
		gateway.WriteForbiddenError(r, w, nil, "plugin can not target different namespace")
		return nil, nil
	}

	url := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/files%s?raw=true",
		os.Getenv("DIREKTIV_API_PORT"), tnf.Namespace, tnf.File)

	// request failed if nil and response already written
	resp, err := doRequest(r, http.MethodGet, url, nil)
	if err != nil {
		gateway.WriteInternalError(r, w, err, "couldn't execute downstream request")
		return nil, nil
	}
	defer resp.Body.Close()

	// copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	if tnf.ContentType != "" {
		w.Header().Set("Content-Type", tnf.ContentType)
	}
	// copy the status code
	w.WriteHeader(resp.StatusCode)

	// copy the response body
	if _, err := io.Copy(w, resp.Body); err != nil {
		gateway.WriteInternalError(r, w, nil, "couldn't write downstream response")
		return nil, nil
	}

	return w, r
}

func init() {
	gateway.RegisterPlugin(&NamespaceFilePlugin{})
}
