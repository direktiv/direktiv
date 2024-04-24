package states

import (
	"context"
	"errors"
	"fmt"
	"time"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	log "github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/senseyeio/duration"
)

//nolint:gochecknoinits
func init() {
	RegisterState(model.StateTypeDelay, Delay)
}

type delayLogic struct {
	*model.DelayState
	Instance
}

// Delay initializes the logic for executing a 'delay' state in a Direktiv workflow instance.
func Delay(instance Instance, state model.State) (Logic, error) {
	delay, ok := state.(*model.DelayState)
	if !ok {
		return nil, derrors.NewInternalError(errors.New("bad state object"))
	}

	sl := new(delayLogic)
	sl.Instance = instance
	sl.DelayState = delay

	return sl, nil
}

// Deadline overwrites the default underlying Deadline function provided by Instance because
// Delay is a multi-step state.
func (logic *delayLogic) Deadline(ctx context.Context) time.Time {
	d, err := duration.ParseISO8601(logic.Duration)
	if err != nil {
		logic.Log(ctx, log.Error, "failed to parse duration: %v", err)
		return time.Now().UTC().Add(DefaultShortDeadline)
	}

	t := d.Shift(time.Now().UTC().Add(DefaultShortDeadline))

	return t
}

// Run implements the Run function for the Logic interface.
//
// The 'delay' state does nothing except pause the workflow for a specified length of time. To
// achieve this, the state must be scheduled in twice. The first time Run is called the state
// schedules its own wakeup. The second time Run is called should be in response to the scheduled
// wakeup.
//
// In every other way, the 'delay' state is equivalent to the 'noop' state. It should only fail
// if performs unnecessary validation on its arguments and finds them broken.
func (logic *delayLogic) Run(ctx context.Context, wakedata []byte) (*Transition, error) {
	first, err := scheduleTwiceConst(logic, wakedata, `""`)
	if err != nil {
		return nil, err
	}

	if first {
		var d duration.Duration

		d, err = duration.ParseISO8601(logic.Duration)
		if err != nil {
			return nil, derrors.NewInternalError(fmt.Errorf("failed to parse delay duration: %w", err))
		}

		t0 := time.Now().UTC()
		t := d.Shift(t0)

		err = logic.Sleep(ctx, t.Sub(t0), "")
		if err != nil {
			return nil, err
		}

		//nolint:nilnil
		return nil, nil
	}

	logic.Log(ctx, log.Info, "Waking up from sleep.")

	return &Transition{
		Transform: logic.Transform,
		NextState: logic.Transition,
	}, nil
}
