package direktiv

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/senseyeio/duration"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/model"
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

func (sl *delayStateLogic) Deadline() time.Time {

	d, err := duration.ParseISO8601(sl.state.Duration)
	if err != nil {
		log.Errorf("failed to parse duration: %v", err)
		return time.Now()
	}

	t := d.Shift(time.Now().Add(time.Second * 5))
	return t

}

func (sl *delayStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *delayStateLogic) ID() string {
	return sl.state.GetID()
}

func (sl *delayStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func (sl *delayStateLogic) LogJQ() interface{} {
	return sl.state.Log
}

func (sl *delayStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(savedata) != 0 {
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

		err = instance.engine.sleep(instance.id, sl.ID(), instance.step, t)
		if err != nil {
			return
		}

		return

	} else if string(wakedata) == sleepWakedata {

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
