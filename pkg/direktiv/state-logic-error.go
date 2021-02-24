package direktiv

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

func (sl *errorStateLogic) Deadline() time.Time {
	return time.Now().Add(time.Second * 5)
}

func (sl *errorStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *errorStateLogic) ID() string {
	return sl.state.GetID()
}

func (sl *errorStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *errorStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	a := make([]interface{}, len(sl.state.Args))

	for i := 0; i < len(a); i++ {
		var x interface{}
		x, err = jqObject(instance.data, sl.state.Args[i])
		if err != nil {
			return
		}
		a[i] = x
	}

	err = instance.Raise(ctx, NewCatchableError(sl.state.Error, sl.state.Message, a...))
	if err != nil {
		return
	}

	transition = &stateTransition{
		Transform: sl.state.Transform,
		NextState: sl.state.Transition,
	}

	return

}
