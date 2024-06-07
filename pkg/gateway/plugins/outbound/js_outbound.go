package outbound

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
	"github.com/dop251/goja"
)

type JSOutboundPlugin struct {
	Script string `mapstructure:"script" yaml:"script"`
}

func (js *JSOutboundPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &JSOutboundPlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (js *JSOutboundPlugin) Type() string {
	return "js-outbound"
}

type response struct {
	Headers http.Header
	Body    string
	Code    int
}

func (js *JSOutboundPlugin) Execute(w http.ResponseWriter, r *http.Request) *http.Request {
	var (
		err error
		b   []byte
	)

	if r.Body != nil {
		b, err = io.ReadAll(r.Body)
		if err != nil {
			gateway.WriteInternalError(r, w, err, "can not set read body for js plugin")
			return nil
		}
		defer r.Body.Close()
	}

	if r.Response == nil {
		r.Response = &http.Response{
			StatusCode: http.StatusOK,
		}
	}

	resp := response{
		Headers: r.Header,
		Body:    string(b),
		Code:    r.Response.StatusCode,
	}

	// extract all response headers and body
	vm := goja.New()
	err = vm.Set("input", resp)
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

	err = vm.Set("sleep", func(t interface{}) {
		tt, ok := t.(int64)
		if !ok {
			return
		}
		time.Sleep(time.Duration(tt) * time.Second)
	})
	if err != nil {
		gateway.WriteInternalError(r, w, err, "can not set sleep function")
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
		o := val.ToObject(vm)
		// make sure the input object got returned
		if o.ExportType() == reflect.TypeOf(resp) {
			// nolint checked before
			responseDone := o.Export().(response)
			for k, v := range responseDone.Headers {
				for a := range v {
					w.Header().Add(k, v[a])
				}
			}

			// nolint
			w.Write([]byte(responseDone.Body))
			w.WriteHeader(responseDone.Code)
		}
	}

	return r
}

func init() {
	gateway.RegisterPlugin(&JSOutboundPlugin{})
}
