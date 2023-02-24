package states

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/model"
)

func TestSwitchGood001(t *testing.T) {
	ctx := context.Background()

	instance := newTesterInstance()

	state := new(model.SwitchState)
	state.Type = model.StateTypeSwitch
	state.DefaultTransform = "a"
	state.DefaultTransition = "b"
	state.Conditions = append(state.Conditions, model.SwitchConditionDefinition{
		Condition:  false,
		Transform:  "c",
		Transition: "d",
	})
	state.Conditions = append(state.Conditions, model.SwitchConditionDefinition{
		Condition:  false,
		Transform:  "e",
		Transition: "f",
	})

	logic, err := Switch(instance, state)
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

	if transition.Transform != state.DefaultTransform {
		t.Error(errors.New("incorrect transform"))
		return
	}

	if transition.NextState != state.DefaultTransition {
		t.Error(errors.New("incorrect transition"))
		return
	}

	trace := instance.getTraceExclude()
	if len(trace) > 0 {
		t.Error(errors.New("state did more than it should"))
		return
	}
}

func TestSwitchGood002(t *testing.T) {
	ctx := context.Background()

	instance := newTesterInstance()

	state := new(model.SwitchState)
	state.Type = model.StateTypeSwitch
	state.DefaultTransform = "a"
	state.DefaultTransition = "b"
	state.Conditions = append(state.Conditions, model.SwitchConditionDefinition{
		Condition:  false,
		Transform:  "c",
		Transition: "d",
	})
	state.Conditions = append(state.Conditions, model.SwitchConditionDefinition{
		Condition:  true,
		Transform:  "e",
		Transition: "f",
	})

	logic, err := Switch(instance, state)
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

	if transition.Transform != state.Conditions[1].Transform {
		t.Error(errors.New("incorrect transform"))
		return
	}

	if transition.NextState != state.Conditions[1].Transition {
		t.Error(errors.New("incorrect transition"))
		return
	}

	trace := instance.getTraceExclude()
	if len(trace) > 0 {
		t.Error(errors.New("state did more than it should"))
		return
	}
}

func TestSwitchBadState(t *testing.T) {
	instance := newTesterInstance()

	state := new(model.NoopState)
	state.Type = model.StateTypeNoop

	_, err := Switch(instance, state)
	if err == nil {
		t.Error(errors.New("should have failed with wrong state type"))
		return
	}
}

func TestSwitchBadMemory(t *testing.T) {
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

	state := new(model.SwitchState)
	state.Type = model.StateTypeSwitch

	logic, err := Switch(instance, state)
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

func TestSwitchBadWakedata(t *testing.T) {
	ctx := context.Background()

	instance := newTesterInstance()

	state := new(model.SwitchState)
	state.Type = model.StateTypeSwitch

	logic, err := Switch(instance, state)
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
