package runtime

import (
	"context"
	"encoding/json"
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
	vm           *sobek.Runtime
	instID       uuid.UUID
	metadata     map[string]string
	onFinish     OnFinishFunc
	onTransition OnTransitionFunc
}

type (
	OnFinishFunc     func(output []byte) error
	OnTransitionFunc func(output []byte, fn string) error
)

var (
	NoOnFinish     = func(output []byte) error { return nil }
	NoOnTransition = func(output []byte, fn string) error { return nil }
)

func New(instID uuid.UUID, metadata map[string]string, mappings string, onFinish OnFinishFunc, onTransition OnTransitionFunc) *Runtime {
	vm := sobek.New()
	vm.SetMaxCallStackSize(256)

	if mappings != "" {
		vm.SetParserOptions(parser.WithSourceMapLoader(func(path string) ([]byte, error) {
			return []byte(mappings), nil
		}))
	}

	rt := &Runtime{
		vm:           vm,
		instID:       instID,
		metadata:     metadata,
		onFinish:     onFinish,
		onTransition: onTransition,
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

	var memory any
	if err := rt.vm.ExportTo(call.Arguments[1], &memory); err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error exporting transition data: %s", err.Error())))
	}
	b, err := json.Marshal(memory)
	if err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error marshaling transition data: %s", err.Error())))
	}
	var f string
	if err := rt.vm.ExportTo(call.Arguments[0], &f); err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error exporting transition fn: %s", err.Error())))
	}
	fName := ParseFuncNameFromText(f)
	if fName == "" {
		panic(rt.vm.ToValue(fmt.Sprintf("error parsing transition fn: %s", f)))
	}

	err = rt.onTransition(b, fName)
	if err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error calling on transition: %s", err.Error())))
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

// TODO: remove return from finish() as it should be the last statement.
func (rt *Runtime) finish(data sobek.Value) sobek.Value {
	var output any
	if err := rt.vm.ExportTo(data, &output); err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error exporting output: %s", err.Error())))
	}
	b, err := json.Marshal(output)
	if err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error marshaling output: %s", err.Error())))
	}

	err = rt.onFinish(b)
	if err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error calling on finish: %s", err.Error())))
	}

	return sobek.Null()
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

// TODO: this need to be removed.
func (rt *Runtime) RunScript(name, src string) (sobek.Value, error) {
	return rt.vm.RunScript(name, src)
}

// TODO: this need to be removed.
func (rt *Runtime) RunString(str string) (sobek.Value, error) {
	return rt.vm.RunScript("", str)
}

// TODO: this need to be removed.
func (rt *Runtime) GetVar(name string) sobek.Value {
	return rt.vm.Get(name)
}

type Script struct {
	InstID   uuid.UUID
	Text     string
	Mappings string
	Fn       string
	Input    string
	Metadata map[string]string
}

func ExecScript(script *Script, onFinish OnFinishFunc, onTransition OnTransitionFunc) error {
	rt := New(script.InstID, script.Metadata, script.Mappings, onFinish, onTransition)

	_, err := rt.vm.RunString(script.Text)
	if err != nil {
		return fmt.Errorf("run script: %w", err)
	}
	start, ok := sobek.AssertFunction(rt.vm.Get(script.Fn))
	if !ok {
		return fmt.Errorf("start function '%s' does not exist", script.Fn)
	}

	var inputMap any
	err = json.Unmarshal([]byte(script.Input), &inputMap)
	if err != nil {
		return fmt.Errorf("unmarshal input: %w", err)
	}

	_, err = start(sobek.Undefined(), rt.vm.ToValue(inputMap))
	if err != nil {
		return fmt.Errorf("invoke start: %w", err)
	}

	return nil
}

func ParseFuncNameFromText(s string) string {
	s = strings.TrimSpace(s)

	const prefix = "function "
	if !strings.HasPrefix(s, prefix) {
		return ""
	}
	s = s[len(prefix):]

	// find the first '(' to isolate the name
	if idx := strings.Index(s, "("); idx != -1 {
		return strings.TrimSpace(s[:idx])
	}

	return ""
}
