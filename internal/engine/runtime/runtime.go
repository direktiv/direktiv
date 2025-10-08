package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/direktiv/direktiv/internal/core"
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

	type setFunc struct {
		name string
		fn   any
	}
	setList := []setFunc{
		{"finish", rt.finish},
		{"transition", rt.transition},
		{"log", rt.log},
		{"print", rt.print},
		{"id", rt.id},
		{"now", rt.now},
		{"fetch", rt.fetch},
		{"fetchSync", rt.fetchSync},
		{"sleep", rt.sleep},
		{"generateAction", rt.action},
	}

	for _, v := range setList {
		if err := vm.Set(v.name, v.fn); err != nil {
			panic(fmt.Sprintf("error setting runtime function '%s': %s", v.name, err.Error()))
		}
	}

	return rt
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

func (rt *Runtime) log(logs ...any) sobek.Value {
	var b strings.Builder
	for a := range logs {
		b.WriteString(fmt.Sprintf("%v", logs[a]))
	}

	slog.Info(b.String(),
		slog.String("status", "dummy"),
		slog.String("state", "dummy"),
		slog.String("path", rt.metadata[core.EngineMappingPath]),
		slog.String("scope", "instance"),
		slog.String("namespace", "test"),
		slog.String("id", rt.instID.String()),
		slog.String("callpath",
			fmt.Sprintf("/%s/", rt.instID.String())))

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

func ExecScript(instID uuid.UUID, script string, mappings string, fn string,
	input string, metadata map[string]string,
) ([]byte, error) {
	// add commands

	rt := New(instID, metadata, mappings)

	_, err := rt.vm.RunString(script)
	if err != nil {
		return nil, fmt.Errorf("run script: %w", err)
	}
	start, ok := sobek.AssertFunction(rt.vm.Get(fn))
	if !ok {
		return nil, fmt.Errorf("start function '%s' does not exist", fn)
	}

	var inputMap any
	err = json.Unmarshal([]byte(input), &inputMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshal input: %w", err)
	}

	ret, err := start(sobek.Undefined(), rt.vm.ToValue(inputMap))
	if err != nil {
		return nil, fmt.Errorf("invoke start: %w", err)
	}
	var output any
	if err := rt.vm.ExportTo(ret, &output); err != nil {
		return nil, fmt.Errorf("export output: %w", err)
	}
	b, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("marshal output: %w", err)
	}

	return b, nil
}
