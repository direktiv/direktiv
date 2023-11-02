package gateway

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/dop251/goja"
)

type jsExecutionPlugin struct {
	conf jsExecutionPluginConfig
}

type jsExecutionPluginConfig struct {
	Script string `json:"script"`
}

func (p *jsExecutionPlugin) build(c map[string]interface{}) (serve, error) {
	if err := unmarshalConfig(c, &p.conf); err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, r *http.Request) bool {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading request body: %s", err), http.StatusInternalServerError)

			return false
		}
		defer r.Body.Close()

		vm := goja.New()
		err = vm.Set("body", string(bodyBytes)) // TODO: add metrics here
		if err != nil {
			slog.Info("error setting body", "error", err)

			return false
		}
		scriptWrapper := fmt.Sprintf(`function transform() { %s } transform();`, p.conf.Script)
		_, err = vm.RunString(scriptWrapper)
		if err != nil {
			http.Error(w, fmt.Sprintf("Script execution error: %s", err), http.StatusInternalServerError)
			slog.Info("Script execution error", "error", err) // TODO: add metrics here

			return false
		}

		modifiedBody := vm.Get("body").String()

		buffer := bytes.NewBufferString(modifiedBody)
		r.Body = io.NopCloser(buffer)
		r.ContentLength = int64(buffer.Len()) // Set the correct Content-Length

		r.Header.Set("Content-Length", strconv.Itoa(buffer.Len()))

		slog.Info("changed body", "body", modifiedBody)

		return true
	}, nil
}

func (p *jsExecutionPlugin) getSchema() interface{} {
	return jsExecutionPluginConfig{}
}

//nolint:gochecknoinits
func init() {
	// Register the plugin in the registry
	registry["js_execution_plugin"] = &jsExecutionPlugin{}
}
