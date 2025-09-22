package engine

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/internal/telemetry"
	"github.com/google/uuid"
	"github.com/grafana/sobek"
)

type Commands struct {
	vm *sobek.Runtime
	id uuid.UUID
}

func InjectCommands(vm *sobek.Runtime) {
	cmds := &Commands{
		vm: vm,
	}

	vm.Set("finish", cmds.finish)
	vm.Set("transition", cmds.transition)
	vm.Set("log", cmds.log)
}

func (cmds *Commands) log(call sobek.FunctionCall) sobek.Value {
	telemetry.LogInstance(context.Background(), telemetry.LogLevelInfo, fmt.Sprintf("%v", call.Arguments[0].Export()))

	return sobek.Undefined()
}

// transition needs to throw uncatchable errors
// not: panic(cmds.vm.NewTypeError("transition requires a function and a payload"))
// or go errors: panic(cmds.vm.NewGoError(fmt.Errorf("finish requires one argument, but got %d", len(call.Arguments))))
func (cmds *Commands) transition(call sobek.FunctionCall) sobek.Value {
	if len(call.Arguments) != 2 {
		panic(fmt.Errorf("transition requires a function and a payload"))
	}

	fn, ok := sobek.AssertFunction(call.Arguments[0])
	if !ok {
		panic(fmt.Errorf("first parameter of transition is not a function"))
	}

	value, err := fn(sobek.Undefined(), call.Arguments[1])
	if err != nil {
		panic(fmt.Errorf("error executing the transition: %s", err.Error()))
	}

	return value
}

func (cmds *Commands) finish(call sobek.FunctionCall) sobek.Value {

	if len(call.Arguments) != 1 {
		panic(fmt.Errorf("finish requires one argument, but got %d", len(call.Arguments)))
	}

	return call.Arguments[0]
}
