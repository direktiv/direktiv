package target

import (
	"fmt"
	"io"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
)

const debugPluginName = "debug-target"

type DebugPlugin struct{}

var _ core.PluginV2 = &DebugPlugin{}

func (ba *DebugPlugin) Construct(config core.PluginConfigV2) (core.PluginV2, error) {
	return &DebugPlugin{}, nil
}

func (ba *DebugPlugin) Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("reading request body: %w", err)
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

	gateway2.WriteJSON(w, response)

	return r, nil
}

func (ba *DebugPlugin) Type() string {
	return debugPluginName
}

func (ba *DebugPlugin) Config() interface{} {
	return nil
}

func init() {
	plugins.RegisterPlugin(&DebugPlugin{})
}
