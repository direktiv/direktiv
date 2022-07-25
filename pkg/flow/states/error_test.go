package states

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/model"
)

func TestErrorGood(t *testing.T) {

	ctx := context.Background()

	instance := newTesterInstance()

	state := new(model.ErrorState)
	state.Type = model.StateTypeError
	state.Transform = "a"
	state.Transition = "b"
	state.Error = "thing went wrong: %v"
	state.Args = []string{"5"}

	logic, err := Error(instance, state)
	if err != nil {
		t.Error(err)
		return
	}

	if logic.Deadline(ctx).Sub(time.Now()) > DefaultShortDeadline {
		t.Error(errors.New("deadline too long"))
		return
	}

	transition, err := logic.Run(ctx, instance.getWakedata())
	if err != nil {
		t.Error(err)
		return
	}

	if instance.dt() > 200*time.Millisecond {
		t.Error(errors.New("ran longer than acceptable"))
		return
	}

	if transition == nil {
		t.Error(errors.New("expected non-null transition"))
		return
	}

	if transition.Transform != state.Transform {
		t.Error(errors.New("incorrect transform"))
		return
	}

	if transition.NextState != state.Transition {
		t.Error(errors.New("incorrect transition"))
		return
	}

	trace := instance.getTraceExclude()
	if len(trace) == 0 {
		t.Error(errors.New("state did nothing"))
		return
	}
	if len(trace) > 1 {
		t.Error(errors.New("state did more than it should"))
		return
	}
	if trace[0] != "Raise" {
		t.Error(errors.New("didn't raise any errors"))
		return
	}

}

func TestErrorBadState(t *testing.T) {

	instance := newTesterInstance()

	state := new(model.NoopState)
	state.Type = model.StateTypeNoop

	_, err := Error(instance, state)
	if err == nil {
		t.Error(errors.New("should have failed with wrong state type"))
		return
	}

}

func TestErrorBadMemory(t *testing.T) {

	ctx := context.Background()

	instance := newTesterInstance()
	instance.SetMemory(ctx, map[string]int{
		"a": 5,
	})
	instance.resetTrace()

	state := new(model.ErrorState)
	state.Type = model.StateTypeError

	logic, err := Error(instance, state)
	if err != nil {
		t.Error(err)
		return
	}

	transition, err := logic.Run(ctx, instance.getWakedata())

	if instance.dt() > 200*time.Millisecond {
		t.Error(errors.New("ran longer than acceptable"))
		return
	}

	if transition != nil {
		t.Error(errors.New("expected null transition"))
		return
	}

	if err == nil {
		t.Error(errors.New("expected an error"))
		return
	}

}

func TestErrorBadWakedata(t *testing.T) {

	ctx := context.Background()

	instance := newTesterInstance()

	state := new(model.ErrorState)
	state.Type = model.StateTypeError

	logic, err := Error(instance, state)
	if err != nil {
		t.Error(err)
		return
	}

	transition, err := logic.Run(ctx, marshal("bad"))

	if instance.dt() > 200*time.Millisecond {
		t.Error(errors.New("ran longer than acceptable"))
		return
	}

	if transition != nil {
		t.Error(errors.New("expected null transition"))
		return
	}

	if err == nil {
		t.Error(errors.New("expected an error"))
		return
	}

}
