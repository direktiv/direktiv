package flow

import (
	"context"
	"errors"
	"time"

	"github.com/direktiv/direktiv/pkg/model"
)

type noopStateLogic struct {
	*model.NoopState
}

func initNoopStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	noop, ok := state.(*model.NoopState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(noopStateLogic)
	sl.NoopState = noop

	return sl, nil

}

func (sl *noopStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {
	return time.Now().Add(defaultDeadline)
}

func (sl *noopStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {
	return nil
}

func (sl *noopStateLogic) Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error) {

	if im.GetMemory() != nil {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	transition = &stateTransition{
		Transform: sl.Transform,
		NextState: sl.Transition,
	}

	return

}
