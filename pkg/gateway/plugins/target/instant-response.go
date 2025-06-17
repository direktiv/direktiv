package target

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
)

type InstantResponsePlugin struct {
	StatusCode    int    `mapstructure:"status_code"    yaml:"status_code"`
	StatusMessage string `mapstructure:"status_message" yaml:"status_message"`
	ContentType   string `mapstructure:"content_type"   yaml:"content_type"`
}

func (ir *InstantResponsePlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &InstantResponsePlugin{
		StatusCode:    http.StatusOK,
		StatusMessage: "This is the end!",
	}
	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (ir *InstantResponsePlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	if gateway.IsJSON(ir.StatusMessage) {
		w.Header().Add("Content-Type", "application/json")
	}

	if ir.ContentType != "" {
		w.Header().Set("Content-Type", ir.ContentType)
	}

	w.WriteHeader(ir.StatusCode)
	// nolint
	w.Write([]byte(ir.StatusMessage))

	return w, r
}

func (ir *InstantResponsePlugin) Type() string {
	return "instant-response"
}

func init() {
	gateway.RegisterPlugin(&InstantResponsePlugin{})
}
