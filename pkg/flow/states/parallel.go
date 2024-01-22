package states

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/senseyeio/duration"
)

func init() {
	RegisterState(model.StateTypeParallel, Parallel)
}

type parallelLogic struct {
	*model.ParallelState
	Instance
}

// Parallel initializes the logic for executing a 'parallel' state in a Direktiv workflow instance.
func Parallel(instance Instance, state model.State) (Logic, error) {
	parallel, ok := state.(*model.ParallelState)
	if !ok {
		return nil, derrors.NewInternalError(errors.New("bad state object"))
	}

	sl := new(parallelLogic)
	sl.Instance = instance
	sl.ParallelState = parallel

	return sl, nil
}

// Deadline overwrites the default underlying Deadline function provided by Instance because
// Parallel is a multi-step state.
func (logic *parallelLogic) Deadline(ctx context.Context) time.Time {
	d, err := duration.ParseISO8601(logic.Timeout)
	if err != nil {
		if logic.Timeout != "" {
			// logic.Log(ctx, log.Error, "failed to parse timeout: %v", err)
		}
		return time.Now().UTC().Add(DefaultLongDeadline)
	}

	t := d.Shift(time.Now().UTC().Add(DefaultLongDeadline))

	return t
}

// Run implements the Run function for the Logic interface.
//
// The 'parallel' state ...
// To achieve this, the state must be scheduled in at least twice. The first time Run is called
// the state queues up the action and schedules a timeout for it. The second time Run is called
// should be in response to the action's completion. But it could also be because of the
// timeout. If the action times out or fails, the action logic may attempt to retry it, which
// means that the number of times this logic can run may vary.
func (logic *parallelLogic) Run(ctx context.Context, wakedata []byte) (*Transition, error) {
	// first schedule
	if len(wakedata) == 0 {
		err := noMemory(logic)
		if err != nil {
			return nil, err
		}

		err = logic.scheduleFirstActions(ctx)
		if err != nil {
			return nil, err
		}

		//nolint:nilnil
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

func (logic *parallelLogic) scheduleFirstActions(ctx context.Context) error {
	children := make([]*ChildInfo, 0)

	for i := range logic.Actions {
		action := &logic.Actions[i]

		child, err := logic.scheduleAction(ctx, action, 0)
		if err != nil {
			return err
		}

		children = append(children, child)
	}

	// logic.Log(ctx, log.Info, "Sleeping until children return.")

	err := logic.SetMemory(ctx, children)
	if err != nil {
		return err
	}

	return nil
}

func (logic *parallelLogic) scheduleAction(ctx context.Context, action *model.ActionDefinition, attempt int) (*ChildInfo, error) {
	input, files, err := generateActionInput(ctx, &generateActionInputArgs{
		Instance: logic.Instance,
		Source:   logic.GetInstanceData(),
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
	})
	if err != nil {
		return nil, err
	}

	return child, nil
}

func (logic *parallelLogic) scheduleRetryAction(ctx context.Context, retry *actionRetryInfo) error {
	// logic.Log(ctx, log.Info, "Retrying...")

	action := &logic.Actions[retry.Idx]

	child, err := logic.scheduleAction(ctx, action, retry.Children[retry.Idx].Attempts)
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

func (logic *parallelLogic) processActionResults(ctx context.Context, children []*ChildInfo, results *actionResultPayload) (*Transition, error) {
	var err error

	var found bool
	var idx int
	var completed int

	for i, lid := range children {
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

	// logic.Log(ctx, "Child '%s' returned.", id)

	if results.ErrorCode != "" {
		// logic.Log(ctx, log.Error, "Action raised catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)

		err = derrors.NewCatchableError(results.ErrorCode, results.ErrorMessage)
		d, err := preprocessRetry(logic.Actions[idx].Retries, sd.Attempts, err)
		if err != nil {
			return nil, err
		}

		// logic.Log(ctx, log.Info, "Scheduling retry attempt in: %v.", d)

		return nil, scheduleRetry(ctx, logic.Instance, children, idx, d)
	}

	if results.ErrorMessage != "" {
		// logic.Log(ctx, log.Error, "Action crashed due to an internal error: %v", results.ErrorMessage)
		return nil, derrors.NewInternalError(errors.New(results.ErrorMessage))
	}

	var x interface{}

	err = json.Unmarshal(results.Output, &x)
	if err != nil {
		x = base64.StdEncoding.EncodeToString(results.Output)
	}

	children[idx].Results = x

	var ready bool

	switch logic.Mode {
	case model.BranchModeAnd:

		if results.ErrorCode != "" {
			// logic.Log(ctx, log.Error, "Action raised catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)

			err = derrors.NewCatchableError(results.ErrorCode, results.ErrorMessage)

			d, err := preprocessRetry(logic.Actions[idx].Retries, children[idx].Attempts, err)
			if err != nil {
				return nil, err
			}

			// logic.Log(ctx, log.Info, "Scheduling retry attempt in: %v.", d)

			err = scheduleRetry(ctx, logic.Instance, children, idx, d)
			if err != nil {
				return nil, err
			}

			//nolint:nilnil
			return nil, nil
		}

		if results.ErrorMessage != "" {
			return nil, derrors.NewInternalError(errors.New(results.ErrorMessage))
		}

		children[idx].Complete = true
		if children[idx].Complete {
			completed++
		}

		// logic.Log(ctx, log.Info, "Action returned. (%d/%d)", completed, len(children))

		if completed == len(children) {
			ready = true
		}

	case model.BranchModeOr:

		if results.ErrorCode != "" {
			// logic.Log(ctx, "Action raised catchable error '%s': %s.", results.ErrorCode, results.ErrorMessage)

			err = derrors.NewCatchableError(results.ErrorCode, results.ErrorMessage)

			d, err := preprocessRetry(logic.Actions[idx].Retries, children[idx].Attempts, err)
			if err == nil {
				err = scheduleRetry(ctx, logic.Instance, children, idx, d)
				if err != nil {
					return nil, err
				}
				//nolint:nilnil
				return nil, nil
			}
		} else if results.ErrorMessage != "" {
			// logic.Log(ctx, log.Error, "Branch %d crashed due to an internal error: %s", idx, results.ErrorMessage)

			err = derrors.NewInternalError(errors.New(results.ErrorMessage))
			if err != nil {
				return nil, err
			}

			//nolint:nilnil
			return nil, nil
		} else {
			ready = true
		}

		children[idx].Complete = true
		completed++

		// logic.Log(ctx, log.Info, "Action returned. (%d/%d)", completed, len(children))

		if !ready && completed == len(children) {
			err = derrors.NewCatchableError(ErrCodeAllBranchesFailed, "all branches failed")
			return nil, err
		}

	default:
		return nil, derrors.NewInternalError(errors.New("unrecognized branch mode"))
	}

	if !ready {
		return nil, logic.SetMemory(ctx, children)
	}

	var finalResults []interface{}

	for i := range children {
		finalResults = append(finalResults, children[i].Results)
	}

	err = logic.StoreData("return", finalResults)
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}

	return &Transition{
		Transform: logic.Transform,
		NextState: logic.Transition,
	}, nil
}
