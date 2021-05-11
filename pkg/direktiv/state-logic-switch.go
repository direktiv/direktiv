package direktiv

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vorteil/direktiv/pkg/model"
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

func (sl *switchStateLogic) Deadline() time.Time {
	return time.Now().Add(time.Second * 5)
}

func (sl *switchStateLogic) Retries() *model.RetryDefinition {
	return sl.state.RetryDefinition()
}

func (sl *switchStateLogic) ErrorCatchers() []model.ErrorDefinition {
	return sl.state.ErrorDefinitions()
}

func (sl *switchStateLogic) ID() string {
	return sl.state.GetID()
}

func (sl *switchStateLogic) LivingChildren(savedata []byte) []stateChild {
	return nil
}

func truth(x interface{}) bool {

	var success bool

	if x != nil {
		switch x.(type) {
		case bool:
			if x.(bool) {
				success = true
			}
		case string:
			if x.(string) != "" {
				success = true
			}
		case int:
			if x.(int) != 0 {
				success = true
			}
		case []interface{}:
			if len(x.([]interface{})) > 0 {
				success = true
			}
		case map[string]interface{}:
			if len(x.(map[string]interface{})) > 0 {
				success = true
			}
		default:
		}
	}

	return success

}

func (sl *switchStateLogic) LogJQ() string {
	return sl.state.Log
}

func (sl *switchStateLogic) Run(ctx context.Context, instance *workflowLogicInstance, savedata, wakedata []byte) (transition *stateTransition, err error) {

	if len(savedata) != 0 {
		err = NewInternalError(errors.New("got unexpected savedata"))
		return
	}

	if len(wakedata) != 0 {
		err = NewInternalError(errors.New("got unexpected wakedata"))
		return
	}

	for i, condition := range sl.state.Conditions {

		var x interface{}
		x, err = jqOne(instance.data, condition.Condition)
		if err != nil {
			err = NewInternalError(fmt.Errorf("switch condition %d condition failed to run: %v", i, err))
			return
		}

		if truth(x) {
			instance.Log("Switch condition %d succeeded", i)
			transition = &stateTransition{
				Transform: condition.Transform,
				NextState: condition.Transition,
			}
			return
		}

	}

	instance.Log("No switch conditions succeeded")
	transition = &stateTransition{
		Transform: sl.state.DefaultTransform,
		NextState: sl.state.DefaultTransition,
	}

	return

}
