package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/direktiv/direktiv/internal/telemetry"
	"github.com/google/uuid"
	"github.com/grafana/sobek"
	"github.com/grafana/sobek/parser"
)

type Runtime struct {
	vm       *sobek.Runtime
	instID   uuid.UUID
	metadata map[string]string
}

func New(instID uuid.UUID, metadata map[string]string, mappings string) *Runtime {
	vm := sobek.New()
	vm.SetMaxCallStackSize(256)

	if mappings != "" {
		vm.SetParserOptions(parser.WithSourceMapLoader(func(path string) ([]byte, error) {
			return []byte(mappings), nil
		}))
	}

	rt := &Runtime{
		vm:       vm,
		instID:   instID,
		metadata: metadata,
	}

	vm.Set("finish", rt.finish)
	vm.Set("transition", rt.transition)
	vm.Set("log", rt.log)
	vm.Set("print", rt.print)
	vm.Set("id", rt.id)
	vm.Set("now", rt.now)
	vm.Set("fetch", rt.fetch)
	vm.Set("fetchSync", rt.fetchSync)
	vm.Set("sleep", rt.sleep)
	vm.Set("generateAction", rt.action)

	return rt
}

func (rt *Runtime) action(call sobek.FunctionCall) sobek.Value {
	// imgObject := call.Argument(0).ToObject(rt.vm)

	actionFunc := func(call sobek.FunctionCall) sobek.Value {
		return rt.vm.ToValue("return value")
	}

	return rt.vm.ToValue(actionFunc)
}

func (rt *Runtime) sleep(seconds int) sobek.Value {
	time.Sleep(time.Duration(seconds) * time.Second)
	return sobek.Undefined()
}

func (rt *Runtime) now() *sobek.Object {
	t := time.Now()

	obj := rt.vm.NewObject()

	obj.Set("unix", func() sobek.Value {
		return rt.vm.ToValue(t.Unix())
	})

	obj.Set("format", func(format string) sobek.Value {
		return rt.vm.ToValue(t.Format(format))
	})

	return obj
}

func (rt *Runtime) id() sobek.Value {
	return rt.vm.ToValue(rt.instID)
}

func (rt *Runtime) log(logs ...string) sobek.Value {
	telemetry.LogInstance(context.Background(), telemetry.LogLevelInfo, strings.Join(logs, " "))
	return sobek.Undefined()
}

func (rt *Runtime) transition(call sobek.FunctionCall) sobek.Value {
	if len(call.Arguments) != 2 {
		panic(rt.vm.ToValue("transition requires a function and a payload"))
	}

	fn, ok := sobek.AssertFunction(call.Arguments[0])
	if !ok {
		panic(rt.vm.ToValue("first parameter of transition is not a function"))
	}

	value, err := fn(sobek.Undefined(), call.Arguments[1])
	if err != nil {
		exception := &sobek.Exception{}
		if errors.As(err, &exception) {
			panic(err)
		} else {
			panic(rt.vm.ToValue(fmt.Sprintf("error executing transition: %s", err.Error())))
		}
	}

	return value
}

func (rt *Runtime) finish(data any) sobek.Value {
	return rt.vm.ToValue(data)
}

func (rt *Runtime) print(args ...any) {
	fmt.Print(args[0])
	if len(args) > 1 {
		for _, arg := range args[1:] {
			fmt.Print(" ")
			fmt.Print(arg)
		}
	}
	fmt.Println()
}

func (rt *Runtime) RunScript(name, src string) (sobek.Value, error) {
	return rt.vm.RunScript(name, src)
}

func (rt *Runtime) RunString(str string) (sobek.Value, error) {
	return rt.vm.RunScript("", str)
}

// GetVar the specified variable in the global context.
func (rt *Runtime) GetVar(name string) sobek.Value {
	return rt.vm.Get(name)
}
