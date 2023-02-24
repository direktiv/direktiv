package states

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/model"
)

const d1s = "PT1S"

func TestDelayGood001(t *testing.T) {

	ctx := context.Background()

	instance := newTesterInstance()

	state := new(model.DelayState)
	state.Type = model.StateTypeDelay
	state.Transform = "a"
	state.Transition = "b"

	delay := time.Second * 1
	state.Duration = d1s

	logic, err := Delay(instance, state)
	if err != nil {
		t.Error(err)
		return
	}

	if time.Until(logic.Deadline(ctx)) < delay-time.Second {
		t.Error(errors.New("deadline too short"))
		return
	}

	transition, err := logic.Run(ctx, instance.getWakedata(), "", 0)
	if err != nil {
		t.Error(err)
		return
	}

	if transition != nil {
		t.Error(errors.New("expected null transition"))
		return
	}

	transition, err = logic.Run(ctx, instance.getWakedata(), "", 0)
	if err != nil {
		t.Error(err)
		return
	}

	if instance.dtCPU() > 200*time.Millisecond {
		t.Error(errors.New("ran longer than acceptable"))
		return
	}

	if instance.dt() < delay {
		t.Error(errors.New("didn't sleep for long enough"))
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
	if trace[0] != "Sleep" {
		t.Error(errors.New("didn't raise any errors"))
		return
	}

}

func TestDelayGood002(t *testing.T) {

	ctx := context.Background()

	instance := newTesterInstance()

	state := new(model.DelayState)
	state.Type = model.StateTypeDelay
	state.Transform = "a"
	state.Transition = "b"

	delay := time.Hour * 12
	state.Duration = "PT12H"

	logic, err := Delay(instance, state)
	if err != nil {
		t.Error(err)
		return
	}

	if time.Until(logic.Deadline(ctx)) < delay-time.Second {
		t.Error(errors.New("deadline too short"))
		return
	}

	transition, err := logic.Run(ctx, instance.getWakedata(), "", 0)
	if err != nil {
		t.Error(err)
		return
	}

	if transition != nil {
		t.Error(errors.New("expected null transition"))
		return
	}

	transition, err = logic.Run(ctx, instance.getWakedata(), "", 0)
	if err != nil {
		t.Error(err)
		return
	}

	if instance.dtCPU() > 200*time.Millisecond {
		t.Error(errors.New("ran longer than acceptable"))
		return
	}

	if instance.dt() < delay {
		t.Error(errors.New("didn't sleep for long enough"))
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
	if trace[0] != "Sleep" {
		t.Error(errors.New("didn't raise any errors"))
		return
	}

}

func TestDelayBadState(t *testing.T) {

	instance := newTesterInstance()

	state := new(model.NoopState)
	state.Type = model.StateTypeNoop
	state.Transform = "a"
	state.Transition = "b"

	_, err := Delay(instance, state)
	if err == nil {
		t.Error(errors.New("should have failed with wrong state type"))
		return
	}

}

func TestDelayBadMemory(t *testing.T) {

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

	state := new(model.DelayState)
	state.Type = model.StateTypeDelay
	state.Transform = "a"
	state.Transition = "b"

	state.Duration = d1s

	logic, err := Delay(instance, state)
	if err != nil {
		t.Error(err)
		return
	}

	transition, err := logic.Run(ctx, nil, "", 0)

	if instance.dtCPU() > 200*time.Millisecond {
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

	if instance.dt() > 200*time.Millisecond {
		t.Error(errors.New("scheduled a delay despite failing"))
		return
	}

}

func TestDelayBadWakedata(t *testing.T) {

	ctx := context.Background()

	instance := newTesterInstance()

	state := new(model.DelayState)
	state.Type = model.StateTypeDelay
	state.Transform = "a"
	state.Transition = "b"

	state.Duration = d1s

	logic, err := Delay(instance, state)
	if err != nil {
		t.Error(err)
		return
	}

	transition, err := logic.Run(ctx, marshal("bad"), "", 0)

	if instance.dtCPU() > 200*time.Millisecond {
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

	if instance.dt() > 200*time.Millisecond {
		t.Error(errors.New("scheduled a delay despite failing"))
		return
	}

}
