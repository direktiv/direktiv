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

func (cmds *Commands) log(call sobek.FunctionCall) sobek.Value {
	telemetry.LogInstance(context.Background(), telemetry.LogLevelInfo, fmt.Sprintf("%v", call.Arguments[0].Export()))

	return sobek.Undefined()
}

func (cmds *Commands) transition(call sobek.FunctionCall) sobek.Value {
	// TODO CHECK IF THERE ARE TWO ARGUMENTS
	fn, ok := sobek.AssertFunction(call.Arguments[0])
	if !ok {
		panic(cmds.vm.NewTypeError("first parameter of transition is not a function"))
	}

	fn(call.Arguments[1])

	return sobek.Undefined()
}

func (cmds *Commands) finish(call sobek.FunctionCall) sobek.Value {
	return sobek.Undefined()
}
