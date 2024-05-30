package inbound

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway/plugins"
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

	Consumer *core.ConsumerFile

	// url params of type /{id}
	URLParams map[string]string

	// response code after executing
	Status int
}

func (js *JSInboundPlugin) ExecutePlugin(c *core.ConsumerFile,
	w http.ResponseWriter, r *http.Request,
) bool {
	var (
		err error
		b   []byte
	)

	if r.Body != nil {
		b, err = io.ReadAll(r.Body)
		if err != nil {
			plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
				"can not set read body for js inbound plugin", err)

			return false
		}
		defer r.Body.Close()
	}

	vm := goja.New()

	// add consumer
	if c == nil {
		c = &core.ConsumerFile{}
	}

	// add url param
	urlParams := make(map[string]string)

	up := r.Context().Value(plugins.URLParamCtxKey)
	if up != nil {
		// nolint we know it is from us
		urlParams = up.(map[string]string)
	}

	req := request{
		Headers:   r.Header,
		Queries:   r.URL.Query(),
		Body:      string(b),
		Consumer:  c,
		URLParams: urlParams,
		Status:    0,
	}

	// extract all response headers and body
	err = vm.Set("input", req)
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"can not set input object", err)

		return false
	}

	err = vm.Set("log", func(txt interface{}) {
		slog.Info("js log", slog.Any("log", txt))
	})
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"can not set log function", err)

		return false
	}

	script := fmt.Sprintf("function run() { %s; return input } run()",
		js.config.Script)

	val, err := vm.RunScript("plugin", script)
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
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
			addHeader(responseDone.Headers, r.Header)

			newQuery := make(url.Values)
			for k, v := range responseDone.Queries {
				for a := range v {
					newQuery.Add(k, v[a])
				}
			}
			r.URL.RawQuery = newQuery.Encode()
			r.Body = io.NopCloser(strings.NewReader(responseDone.Body))

			// script set status code and stop executing other plugins
			if responseDone.Status > 0 {
				return serveResponse(w, responseDone)
			}
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

// serveResponse is writing the response directly to the client if the a status
// code is set within the Javascript.
func serveResponse(w http.ResponseWriter, req request) bool {
	// writing headers to response
	addHeader(req.Headers, w.Header())

	// set was the incoming content-length
	w.Header().Del("Content-Length")

	// set status from script
	w.WriteHeader(req.Status)

	// write response body
	// nolint
	w.Write([]byte(req.Body))

	return false
}

func addHeader(getHeader, setHeader http.Header) {
	for k, v := range getHeader {
		for a := range v {
			setHeader.Add(k, v[a])
		}
	}
}
