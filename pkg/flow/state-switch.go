package flow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/model"
)

type switchStateLogic struct {
	state *model.SwitchState
}

func initSwitchStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	switchState, ok := state.(*model.SwitchState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(switchStateLogic)
	sl.state = switchState

	return sl, nil
}

func (sl *switchStateLogic) Type() string {
	return model.StateTypeSwitch.String()
}

func (sl *switchStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {
	return time.Now().Add(defaultDeadline)
}

func (sl *switchStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *switchStateLogic) ID() string {
	return sl.state.GetID()
}

func (sl *switchStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {
	return nil
}

func (sl *switchStateLogic) LogJQ() interface{} {
	return sl.state.Log
}

func (sl *switchStateLogic) Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error) {

	if im.GetMemory() != nil {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	for i, condition := range sl.state.Conditions {

		var x interface{}
		x, err = jqOne(im.data, condition.Condition)
		if err != nil {
			err = NewInternalError(fmt.Errorf("switch condition %d condition failed to run: %v", i, err))
			return
		}

		if truth(x) {
			engine.logToInstance(ctx, time.Now(), im.in, "Switch condition %d succeeded", i)
			transition = &stateTransition{
				Transform: condition.Transform,
				NextState: condition.Transition,
			}
			return
		}

	}

	engine.logToInstance(ctx, time.Now(), im.in, "No switch conditions succeeded")
	transition = &stateTransition{
		Transform: sl.state.DefaultTransform,
		NextState: sl.state.DefaultTransition,
	}

	return

}
