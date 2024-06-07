package target

import (
	"io"
	"net/http"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
)

const debugPluginName = "debug-target"

type DebugPlugin struct{}

var _ core.Plugin = &DebugPlugin{}

func (ba *DebugPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	return &DebugPlugin{}, nil
}

func (ba *DebugPlugin) Execute(w http.ResponseWriter, r *http.Request) *http.Request {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		gateway.WriteInternalError(r, w, err, "reading request body")
		return nil
	}

	response := struct {
		Headers http.Header `json:"headers"`
		Body    string      `json:"body"`
		Text    string      `json:"text"`
	}{
		Headers: r.Header,
		Body:    string(body),
		Text:    "from debug plugin",
	}

	gateway.WriteJSON(w, response)

	return r
}

func (ba *DebugPlugin) Type() string {
	return debugPluginName
}

func init() {
	gateway.RegisterPlugin(&DebugPlugin{})
}
