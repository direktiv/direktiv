package target

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
)

type NamespaceFileServerPlugin struct {
	Dir       string `mapstructure:"dir"`
	Namespace string `mapstructure:"namespace"`
}

func (tnf *NamespaceFileServerPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &NamespaceFileServerPlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (tnf *NamespaceFileServerPlugin) Type() string {
	return "target-fileserver"
}

func (tnf *NamespaceFileServerPlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	currentNS := gateway.ExtractContextEndpoint(r).Namespace
	if tnf.Namespace == "" {
		tnf.Namespace = currentNS
	}
	if tnf.Namespace != currentNS && currentNS != core.SystemNamespace {
		gateway.WriteForbiddenError(r, w, nil, "plugin can not target different namespace")

		return nil, nil
	}
	if gateway.ExtractContextURLPattern(r) == "" {
		gateway.WriteInternalError(r, w, errors.New("empty extract pattern"), "plugin couldn't parse url")

		return nil, nil
	}
	parts := strings.Split(r.URL.Path, gateway.ExtractContextURLPattern(r))
	if len(parts) != 2 {
		gateway.WriteInternalError(r, w, errors.New("unexpected request url"), "plugin couldn't parse url")

		return nil, nil
	}

	url := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/files%s?raw=true",
		os.Getenv("DIREKTIV_API_PORT"), tnf.Namespace, "/"+parts[1])
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
	gateway.RegisterPlugin(&NamespaceFileServerPlugin{})
}
