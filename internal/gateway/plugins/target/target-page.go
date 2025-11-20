package target

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/gateway"
	"gopkg.in/yaml.v3"
)

type PagePlugin struct {
	File            string `mapstructure:"file"`
	pageFileContent string
	namespace       string
}

func (tnf *PagePlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &PagePlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}
	tnf.namespace = config.Namespace

	if pl.File == "" {
		return nil, fmt.Errorf("file is required")
	}

	if !strings.HasPrefix(pl.File, "/") {
		pl.File = "/" + pl.File
	}
	url := fmt.Sprintf("http://localhost:%s/api/v2/namespaces/%s/files%s?raw=true",
		os.Getenv("DIREKTIV_API_PORT"), config.Namespace, pl.File)

	//nolint:gosec,noctx
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
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("couldn't read page file: %w", err)
	}
	pl.pageFileContent = string(data)

	return pl, nil
}

func (tnf *PagePlugin) Type() string {
	return "target-page"
}

func (tnf *PagePlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	if gateway.ExtractContextURLPattern(r) == "" {
		gateway.WriteInternalError(r, w, errors.New("empty extract pattern"), "plugin couldn't parse url")

		return nil, nil
	}
	parts := strings.Split(r.URL.Path, gateway.ExtractContextURLPattern(r))
	if len(parts) != 2 {
		gateway.WriteInternalError(r, w, errors.New("unexpected request url"), "plugin couldn't parse url")

		return nil, nil
	}
	if parts[1] == "" || parts[1] == "index" || parts[1] == "/index.html" {
		http.ServeFile(w, r, "/app/ui/ui-pages.html")

		return w, r
	}

	if parts[1] == "page.json" {
		p := map[string]any{}
		err := yaml.Unmarshal([]byte(tnf.pageFileContent), &p)
		if err != nil {
			gateway.WriteInternalError(r, w, err, "plugin couldn't parse page file")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(p)

		return w, r
	}

	gateway.WriteJSONError(w, http.StatusNotFound, "", "gateway couldn't pages route")

	return nil, nil
}

func init() {
	gateway.RegisterPlugin(&PagePlugin{})
}
