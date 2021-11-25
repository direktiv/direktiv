package flow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/model"
	"github.com/senseyeio/duration"
)

type delayStateLogic struct {
	state *model.DelayState
}

func initDelayStateLogic(wf *model.Workflow, state model.State) (stateLogic, error) {

	delay, ok := state.(*model.DelayState)
	if !ok {
		return nil, NewInternalError(errors.New("bad state object"))
	}

	sl := new(delayStateLogic)
	sl.state = delay

	return sl, nil

}

func (sl *delayStateLogic) Type() string {
	return model.StateTypeDelay.String()
}

func (sl *delayStateLogic) Deadline(ctx context.Context, engine *engine, im *instanceMemory) time.Time {

	d, err := duration.ParseISO8601(sl.state.Duration)
	if err != nil {
		engine.logToInstance(ctx, time.Now(), im.in, "failed to parse duration: %v", err)
		return time.Now().Add(defaultDeadline)
	}

	t := d.Shift(time.Now().Add(defaultDeadline))
	return t

}

func (sl *delayStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *delayStateLogic) ID() string {
	return sl.state.GetID()
}

func (sl *delayStateLogic) LivingChildren(ctx context.Context, engine *engine, im *instanceMemory) []stateChild {
	return nil
}

func (sl *delayStateLogic) LogJQ() interface{} {
	return sl.state.Log
}

func (sl *delayStateLogic) Run(ctx context.Context, engine *engine, im *instanceMemory, wakedata []byte) (transition *stateTransition, err error) {

	if im.GetMemory() != nil {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) == 0 {

		var d duration.Duration

		d, err = duration.ParseISO8601(sl.state.Duration)
		if err != nil {
			err = NewInternalError(fmt.Errorf("failed to parse delay duration: %v", err))
			return
		}

		t := d.Shift(time.Now())

		err = engine.InstanceSleep(ctx, im, sl.ID(), t)
		if err != nil {
			return
		}

		return

	} else if string(wakedata) == sleepWakedata {

		engine.logToInstance(ctx, time.Now(), im.in, "Waking up from sleep.")

		transition = &stateTransition{
			Transform: sl.state.Transform,
			NextState: sl.state.Transition,
		}

		return

	} else {

		err = NewInternalError(fmt.Errorf("unexpected wakedata for delay state: %s", wakedata))

		return

	}

}
