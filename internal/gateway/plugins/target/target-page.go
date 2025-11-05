package target

import (
	"fmt"
	"net/http"
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
	w.Write([]byte("page plugin works"))
	w.WriteHeader(http.StatusOK)

	return w, r
}

func init() {
	gateway.RegisterPlugin(&PagePlugin{})
}
