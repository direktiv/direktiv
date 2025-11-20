package runtime

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/direktiv/direktiv/internal/core"
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
	onSubflow    OnSubflowFunc
	//nolint:containedctx
	ctx context.Context
}

type (
	OnFinishFunc     func(output []byte) error
	OnTransitionFunc func(output []byte, fn string) error
	OnActionFunc     func(svcID string) error

	OnSubflowFunc func(ctx context.Context, path string, input []byte) ([]byte, error)
)

var (
	NoOnFinish     = func(output []byte) error { return nil }
	NoOnTransition = func(output []byte, fn string) error { return nil }
	NoOnAction     = func(svcID string) error { return nil }
	NoOnSubflow    = func(ctx context.Context, path string, input []byte) ([]byte, error) { return nil, nil }
)

func New(ctx context.Context, instID uuid.UUID, metadata map[string]string, mappings string,
	onFinish OnFinishFunc, onTransition OnTransitionFunc, onAction OnActionFunc, onSubflow OnSubflowFunc,
) *Runtime {
	vm := sobek.New()
	vm.SetMaxCallStackSize(256)

	if mappings != "" {
		vm.SetParserOptions(parser.WithSourceMapLoader(func(path string) ([]byte, error) {
			return []byte(mappings), nil
		}))
	}

	rt := &Runtime{
		ctx:          ctx,
		vm:           vm,
		instID:       instID,
		metadata:     metadata,
		onFinish:     onFinish,
		onTransition: onTransition,
		onAction:     onAction,
		onSubflow:    onSubflow,
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
		{"getSecret", rt.secret},
		{"execSubflow", rt.execSubflow},
	}

	for _, v := range setList {
		if err := vm.Set(v.name, v.fn); err != nil {
			panic(fmt.Sprintf("error setting runtime function '%s': %s", v.name, err.Error()))
		}
	}

	return rt
}

func (rt *Runtime) secret(secretName string) sobek.Value {
	secretJsonMap := rt.metadata[core.EngineMappingSecrets]
	us := make(map[string]string)
	json.Unmarshal([]byte(secretJsonMap), &us)

	value, ok := us[secretName]
	if !ok {
		panic(rt.vm.ToValue(fmt.Sprintf("error fetching secret %s: %s",
			secretName, fmt.Errorf("secret not available"))))
	}

	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error fetching secret %s: %s",
			secretName, err)))
	}

	return rt.vm.ToValue(string(decoded))
}

func (rt *Runtime) secrets(secretNames []string) sobek.Value {
	secretJsonMap := rt.metadata[core.EngineMappingSecrets]

	us := make(map[string]string)
	json.Unmarshal([]byte(secretJsonMap), &us)

	retSecrets := make(map[string]string)
	for i := range secretNames {
		value, ok := us[secretNames[i]]
		if !ok {
			panic(rt.vm.ToValue(fmt.Sprintf("error fetching secret %s: %s",
				secretNames[i], fmt.Errorf("secret not available"))))
		}

		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			panic(rt.vm.ToValue(fmt.Sprintf("error fetching secret %s: %s",
				secretNames[i], err)))
		}

		retSecrets[secretNames[i]] = string(decoded)
	}

	return rt.vm.ToValue(retSecrets)
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

	out, err := rt.onSubflow(rt.ctx, path, b)
	if err != nil {
		panic(rt.vm.ToValue(fmt.Sprintf("error calling on subflow: %s", err.Error())))
	}

	return rt.vm.ToValue(out)
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

type Script struct {
	InstID   uuid.UUID
	Text     string
	Mappings string
	Fn       string
	Input    string
	Metadata map[string]string
}

func ExecScript(ctx context.Context, script *Script, onFinish OnFinishFunc,
	onTransition OnTransitionFunc, onAction OnActionFunc, onSubflow OnSubflowFunc,
) error {
	rt := New(ctx, script.InstID, script.Metadata, script.Mappings, onFinish,
		onTransition, onAction, onSubflow)

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
