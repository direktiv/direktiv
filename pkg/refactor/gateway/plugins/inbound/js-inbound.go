package inbound

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/dop251/goja"
)

const (
	JSInboundPluginName = "js-inbound"
)

type JSInboundConfig struct {
	Script string `mapstructure:"script" yaml:"script"`
}

// JSInboundPlugin allows to modify headers, contents and query params of the request.
type JSInboundPlugin struct {
	config *JSInboundConfig
}

func ConfigureJSInbound(config interface{}, _ string) (core.PluginInstance, error) {
	jsConfig := &JSInboundConfig{}

	err := plugins.ConvertConfig(config, jsConfig)
	if err != nil {
		return nil, err
	}

	return &JSInboundPlugin{
		config: jsConfig,
	}, nil
}

func (js *JSInboundPlugin) Config() interface{} {
	return js.config
}

func (js *JSInboundPlugin) Type() string {
	return JSInboundPluginName
}

type Query struct {
	U url.Values
}

func (q Query) Get(key string) []string {
	return q.U[key]
}

func (q Query) Set(key string, value string) {
	q.U.Set(key, value)
}

func (q Query) Add(key string, value string) {
	q.U.Add(key, value)
}

func (q Query) Delete(key string) {
	q.U.Del(key)
}

type Values struct {
	Something map[string][]string
}

type request struct {
	Headers http.Header
	// Headers shared.Headers
	Queries url.Values
	// Queries shared.Query
	Body string
}

func (js *JSInboundPlugin) ExecutePlugin(_ *core.ConsumerFile,
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
				"can not set read body for js inbound plugin", err)

			return false
		}
		defer r.Body.Close()
	}

	req := request{
		Headers: r.Header,
		Queries: r.URL.Query(),
		Body:    string(b),
	}

	// extract all response headers and body
	vm := goja.New()
	err = vm.Set("input", req)
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
		r.Header = http.Header{}

		o := val.ToObject(vm)
		// make sure the input object got returned
		if o.ExportType() == reflect.TypeOf(req) {
			// nolint checked before
			responseDone := o.Export().(request)
			for k, v := range responseDone.Headers {
				for a := range v {
					r.Header.Add(k, v[a])
				}
			}

			newQuery := make(url.Values)
			for k, v := range responseDone.Queries {
				for a := range v {
					newQuery.Add(k, v[a])
				}
			}
			r.URL.RawQuery = newQuery.Encode()
			r.Body = io.NopCloser(strings.NewReader(responseDone.Body))
		}
	}

	return true
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		JSInboundPluginName,
		plugins.InboundPluginType,
		ConfigureJSInbound))
}
