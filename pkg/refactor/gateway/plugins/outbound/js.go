package outbound

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/dop251/goja"
)

const (
	JSOutboundPluginName = "js-outbound"
)

type JSOutboundConfig struct {
	Script string `mapstructure:"script" yaml:"script"`
}

type JSOutboundPlugin struct {
	config *JSOutboundConfig
}

func ConfigureKeyAuthPlugin(config interface{}, _ string) (plugins.PluginInstance, error) {
	jsConfig := &JSOutboundConfig{}

	err := plugins.ConvertConfig(config, jsConfig)
	if err != nil {
		return nil, err
	}

	return &JSOutboundPlugin{
		config: jsConfig,
	}, nil
}

type response struct {
	// Headers http.Header
	Headers Headers
	Body    string
	Code    int
}

type Headers struct {
	H http.Header
}

func (h Headers) Get(key string) []string {
	return h.H[key]
}

func (h Headers) Set(key string, value string) {
	h.H.Set(key, value)
}

func (h Headers) Add(key string, value string) {
	h.H.Add(key, value)
}

func (h Headers) Delete(key string) {
	h.H.Del(key)
}

func (js *JSOutboundPlugin) ExecutePlugin(_ *core.ConsumerBase,
	w http.ResponseWriter, r *http.Request,
) bool {
	var (
		err error
		b   []byte
	)

	if r.Body != nil {
		b, err = io.ReadAll(r.Body)
		if err != nil {
			plugins.ReportError(w, http.StatusInternalServerError,
				"can not set read body for js plugin", err)

			return false
		}
		defer r.Body.Close()
	}

	if r.Response == nil {
		r.Response = &http.Response{
			StatusCode: http.StatusOK,
		}
	}

	resp := response{
		Headers: Headers{
			H: r.Header,
		},
		Body: string(b),
		Code: r.Response.StatusCode,
	}

	// extract all response headers and body
	vm := goja.New()
	err = vm.Set("input", resp)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not set input object", err)

		return false
	}

	err = vm.Set("log", func(txt interface{}) {
		slog.Info("js log", slog.Any("log", txt))
	})
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not set log function", err)

		return false
	}

	err = vm.Set("sleep", func(t interface{}) {
		tt, ok := t.(int64)
		if !ok {
			return
		}
		time.Sleep(time.Duration(tt) * time.Second)
	})
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not set sleep function", err)

		return false
	}

	script := fmt.Sprintf("function run() { %s; return input } run()",
		js.config.Script)

	val, err := vm.RunScript("plugin", script)
	if err != nil {
		plugins.ReportError(w, http.StatusInternalServerError,
			"can not execute script", err)

		return false
	}

	if val != nil && !val.Equals(goja.Undefined()) {
		o := val.ToObject(vm)
		// make sure the input object got returned
		if o.ExportType() == reflect.TypeOf(resp) {
			// nolint checked before
			responseDone := o.Export().(response)
			for k, v := range responseDone.Headers.H {
				for a := range v {
					w.Header().Add(k, v[a])
				}
			}

			// nolint
			w.Write([]byte(responseDone.Body))
			w.WriteHeader(responseDone.Code)
		}
	}

	return true
}

func (js *JSOutboundPlugin) Config() interface{} {
	return js.config
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		JSOutboundPluginName,
		plugins.OutboundPluginType,
		ConfigureKeyAuthPlugin))
}
