package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/telemetry"
	"github.com/google/uuid"
	"github.com/grafana/sobek"
)

type Commands struct {
	vm       *sobek.Runtime
	instID   uuid.UUID
	metadata map[string]string
}

func InjectCommands(vm *sobek.Runtime, instID uuid.UUID, metadata map[string]string) {
	cmds := &Commands{
		vm:       vm,
		instID:   instID,
		metadata: metadata,
	}

	vm.Set("finish", cmds.finish)
	vm.Set("transition", cmds.transition)
	vm.Set("log", cmds.log)
	vm.Set("print", cmds.print)
	vm.Set("id", cmds.id)
	vm.Set("now", cmds.now)
	vm.Set("fetch", cmds.fetch)
	vm.Set("fetchSync", cmds.fetchSync)
	vm.Set("sleep", cmds.sleep)
	vm.Set("generateAction", cmds.action)
}

func (cmds *Commands) sleep(seconds int) sobek.Value {
	time.Sleep(time.Duration(seconds) * time.Second)
	return sobek.Undefined()
}

func (cmds *Commands) now() *sobek.Object {
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

func (cmds *Commands) id() sobek.Value {
	return cmds.vm.ToValue(cmds.instID)
}

func (cmds *Commands) log(logs ...string) sobek.Value {

	ctx := context.WithValue(context.Background(), "scope", "instance")
	ctx = context.WithValue(ctx, "namespace", cmds.metadata[core.FlowActionScopeNamespace])
	ctx = context.WithValue(ctx, "callpath", fmt.Sprintf("/%s/", cmds.instID.String()))

	slog.Info(strings.Join(logs, " "), slog.String("scope", "instance"), slog.String("namespace", "test"), slog.String("callpath", fmt.Sprintf("/%s/", cmds.instID.String())))

	// query=scope:=instance namespace:=test callpath:/d460f882-46a9-47ff-8ccf-e431e9b5f128/* _time:>2025-10-01T11:48:02.462000000Z

	telemetry.LogInstance(context.Background(), telemetry.LogLevelInfo, strings.Join(logs, " "))
	return sobek.Undefined()
}

func (cmds *Commands) transition(call sobek.FunctionCall) sobek.Value {
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

func (cmds *Commands) finish(data any) sobek.Value {
	return cmds.vm.ToValue(data)
}

func (cmds *Commands) print(args ...any) {
	fmt.Print(args[0])
	if len(args) > 1 {
		for _, arg := range args[1:] {
			fmt.Print(" ")
			fmt.Print(arg)
		}
	}
	fmt.Println()
}
