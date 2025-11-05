package target

import (
	"fmt"
	"io"
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
	url := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/files%s?raw=true",
		os.Getenv("DIREKTIV_API_PORT"), config.Namespace, pl.File)

	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("direktiv api error: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("page file not found")
	}
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("couldn't fetch page file: %s", body)
	}

	return pl, nil
}

func (tnf *PagePlugin) Type() string {
	return "target-page"
}

func (tnf *PagePlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello world from Pages!"))

	return w, r
}

func init() {
	gateway.RegisterPlugin(&PagePlugin{})
}
