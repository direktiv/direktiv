package outbound

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"time"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/gateway"
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

func (js *JSOutboundPlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	//nolint:forcetypeassert
	rr := w.(*httptest.ResponseRecorder)
	w = httptest.NewRecorder()

	var err error

	resp := response{
		Headers: rr.Header(),
		Body:    rr.Body.String(),
		Code:    rr.Code,
	}

	// extract all response headers and body
	vm := goja.New()
	err = vm.Set("input", resp)
	if err != nil {
		gateway.WriteInternalError(r, w, err, "can not set input object")
		return nil, nil
	}

	err = vm.Set("log", func(txt interface{}) {
		slog.Info("js log", slog.Any("log", txt))
	})
	if err != nil {
		gateway.WriteInternalError(r, w, err, "can not set log function")
		return nil, nil
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
		return nil, nil
	}

	script := fmt.Sprintf("function run() { %s; return input } run()", js.Script)

	val, err := vm.RunScript("plugin", script)
	if err != nil {
		gateway.WriteInternalError(r, w, err, "can not execute script")
		return nil, nil
	}

	if val != nil && !val.Equals(goja.Undefined()) {
		o := val.ToObject(vm)
		// make sure the input object got returned
		if o.ExportType() == reflect.TypeOf(resp) {
			// nolint checked before
			responseDone := o.Export().(response)
			for k, v := range responseDone.Headers {
				for a := range v {
					w.Header().Set(k, v[a])
				}
			}
			w.WriteHeader(responseDone.Code)
			_, _ = w.Write([]byte(responseDone.Body))
		}
	}

	return w, r
}

func init() {
	gateway.RegisterPlugin(&JSOutboundPlugin{})
}
