package states

import (
	"context"
	"errors"
	"fmt"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/model"
)

func init() {
	RegisterState(model.StateTypeSwitch, Switch)
}

type switchLogic struct {
	*model.SwitchState
	Instance
}

// Switch initializes the logic for executing a 'switch' state in a Direktiv workflow instance.
func Switch(instance Instance, state model.State) (Logic, error) {
	s, ok := state.(*model.SwitchState)
	if !ok {
		return nil, derrors.NewInternalError(errors.New("bad state object"))
	}

	sl := new(switchLogic)
	sl.Instance = instance
	sl.SwitchState = s

	return sl, nil
}

// Run implements the Run function for the Logic interface.
//
// The 'switch' evaluates one or more conditions against the instance data to determine which
// transform and which transition to use. The logic only needs to be scheduled in once. The
// most likely way for the logic to fail is a JQ error against the instance data.
func (logic *switchLogic) Run(ctx context.Context, wakedata []byte) (*Transition, error) {
	err := scheduleOnce(logic, wakedata)
	if err != nil {
		return nil, err
	}

	for i, condition := range logic.Conditions {
		var x interface{}
		x, err = jqOne(logic.GetInstanceData(), condition.Condition)
		if err != nil {
			return nil, derrors.NewInternalError(fmt.Errorf("switch condition %d condition failed to run: %w", i, err))
		}

		if truth(x) {
			// logic.Log(ctx, log.Info, "Switch condition %d succeeded", i)

			return &Transition{
				Transform: condition.Transform,
				NextState: condition.Transition,
			}, nil
		}
	}

	// logic.Log(ctx, log.Info, "No switch conditions succeeded")

	return &Transition{
		Transform: logic.DefaultTransform,
		NextState: logic.DefaultTransition,
	}, nil
}
