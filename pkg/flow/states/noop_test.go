package states

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/model"
)

func TestNoopGood(t *testing.T) {
	ctx := context.Background()

	instance := newTesterInstance()

	state := new(model.NoopState)
	state.Type = model.StateTypeNoop
	state.Transform = "a"
	state.Transition = "b"

	logic, err := Noop(instance, state)
	if err != nil {
		t.Error(err)
		return
	}

	if time.Until(logic.Deadline(ctx)) > DefaultShortDeadline {
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
	if len(trace) > 0 {
		t.Error(errors.New("state did more than it should"))
		return
	}
}

func TestNoopBadState(t *testing.T) {
	instance := newTesterInstance()

	state := new(model.DelayState)
	state.Type = model.StateTypeDelay
	state.Transform = "a"
	state.Transition = "b"

	_, err := Noop(instance, state)
	if err == nil {
		t.Error(errors.New("should have failed with wrong state type"))
		return
	}
}

func TestNoopBadMemory(t *testing.T) {
	ctx := context.Background()

	instance := newTesterInstance()
	err := instance.SetMemory(ctx, map[string]int{
		"a": 5,
	})
	if err != nil {
		t.Error(err)
		return
	}

	instance.resetTrace()

	state := new(model.NoopState)
	state.Type = model.StateTypeNoop
	state.Transform = "a"
	state.Transition = "b"

	logic, err := Noop(instance, state)
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

func TestNoopBadWakedata(t *testing.T) {
	ctx := context.Background()

	instance := newTesterInstance()

	state := new(model.NoopState)
	state.Type = model.StateTypeNoop
	state.Transform = "a"
	state.Transition = "b"

	logic, err := Noop(instance, state)
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
