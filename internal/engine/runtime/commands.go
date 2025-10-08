package runtime

import (
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

func (cmds *Runtime) sleep(seconds int) sobek.Value {
	time.Sleep(time.Duration(seconds) * time.Second)
	return sobek.Undefined()
}

func (cmds *Runtime) now() *sobek.Object {
	t := time.Now()

	obj := cmds.vm.NewObject()

	obj.Set("unix", func() sobek.Value {
		return cmds.vm.ToValue(t.Unix())
	})

	obj.Set("format", func(format string) sobek.Value {
		return cmds.vm.ToValue(t.Format(format))
	})

	return obj
}

func (cmds *Runtime) id() sobek.Value {
	return cmds.vm.ToValue(cmds.instID)
}

func (cmds *Runtime) log(logs ...any) sobek.Value {
	var b strings.Builder
	for a := range logs {
		b.WriteString(fmt.Sprintf("%v", logs[a]))
	}

	slog.Info(b.String(),
		slog.String("status", "dummy"),
		slog.String("state", "dummy"),
		slog.String("path", cmds.metadata[core.EngineMappingPath]),
		slog.String("scope", "instance"),
		slog.String("namespace", "test"),
		slog.String("id", cmds.instID.String()),
		slog.String("callpath",
			fmt.Sprintf("/%s/", cmds.instID.String())))

	return sobek.Undefined()
}

func (cmds *Runtime) transition(call sobek.FunctionCall) sobek.Value {
	if len(call.Arguments) != 2 {
		panic(cmds.vm.ToValue("transition requires a function and a payload"))
	}

	fn, ok := sobek.AssertFunction(call.Arguments[0])
	if !ok {
		panic(cmds.vm.ToValue("first parameter of transition is not a function"))
	}

	value, err := fn(sobek.Undefined(), call.Arguments[1])
	if err != nil {
		exception := &sobek.Exception{}
		if errors.As(err, &exception) {
			panic(err)
		} else {
			panic(cmds.vm.ToValue(fmt.Sprintf("error executing transition: %s", err.Error())))
		}
	}

	return value
}

func (cmds *Runtime) finish(data any) sobek.Value {
	return cmds.vm.ToValue(data)
}

func (cmds *Runtime) print(args ...any) {
	fmt.Print(args[0])
	if len(args) > 1 {
		for _, arg := range args[1:] {
			fmt.Print(" ")
			fmt.Print(arg)
		}
	}
	fmt.Println()
}

func (cmds *Runtime) RunScript(name, src string) (sobek.Value, error) {
	return cmds.vm.RunScript(name, src)
}

func (cmds *Runtime) RunString(str string) (sobek.Value, error) {
	return cmds.vm.RunScript("", str)
}
