package flow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/model"
)

type errorStateLogic struct {
	*model.ErrorState
}

func initErrorStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	err, ok := state.(*model.ErrorState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(errorStateLogic)
	sl.ErrorState = err

	return sl, nil

}

func (sl *errorStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {
	return time.Now().Add(defaultDeadline)
}

func (sl *errorStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {
	return nil
}

func (sl *errorStateLogic) Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error) {

	if im.GetMemory() != nil {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	a := make([]interface{}, len(sl.Args))

	for i := 0; i < len(a); i++ {
		var x interface{}
		x, err = jqOne(im.data, sl.Args[i])
		if err != nil {
			return
		}
		a[i] = x
	}

	x, err := jqOne(im.data, sl.Message)
	if err != nil {
		return
	}

	msg, ok := x.(string)
	if !ok {
		msg = fmt.Sprintf("%v", x)
	}

	err = engine.InstanceRaise(ctx, im, NewCatchableError(sl.Error, msg, a...))
	if err != nil {
		return
	}

	transition = &stateTransition{
		Transform: sl.Transform,
		NextState: sl.Transition,
	}

	return

}
