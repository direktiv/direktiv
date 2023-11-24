package outbound

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
	"github.com/dop251/goja"
)

const (
	JSOutboundPluginName = "js-outbound"
)

type JSOutboundConfig struct {
	Script string `yaml:"script" mapstructure:"script"`
}

type JSOutboundPlugin struct {
	config *JSOutboundConfig
}

func ConfigureKeyAuthPlugin(config interface{}, ns string) (plugins.PluginInstance, error) {
	jsConfig := &JSOutboundConfig{}

	err := plugins.ConvertConfig(JSOutboundPluginName, config, jsConfig)
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
	h http.Header
}

func (h Headers) Get(key string) []string {
	return h.h[key]
}

func (h Headers) Set(key string, value string) {
	h.h.Set(key, value)
}

func (h Headers) Add(key string, value string) {
	h.h.Add(key, value)
}

func (h Headers) Delete(key string) {
	h.h.Del(key)
}

func (js *JSOutboundPlugin) ExecutePlugin(ctx context.Context, c *spec.ConsumerFile,
	w http.ResponseWriter, r *http.Request) bool {

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
			h: r.Header,
		},
		Body: string(b),
		Code: r.Response.StatusCode,
	}

	fmt.Println("COMING IN ")
	fmt.Println(resp)

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
			responseDone := o.Export().(response)
			for k, v := range responseDone.Headers.h {
				for a := range v {
					w.Header().Add(k, v[a])
				}
			}
			w.Write([]byte(responseDone.Body))
			w.WriteHeader(responseDone.Code)
		}
	}

	return true
}

func (js *JSOutboundPlugin) Config() interface{} {
	return js.config
}

func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		JSOutboundPluginName,
		plugins.OutboundPluginType,
		ConfigureKeyAuthPlugin))
}
