package target

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/gateway"
	"github.com/direktiv/direktiv/pkg/filestore"
)

type NamespaceFileServerPlugin struct {
	AllowPaths []string `mapstructure:"allow_paths"`
	DenyPaths  []string `mapstructure:"deny_paths"`
	Namespace  string   `mapstructure:"namespace"`
}

func (tnf *NamespaceFileServerPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &NamespaceFileServerPlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}
	for _, path := range pl.AllowPaths {
		_, err = filestore.ValidatePath(path)
		if err != nil {
			return nil, err
		}
		if path == "" {
			return nil, errors.New("allow_paths has an empty path")
		}
	}
	for _, path := range pl.DenyPaths {
		_, err = filestore.ValidatePath(path)
		if err != nil {
			return nil, err
		}
		if path == "" {
			return nil, errors.New("deny_paths has an empty path")
		}
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

	for _, path := range tnf.DenyPaths {
		if strings.HasPrefix("/"+parts[1], path) {
			gateway.WriteInternalError(r, w, errors.New("unexpected request url"), "path is denied")

			return nil, nil
		}
	}
	allowed := false
	for _, path := range tnf.AllowPaths {
		if strings.HasPrefix("/"+parts[1], path) {
			allowed = true
		}
	}
	if !allowed {
		gateway.WriteInternalError(r, w, errors.New("unexpected request url"), "path is not allowed")
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
