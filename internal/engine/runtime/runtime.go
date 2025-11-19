package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/direktiv/direktiv/internal/core"
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
	onAction     OnActionFunc

	tracingPack *tracingPack
}

type (
	OnFinishFunc     func(output []byte) error
	OnTransitionFunc func(output []byte, fn string) error
	OnActionFunc     func(svcID string) error
)

var (
	NoOnFinish     = func(output []byte) error { return nil }
	NoOnTransition = func(output []byte, fn string) error { return nil }
	NoOnAction     = func(svcID string) error { return nil }
)

func New(instID uuid.UUID, metadata map[string]string, mappings string,
	onFinish OnFinishFunc, onTransition OnTransitionFunc, onAction OnActionFunc,
) *Runtime {
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
		onAction:     onAction,
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
		{"getSecrets", rt.secrets},
		{"execSubflow", rt.execSubflow},
	}

	for _, v := range setList {
		if err := vm.Set(v.name, v.fn); err != nil {
			panic(fmt.Sprintf("error setting runtime function '%s': %s", v.name, err.Error()))
		}
	}

	return rt
}

func (rt *Runtime) WithTracingPack(tp *tracingPack) *Runtime {
	rt.tracingPack = tp
	return rt
}

func (rt *Runtime) secrets(secretNames []string) sobek.Value {
	// rt.tracingPack.span.AddEvent("fetching secrets")

	// s := make(map[string]string)
	rt.tracingPack.span.AddEvent("fetching secrets")

	// s := make(map[string]string)

	// for i := range secretNames {
	// 	secret, err := rt.secretsManager.Get(rt.tracingPack.ctx,
	// 		rt.tracingPack.namespace, secretNames[i])
	// 	if err != nil {
	// 		panic(rt.vm.ToValue(fmt.Sprintf("error fetching secret %s: %s",
	// 			secretNames[i], err.Error())))
	// 	}

	// 	s[secretNames[i]] = string(secret.Data)
	// }

	// return rt.vm.ToValue(s)
	// for i := range secretNames {
	// 	secret, err := rt.secretsManager.Get(rt.tracingPack.ctx,
	// 		rt.tracingPack.namespace, secretNames[i])
	// 	if err != nil {
	// 		panic(rt.vm.ToValue(fmt.Sprintf("error fetching secret %s: %s",
	// 			secretNames[i], err.Error())))
	// 	}

	// 	s[secretNames[i]] = string(secret.Data)
	// }

	// return rt.vm.ToValue(s)

	return nil
}

func (rt *Runtime) sleep(seconds int) sobek.Value {
	rt.tracingPack.span.AddEvent("calling sleep")
	time.Sleep(time.Duration(seconds) * time.Second)

	return sobek.Undefined()
}

func (rt *Runtime) now() *sobek.Object {
	rt.tracingPack.span.AddEvent("calling now")
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
	rt.tracingPack.span.AddEvent("calling id")
	return rt.vm.ToValue(rt.instID)
}

func (rt *Runtime) log(logs ...string) sobek.Value {
	rt.tracingPack.span.AddEvent("calling log")

	// protect victoria logs from falling over without
	msg := strings.Join(logs, " ")
	if msg == "" {
		msg = " "
	}

	telemetry.LogInstance(rt.tracingPack.ctx, telemetry.LogLevelInfo, msg)

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

	// otel: end previous and start new one
	rt.tracingPack.tracingTransition(fName)

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

func (rt *Runtime) execSubflow(call sobek.FunctionCall) sobek.Value {
	if len(call.Arguments) != 3 {
		panic(rt.vm.ToValue("exec requires a path, a function and a payload"))
	}

	var path string
	if err := rt.vm.ExportTo(call.Arguments[0], &path); err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error exporting exec path: %s", err.Error())))
	}
	var f string
	if err := rt.vm.ExportTo(call.Arguments[1], &f); err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error exporting exec fn: %s", err.Error())))
	}
	var memory any
	if err := rt.vm.ExportTo(call.Arguments[2], &memory); err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error exporting exec data: %s", err.Error())))
	}
	b, err := json.Marshal(memory)
	if err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error marshaling transition data: %s", err.Error())))
	}

	return rt.vm.ToValue("" + path + "/" + f + "/" + string(b))
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

	// otel: finish span from transition
	rt.tracingPack.tracingFinish()

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

type Script struct {
	InstID   uuid.UUID
	Text     string
	Mappings string
	Fn       string
	Input    string
	Metadata map[string]string
}

func ExecScript(ctx context.Context, script *Script, onFinish OnFinishFunc,
	onTransition OnTransitionFunc, onAction OnActionFunc,
) error {
	tp := newTracingPack(ctx, script.Metadata[core.EngineMappingNamespace],
		script.InstID.String(), script.Metadata[core.EngineMappingCaller],
		script.Metadata[core.EngineMappingPath])
	defer tp.finish()

	rt := New(script.InstID, script.Metadata, script.Mappings, onFinish,
		onTransition, onAction).WithTracingPack(tp)

	tp.tracingStart(script.Fn)
	telemetry.LogInstance(tp.ctx, telemetry.LogLevelInfo,
		fmt.Sprintf("transitioning to '%s'", script.Fn))

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
		rt.tracingPack.handleError(err)
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
