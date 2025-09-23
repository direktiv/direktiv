package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/internal/telemetry"
	"github.com/google/uuid"
	"github.com/grafana/sobek"
)

type Commands struct {
	vm     *sobek.Runtime
	instID uuid.UUID
}

func InjectCommands(vm *sobek.Runtime, instID uuid.UUID) {
	cmds := &Commands{
		vm:     vm,
		instID: instID,
	}

	vm.Set("finish", cmds.finish)
	vm.Set("transition", cmds.transition)
	vm.Set("log", cmds.log)
	vm.Set("print", cmds.print)
	vm.Set("id", cmds.id)
	vm.Set("now", cmds.now)
}

func (cmds *Commands) now(_ sobek.FunctionCall) *sobek.Object {
	t := time.Now()

	obj := cmds.vm.NewObject()
	obj.Set("Unix", func(call sobek.FunctionCall) sobek.Value {
		fmt.Println("ssss")
		return cmds.vm.ToValue(t.Unix())
	})
	obj.Set("Add", func(call sobek.FunctionCall) sobek.Value {
		fmt.Println("ssss")
		return cmds.vm.ToValue(t.Unix())
	})
	obj.Set("Format", func(call sobek.FunctionCall) sobek.Value {
		if len(call.Arguments) < 1 {
			panic(cmds.vm.ToValue("time format required"))
		}
		layout := call.Arguments[0].String()

		return cmds.vm.ToValue(t.Format(layout))
	})
	obj.Set("After", func(call sobek.FunctionCall) sobek.Value {
		fmt.Println("ssss")
		return cmds.vm.ToValue(t.Unix())
	})
	obj.Set("Before", func(call sobek.FunctionCall) sobek.Value {
		fmt.Println("ssss")
		return cmds.vm.ToValue(t.Unix())
	})

	return obj
}

func (cmds *Commands) id(_ sobek.FunctionCall) sobek.Value {
	return cmds.vm.ToValue(cmds.instID)
}

func (cmds *Commands) log(call sobek.FunctionCall) sobek.Value {
	telemetry.LogInstance(context.Background(), telemetry.LogLevelInfo, fmt.Sprintf("%v", call.Arguments[0].Export()))

	return sobek.Undefined()
}

// transition needs to throw uncatchable errors
// panic(cmds.vm.NewGoError(fmt.Errorf("finish requires one argument, but got %d", len(call.Arguments)))).
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

func (cmds *Commands) finish(call sobek.FunctionCall) sobek.Value {
	if len(call.Arguments) != 1 {
		panic(cmds.vm.ToValue(fmt.Sprintf("finish requires one argument, but got %d", len(call.Arguments))))
	}

	return call.Arguments[0]
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
