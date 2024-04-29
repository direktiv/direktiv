package states

import (
	"context"
	"errors"
	"fmt"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/model"
)

//nolint:gochecknoinits
func init() {
	RegisterState(model.StateTypeError, Error)
}

type errorLogic struct {
	*model.ErrorState
	Instance
}

// Error initializes the logic for executing an 'error' state in a Direktiv workflow instance.
func Error(instance Instance, state model.State) (Logic, error) {
	noop, ok := state.(*model.ErrorState)
	if !ok {
		return nil, derrors.NewInternalError(errors.New("bad state object"))
	}

	sl := new(errorLogic)
	sl.Instance = instance
	sl.ErrorState = noop

	return sl, nil
}

// Run implements the Run function for the Logic interface.
//
// The 'error' escalates a user-defined error to be thrown by the instance when it terminates.
// The 'error' state does not necessarily indicate the end of the workflow. The instance may
// continue to transition, usually to perform some form of cleanup or reverting of the activies
// undertaken by the workflow so far. The logic only needs to be scheduled in once.
func (logic *errorLogic) Run(ctx context.Context, wakedata []byte) (*Transition, error) {
	err := scheduleOnce(logic, wakedata)
	if err != nil {
		return nil, err
	}

	a := make([]interface{}, len(logic.Args))

	for i := 0; i < len(a); i++ {
		var x interface{}

		x, err = jqOne(logic.GetInstanceData(), logic.Args[i]) //nolint:contextcheck
		if err != nil {
			return nil, err
		}

		a[i] = x
	}

	x, err := jqOne(logic.GetInstanceData(), logic.Message) //nolint:contextcheck
	if err != nil {
		return nil, err
	}

	msg, ok := x.(string)
	if !ok {
		msg = fmt.Sprintf("%v", x)
	}

	err = logic.Raise(ctx, derrors.NewCatchableError(logic.Error, msg, a...))
	if err != nil {
		return nil, err
	}

	return &Transition{
		Transform: logic.Transform,
		NextState: logic.Transition,
	}, nil
}
