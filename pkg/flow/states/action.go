package states

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/senseyeio/duration"
)

func init() {
	RegisterState(model.StateTypeAction, Action)
}

type actionLogic struct {
	*model.ActionState
	Instance
}

// Action initializes the logic for executing an 'action' state in a Direktiv workflow instance.
func Action(instance Instance, state model.State) (Logic, error) {
	action, ok := state.(*model.ActionState)
	if !ok {
		return nil, derrors.NewInternalError(errors.New("bad state object"))
	}

	sl := new(actionLogic)
	sl.Instance = instance
	sl.ActionState = action

	return sl, nil
}

// Deadline overwrites the default underlying Deadline function provided by Instance because
// Action is a multi-step state.
func (logic *actionLogic) Deadline(ctx context.Context) time.Time {
	if logic.Async {
		return time.Now().Add(DefaultShortDeadline)
	}

	d, err := duration.ParseISO8601(logic.Timeout)
	if err != nil {
		if logic.Timeout != "" {
			logic.Log(ctx, "failed to parse timeout: %v", err)
		}
		return time.Now().Add(DefaultLongDeadline)
	}

	t := d.Shift(time.Now().Add(DefaultLongDeadline))

	return t
}

// Run implements the Run function for the Logic interface.
//
// The 'action' state ...
// To achieve this, the state must be scheduled in at least twice. The first time Run is called
// the state queues up the action and schedules a timeout for it. The second time Run is called
// should be in response to the action's completion. But it could also be because of the
// timeout. If the action times out or fails, the action logic may attempt to retry it, which
// means that the number of times this logic can run may vary.
func (logic *actionLogic) Run(ctx context.Context, wakedata []byte) (*Transition, error) {
	// first schedule
	if len(wakedata) == 0 {

		err := noMemory(logic)
		if err != nil {
			return nil, err
		}

		err = logic.scheduleFirstAction(ctx)
		if err != nil {
			return nil, err
		}

		if logic.Async {
			return &Transition{
				Transform: logic.Transform,
				NextState: logic.Transition,
			}, nil
		}

		return nil, nil

	}

	var children []*ChildInfo
	err := logic.UnmarshalMemory(&children)
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}

	// check if this is scheduled in for a retry
	var retry actionRetryInfo
	dec := json.NewDecoder(bytes.NewReader(wakedata))
	dec.DisallowUnknownFields()
	err = dec.Decode(&retry)
	if err == nil {
		return nil, logic.scheduleRetryAction(ctx, &retry)
	}

	// if we make it here, we've surely received action results
	var results actionResultPayload
	dec = json.NewDecoder(bytes.NewReader(wakedata))
	dec.DisallowUnknownFields()
	err = dec.Decode(&results)
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}

	return logic.processActionResults(ctx, children, &results)
}

func (logic *actionLogic) scheduleFirstAction(ctx context.Context) error {
	return logic.scheduleAction(ctx, 0)
}

func (logic *actionLogic) scheduleAction(ctx context.Context, attempt int) error {
	input, files, err := generateActionInput(ctx, &generateActionInputArgs{
		Instance: logic.Instance,
		Source:   logic.GetInstanceData(),
		Action:   logic.Action,
		Files:    logic.Action.Files,
	})
	if err != nil {
		return err
	}

	wfto, err := ISO8601StringtoSecs(logic.Timeout)
	if err != nil {
		return err
	}

	x, err := logic.GetModel()
	if err != nil {
		return derrors.NewInternalError(err)
	}

	fn, err := x.GetFunction(logic.Action.Function)
	if err != nil {
		return derrors.NewInternalError(err)
	}

	child, err := invokeAction(ctx, invokeActionArgs{
		instance: logic.Instance,
		async:    logic.Async,
		fn:       fn,
		input:    input,
		timeout:  wfto,
		files:    files,
		attempt:  attempt,
	})
	if err != nil {
		return err
	}

	if logic.Async {
		return nil
	}

	logic.Log(ctx, "Sleeping until child '%s' returns (%s).", child.ID, fn.GetID())

	var children []*ChildInfo

	children = append(children, child)

	err = logic.SetMemory(ctx, children)
	if err != nil {
		return err
	}

	return nil
}

func (logic *actionLogic) scheduleRetryAction(ctx context.Context, retry *actionRetryInfo) error {
	logic.Log(ctx, "Retrying...")

	err := logic.scheduleAction(ctx, retry.Children[retry.Idx].Attempts)
	if err != nil {
		return err
	}

	return nil
}

func (logic *actionLogic) processActionResults(ctx context.Context, children []*ChildInfo, results *actionResultPayload) (*Transition, error) {
	var err error

	sd := children[0]

	id := sd.ID

	if results.ActionID != id {
		return nil, derrors.NewInternalError(errors.New("incorrect child action ID"))
	}

	logic.Log(ctx, "Child '%s' returned.", id)

	if results.ErrorCode != "" {

		logic.Log(ctx, "Action raised catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)

		err = derrors.NewCatchableError(results.ErrorCode, results.ErrorMessage)
		d, err := preprocessRetry(logic.Action.Retries, sd.Attempts, err)
		if err != nil {
			return nil, err
		}

		logic.Log(ctx, "Scheduling retry attempt in: %v.", d)

		return nil, scheduleRetry(ctx, logic.Instance, children, 0, d)

	}

	if results.ErrorMessage != "" {
		logic.Log(ctx, "Action crashed due to an internal error: %v", results.ErrorMessage)
		return nil, derrors.NewInternalError(errors.New(results.ErrorMessage))
	}

	var x interface{}

	err = json.Unmarshal(results.Output, &x)
	if err != nil {
		x = base64.StdEncoding.EncodeToString(results.Output)
	}

	err = logic.StoreData("return", x)
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}

	return &Transition{
		Transform: logic.Transform,
		NextState: logic.Transition,
	}, nil
}
