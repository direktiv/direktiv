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
	"github.com/direktiv/direktiv/pkg/gateway"
	"github.com/dop251/goja"
)

// JSInboundPlugin allows to modify headers, contents and query params of the request.
type JSInboundPlugin struct {
	Script string `mapstructure:"script" yaml:"script"`
}

func (js *JSInboundPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &JSInboundPlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (js *JSInboundPlugin) Type() string {
	return "js-inbound"
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

	Consumer *core.Consumer

	// url params of type /{id}
	URLParams map[string]string

	// response code after executing
	Status int
}

func (js *JSInboundPlugin) Execute(w http.ResponseWriter, r *http.Request) *http.Request {
	var (
		err error
		b   []byte
	)

	if r.Body != nil {
		b, err = io.ReadAll(r.Body)
		if err != nil {
			gateway.WriteInternalError(r, w, err, "can not set read body for js inbound plugin")
			return nil
		}
		defer r.Body.Close()
	}

	vm := goja.New()

	c := gateway.ExtractContextActiveConsumer(r)

	// add url param
	urlParams := make(map[string]string)
	for _, param := range gateway.ExtractContextURLParams(r) {
		urlParams[param] = r.PathValue(param)
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
		gateway.WriteInternalError(r, w, err, "can not set input object")
		return nil
	}

	err = vm.Set("log", func(txt interface{}) {
		slog.Info("js log", slog.Any("log", txt))
	})
	if err != nil {
		gateway.WriteInternalError(r, w, err, "can not set log function")
		return nil
	}

	script := fmt.Sprintf("function run() { %s; return input } run()",
		js.Script)

	val, err := vm.RunScript("plugin", script)
	if err != nil {
		gateway.WriteInternalError(r, w, err, "can not execute script")
		return nil
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
				// writing headers to response
				addHeader(responseDone.Headers, w.Header())

				// set was the incoming content-length
				w.Header().Del("Content-Length")

				// set status from script
				w.WriteHeader(responseDone.Status)

				// write response body
				// nolint
				w.Write([]byte(responseDone.Body))

				return nil
			}
		}
	}

	return r
}

func init() {
	gateway.RegisterPlugin(&JSInboundPlugin{})
}

func addHeader(getHeader, setHeader http.Header) {
	for k, v := range getHeader {
		for a := range v {
			setHeader.Add(k, v[a])
		}
	}
}
