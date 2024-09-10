package states

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	log "github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/senseyeio/duration"
)

const (
	foreachMaxThreads = 3
)

//nolint:gochecknoinits
func init() {
	RegisterState(model.StateTypeForEach, ForEach)
}

type forEachLogic struct {
	*model.ForEachState
	Instance
}

// ForEach initializes the logic for executing an 'action' state in a Direktiv workflow instance.
func ForEach(instance Instance, state model.State) (Logic, error) {
	forEach, ok := state.(*model.ForEachState)
	if !ok {
		return nil, derrors.NewInternalError(errors.New("bad state object"))
	}

	sl := new(forEachLogic)
	sl.Instance = instance
	sl.ForEachState = forEach

	return sl, nil
}

// Deadline overwrites the default underlying Deadline function provided by Instance because
// Action is a multi-step state.
func (logic *forEachLogic) Deadline(ctx context.Context) time.Time {
	d, err := duration.ParseISO8601(logic.Timeout)
	if err != nil {
		if logic.Timeout != "" {
			logic.Log(ctx, log.Error, "failed to parse timeout: %v", err)
		}

		return time.Now().UTC().Add(DefaultLongDeadline)
	}

	t := d.Shift(time.Now().UTC().Add(DefaultLongDeadline))

	return t
}

// Run implements the Run function for the Logic interface.
//
// The 'foreach' state ...
// To achieve this, the state must be scheduled in at least twice. The first time Run is called
// the state queues up the action and schedules a timeout for it. The second time Run is called
// should be in response to the action's completion. But it could also be because of the
// timeout. If the action times out or fails, the action logic may attempt to retry it, which
// means that the number of times this logic can run may vary.
func (logic *forEachLogic) Run(ctx context.Context, wakedata []byte) (*Transition, error) {
	// first schedule
	if len(wakedata) == 0 {
		err := noMemory(logic)
		if err != nil {
			return nil, err
		}

		transition, err := logic.scheduleFirstActions(ctx)
		if err != nil {
			return nil, err
		}

		return transition, nil
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

func (logic *forEachLogic) scheduleFirstActions(ctx context.Context) (*Transition, error) {
	x, err := jqOne(logic.GetInstanceData(), logic.Array) //nolint:contextcheck
	if err != nil {
		return nil, err
	}

	var array []interface{}
	array, ok := x.([]interface{})
	if !ok {
		return nil, derrors.NewCatchableError(ErrCodeNotArray, "jq produced non-array output")
	}

	if len(array) == 0 {
		return &Transition{
			Transform: logic.Transform,
			NextState: logic.Transition,
		}, nil
	}

	logic.Log(ctx, log.Info, "Generated %d objects to loop over.", len(array))

	children := make([]*ChildInfo, 0)

	for idx, inputSource := range array {
		if idx < foreachMaxThreads {
			child, err := logic.scheduleAction(ctx, inputSource, 0, idx)
			if err != nil {
				return nil, err
			}
			children = append(children, child)
		} else {
			children = append(children, nil)
		}
	}

	err = logic.SetMemory(ctx, children)
	if err != nil {
		return nil, err
	}

	//nolint:nilnil
	return nil, nil
}

func (logic *forEachLogic) scheduleAction(ctx context.Context, inputSource interface{}, attempt, iterator int) (*ChildInfo, error) {
	action := logic.Action

	input, files, err := generateActionInput(ctx, &generateActionInputArgs{
		Instance: logic.Instance,
		Source:   inputSource,
		Action:   action,
		Files:    action.Files,
	})
	if err != nil {
		return nil, err
	}

	wfto, err := ISO8601StringtoSecs(logic.Timeout)
	if err != nil {
		return nil, err
	}

	x, err := logic.GetModel()
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}

	fn, err := x.GetFunction(action.Function)
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}

	child, err := invokeAction(ctx, invokeActionArgs{
		instance: logic.Instance,
		async:    false,
		fn:       fn,
		input:    input,
		timeout:  wfto,
		files:    files,
		attempt:  attempt,
		iterator: iterator,
	})
	if err != nil {
		return nil, err
	}

	return child, nil
}

func (logic *forEachLogic) scheduleRetryAction(ctx context.Context, retry *actionRetryInfo) error {
	logic.Log(ctx, log.Info, "Retrying...")

	x, err := jqOne(logic.GetInstanceData(), logic.Array) //nolint:contextcheck
	if err != nil {
		return err
	}

	var array []interface{}
	array, ok := x.([]interface{})
	if !ok {
		return derrors.NewCatchableError(ErrCodeNotArray, "jq produced non-array output")
	}

	child, err := logic.scheduleAction(ctx, array[retry.Idx], retry.Children[retry.Idx].Attempts, retry.Iterator)
	if err != nil {
		return err
	}

	children := make([]*ChildInfo, 0)
	err = logic.UnmarshalMemory(&children)
	if err != nil {
		return err
	}

	children[retry.Idx] = child

	err = logic.SetMemory(ctx, children)
	if err != nil {
		return err
	}

	return nil
}

//nolint:gocognit
func (logic *forEachLogic) processActionResults(ctx context.Context, children []*ChildInfo, results *actionResultPayload) (*Transition, error) {
	var err error

	var found bool
	var idx int
	var completed int

	for i, lid := range children {
		if lid == nil {
			continue
		}

		if lid.ID == results.ActionID {
			found = true
			if lid.Complete {
				return nil, derrors.NewInternalError(fmt.Errorf("action '%s' already completed", lid.ID))
			}
			idx = i
		}

		if lid.Complete {
			completed++
		}
	}

	if !found {
		return nil, derrors.NewInternalError(fmt.Errorf("action '%s' wasn't expected", results.ActionID))
	}

	sd := children[idx]

	id := sd.ID

	if results.ActionID != id {
		return nil, derrors.NewInternalError(errors.New("incorrect child action ID"))
	}
	logic.AddAttribute("loop-index", fmt.Sprintf("%d", idx))
	ctx = tracing.AddTag(ctx, "branch", idx)
	ctx, end, err := tracing.NewSpan(ctx, "processing action results")
	if err != nil {
		slog.Debug("tracing.NewSpan failed in processActionResults", "error", "err")
	}
	defer end()
	logic.Log(ctx, log.Debug, "Child '%s' returned.", id)

	if results.ErrorCode != "" {
		logic.Log(ctx, log.Error, "[%v] Action raised catchable error '%s': %s.", idx, results.ErrorCode, results.ErrorMessage)

		err = derrors.NewCatchableError(results.ErrorCode, results.ErrorMessage)
		d, err := preprocessRetry(logic.Action.Retries, sd.Attempts, err)
		if err != nil {
			return nil, err
		}

		logic.Log(ctx, log.Info, "[%v] Scheduling retry attempt in: %v.", idx, d)

		return nil, scheduleRetry(ctx, logic.Instance, children, idx, d)
	}

	if results.ErrorMessage != "" {
		logic.Log(ctx, log.Error, "Action crashed due to an internal error: %v", results.ErrorMessage)
		return nil, derrors.NewInternalError(errors.New(results.ErrorMessage))
	}

	children[idx].Complete = true
	completed++
	logic.Log(ctx, log.Info, "[%v] Action returned. (%d/%d)", idx, completed, len(children))

	var x interface{}

	err = json.Unmarshal(results.Output, &x)
	if err != nil {
		x = base64.StdEncoding.EncodeToString(results.Output)
	}

	children[idx].Results = x

	var ready bool
	if completed == len(children) {
		ready = true
	}

	if ready {
		var results []interface{}
		for i := range children {
			results = append(results, children[i].Results)
		}

		err = logic.StoreData("return", results)
		if err != nil {
			return nil, derrors.NewInternalError(err)
		}

		return &Transition{
			Transform: logic.Transform,
			NextState: logic.Transition,
		}, nil
	}

	idx = -1
	var ci *ChildInfo
	for i, child := range children {
		if child == nil {
			idx = i

			x, err := jqOne(logic.GetInstanceData(), logic.Array) //nolint:contextcheck
			if err != nil {
				return nil, err
			}

			var array []interface{}
			array, ok := x.([]interface{})
			if !ok {
				return nil, derrors.NewCatchableError(ErrCodeNotArray, "jq produced non-array output")
			}

			ci, err = logic.scheduleAction(ctx, array[idx], 0, idx)
			if err != nil {
				return nil, err
			}

			break
		}
	}
	if idx >= 0 {
		children[idx] = ci
	}

	err = logic.SetMemory(ctx, children)
	if err != nil {
		return nil, err
	}

	//nolint:nilnil
	return nil, nil
}
