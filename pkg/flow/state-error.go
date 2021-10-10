package flow

import (
	"context"
	"errors"
	"time"

	"github.com/vorteil/direktiv/pkg/model"
)

type errorStateLogic struct {
	state *model.ErrorState
}

func initErrorStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	err, ok := state.(*model.ErrorState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(errorStateLogic)
	sl.state = err

	return sl, nil
}

func (sl *errorStateLogic) Type() string {
	return model.StateTypeError.String()
}

func (sl *errorStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {
	return time.Now().Add(defaultDeadline)
}

func (sl *errorStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *errorStateLogic) ID() string {
	return sl.state.GetID()
}

func (sl *errorStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {
	return nil
}

func (sl *errorStateLogic) LogJQ() interface{} {
	return sl.state.Log
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

	a := make([]interface{}, len(sl.state.Args))

	for i := 0; i < len(a); i++ {
		var x interface{}
		x, err = jqOne(im.data, sl.state.Args[i])
		if err != nil {
			return
		}
		a[i] = x
	}

	err = engine.InstanceRaise(ctx, im, NewCatchableError(sl.state.Error, sl.state.Message, a...))
	if err != nil {
		return
	}

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}
