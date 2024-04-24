package states

import (
	"context"
	"errors"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/model"
)

//nolint:gochecknoinits
func init() {
	RegisterState(model.StateTypeNoop, Noop)
}

type noopLogic struct {
	*model.NoopState
	Instance
}

// Noop initializes the logic for executing a 'noop' state in a Direktiv workflow instance.
func Noop(instance Instance, state model.State) (Logic, error) {
	noop, ok := state.(*model.NoopState)
	if !ok {
		return nil, derrors.NewInternalError(errors.New("bad state object"))
	}

	sl := new(noopLogic)
	sl.Instance = instance
	sl.NoopState = noop

	return sl, nil
}

// Run implements the Run function for the Logic interface.
//
// As the simplest state, 'noop' doesn't perform any special functionality, and even the
// limited functionality it does is handled generically by the engine (transform, log, etc).
// Therefore, this logic should schedule in just once. And the only ways it might fail is
// if it performs unnecessary validation on its arguments and finds them somehow broken.
// Like if either the 'wakedata' or the instance memory is non-nil.
func (logic *noopLogic) Run(ctx context.Context, wakedata []byte) (*Transition, error) {
	err := scheduleOnce(logic, wakedata)
	if err != nil {
		return nil, err
	}

	return &Transition{
		Transform: logic.Transform,
		NextState: logic.Transition,
	}, nil
}
