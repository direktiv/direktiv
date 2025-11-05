package target

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/gateway"
)

type PagePlugin struct {
	File string `mapstructure:"file"`
}

func (tnf *PagePlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &PagePlugin{}

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

func (tnf *PagePlugin) Type() string {
	return "target-page"
}

func (tnf *PagePlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	currentNS := gateway.ExtractContextEndpoint(r).Namespace

	url := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/files%s?raw=true",
		os.Getenv("DIREKTIV_API_PORT"), currentNS, tnf.File)

	resp, err := doRequest(r, http.MethodGet, url, nil)
	if err != nil {
		gateway.WriteInternalError(r, w, err, "couldn't execute downstream request")
		return nil, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		gateway.WriteInternalError(r, w, nil, "page file not found")
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		gateway.WriteInternalError(r, w, nil, fmt.Sprintf("could not fetch page file: statusCode:%s", resp.Status))
		return nil, nil
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello world from Pages!"))

	return w, r
}

func init() {
	gateway.RegisterPlugin(&PagePlugin{})
}
