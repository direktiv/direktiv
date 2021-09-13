package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/senseyeio/duration"

	"github.com/vorteil/direktiv/pkg/model"
)

type engine struct {
	*server
	cancellers     map[string]func()
	cancellersLock sync.Mutex
	stateLogics    map[model.StateType]func(*model.Workflow, model.State) (stateLogic, error)
}

func initEngine(srv *server) (*engine, error) {

	engine := new(engine)

	engine.server = srv

	engine.cancellers = make(map[string]func())

	engine.stateLogics = map[model.StateType]func(*model.Workflow, model.State) (stateLogic, error){
		model.StateTypeNoop: initNoopStateLogic,
		// model.StateTypeAction:        initActionStateLogic,
		// model.StateTypeConsumeEvent:  initConsumeEventStateLogic,
		model.StateTypeDelay: initDelayStateLogic,
		model.StateTypeError: initErrorStateLogic,
		// model.StateTypeEventsAnd:     initEventsAndStateLogic,
		// model.StateTypeEventsXor:     initEventsXorStateLogic,
		// model.StateTypeForEach:       initForEachStateLogic,
		// model.StateTypeGenerateEvent: initGenerateEventStateLogic,
		// model.StateTypeParallel:      initParallelStateLogic,
		model.StateTypeSwitch:   initSwitchStateLogic,
		model.StateTypeValidate: initValidateStateLogic,
		model.StateTypeGetter:   initGetterStateLogic,
		model.StateTypeSetter:   initSetterStateLogic,
	}

	return engine, nil

}

func (engine *engine) Close() error {

	return nil

}

type newInstanceArgs struct {
	Namespace  string
	Path       string
	Ref        string
	Input      []byte
	Caller     string
	CallerData string
}

type subflowCaller struct {
	InstanceID string
	State      string
	Step       int
	Depth      int
}

func (engine *engine) NewInstance(ctx context.Context, args *newInstanceArgs) (*instanceMemory, error) {

	tx, err := engine.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace

	d, err := engine.mux(ctx, nsc, args.Namespace, args.Path, args.Ref)
	if err != nil {
		return nil, err
	}

	inc := tx.Instance
	rtc := tx.InstanceRuntime

	as := args.Path
	if args.Ref != "" {
		as += ":" + args.Ref
	}

	data := marshalInstanceInputData(args.Input)

	// SetFlow()
	rt, err := rtc.Create().SetInput(args.Input).SetData(data).SetMemory("null").SetCallerData(args.CallerData).Save(ctx)
	if err != nil {
		return nil, err
	}

	in, err := inc.Create().SetNamespace(d.ns()).SetWorkflow(d.wf).SetRevision(d.rev()).SetRuntime(rt).SetStatus(StatusPending).SetAs(as).Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	in.Edges.Namespace = d.ns()
	in.Edges.Revision = d.rev()
	in.Edges.Runtime = rt
	in.Edges.Workflow = d.wf

	rt.Edges.Instance = in

	im := new(instanceMemory)
	im.in = in

	err = json.Unmarshal([]byte(im.in.Edges.Runtime.Data), &im.data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(im.in.Edges.Runtime.Memory), &im.memory)
	if err != nil {
		return nil, err
	}

	im.Unwrap()

	t := time.Now()
	engine.pubsub.NotifyInstances(d.ns())
	engine.logToNamespace(ctx, t, d.ns(), "Workflow '%s' has been triggered by %s.", args.Path, args.Caller)
	engine.logToWorkflow(ctx, t, d.wf, "Instance '%s' created by %s.", im.ID().String(), args.Caller)
	engine.logToInstance(ctx, t, in, "Preparing workflow triggered by %s.", args.Caller)

	return im, nil

}

func (engine *engine) start(im *instanceMemory) {

	ctx, err := engine.InstanceLock(im, time.Second*defaultLockWait)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	engine.sugar.Debugf("Starting workflow %v", im.ID().String())

	workflow, err := im.Model()
	if err != nil {
		engine.CrashInstance(ctx, im, NewUncatchableError(ErrCodeWorkflowUnparsable, "failed to parse workflow YAML: %v", err))
		return
	}

	start := workflow.GetStartState()

	engine.Transition(ctx, im, start.GetID(), 0)

}

func (engine *engine) Transition(ctx context.Context, im *instanceMemory, nextState string, attempt int) {

	workflow, err := im.Model()
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return
	}

	oldController := im.Controller()

	if im.Step() == 0 {

		t := time.Now()
		tSoft := time.Now().Add(time.Minute * 15)
		tHard := time.Now().Add(time.Minute * 20)

		if workflow.Timeouts != nil {

			s := workflow.Timeouts.Interrupt

			if s != "" {
				d, err := duration.ParseISO8601(s)
				if err != nil {
					engine.CrashInstance(ctx, im, err)
					return
				}
				tSoft = d.Shift(t)
				tHard = tSoft.Add(time.Minute * 5)
			}

			s = workflow.Timeouts.Kill

			if s != "" {
				d, err := duration.ParseISO8601(s)
				if err != nil {
					engine.CrashInstance(ctx, im, err)
					return
				}
				tHard = d.Shift(t)
			}

		}

		engine.ScheduleSoftTimeout(im, oldController, tSoft)
		engine.ScheduleHardTimeout(im, oldController, tHard)

	}

	// TODO: I don't think this is actually possible?
	// if len(wli.rec.Flow) != wli.step {
	// 	err := errors.New("workflow logic instance aborted for being tardy")
	// 	appLog.Error(err)
	// 	wli.Close()
	// 	return
	// }

	if nextState == "" {
		panic("don't call this function with an empty nextState")
	}

	states := workflow.GetStatesMap()
	state, exists := states[nextState]
	if !exists {
		err = fmt.Errorf("workflow cannot resolve transition: %s", nextState)
		engine.CrashInstance(ctx, im, err)
		return
	}

	init, exists := engine.stateLogics[state.GetType()]
	if !exists {
		err = fmt.Errorf("engine cannot resolve state type: %s", state.GetType().String())
		engine.CrashInstance(ctx, im, err)
		return
	}

	stateLogic, err := init(workflow, state)
	if err != nil {
		err = fmt.Errorf("cannot initialize state logic: %v", err)
		engine.CrashInstance(ctx, im, err)
		return
	}
	im.logic = stateLogic

	flow := append(im.Flow(), nextState)
	deadline := stateLogic.Deadline(ctx, engine, im)

	rt := im.in.Edges.Runtime
	edges := rt.Edges

	im.SetMemory(nil)

	rt, err = rt.Update().
		SetFlow(flow).
		SetController(engine.pubsub.hostname).
		SetAttempts(attempt).
		SetDeadline(deadline).
		SetStateBeginTime(time.Now()).
		SetData(im.MarshalData()).
		SetMemory(im.MarshalMemory()).
		Save(ctx)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return
	}

	engine.pubsub.NotifyInstance(im.in)

	rt.Edges = edges
	im.in.Edges.Runtime = rt

	engine.ScheduleSoftTimeout(im, oldController, deadline)

	engine.runState(ctx, im, nil, nil)

}

func (engine *engine) CrashInstance(ctx context.Context, im *instanceMemory, err error) {

	engine.sugar.Debugf("Instance failed with uncatchable error: %v", err)

	engine.logToInstance(ctx, time.Now(), im.in, "Instance failed with uncatchable error: %s", err.Error())

	err = engine.SetInstanceFailed(ctx, im, err)
	if err != nil {
		engine.sugar.Error(err)
	}

	engine.TerminateInstance(ctx, im)

	engine.FreeInstanceMemory(im)

}

func (engine *engine) TerminateInstance(ctx context.Context, im *instanceMemory) {

	engine.freeResources(im)
	engine.WakeInstanceCaller(ctx, im)
	engine.metricsCompleteState(ctx, im, "", im.ErrorCode(), false)
	engine.metricsCompleteInstance(ctx, im)

}

func (engine *engine) runState(ctx context.Context, im *instanceMemory, wakedata []byte, err error) {

	engine.logRunState(ctx, im, wakedata, err)

	var code string
	var transition *stateTransition

	if err != nil {
		goto failure
	}

	if lq := im.logic.LogJQ(); im.GetMemory() == nil && len(wakedata) == 0 && lq != nil {
		var object interface{}
		object, err = jqOne(im.data, lq)
		if err != nil {
			goto failure
		}

		var data []byte
		data, err = json.MarshalIndent(object, "", "  ")
		if err != nil {
			err = NewInternalError(fmt.Errorf("failed to marshal state data: %w", err))
			goto failure
		}

		engine.UserLog(ctx, im, string(data))
	}

	transition, err = im.logic.Run(ctx, engine, im, wakedata)
	if err != nil {
		goto failure
	}

	err = engine.transformState(ctx, im, transition)
	if err != nil {
		goto failure
	}

next:
	engine.transitionState(ctx, im, transition, code)
	return

failure:

	var breaker int

	if breaker > 10 {
		err = NewInternalError(errors.New("somehow ended up in a catchable error loop"))
	}

	engine.CancelInstanceChildren(ctx, im)

	if cerr, ok := err.(*CatchableError); ok {

		_ = im.StoreData("error", cerr)

		for i, catch := range im.logic.ErrorCatchers() {

			errRegex := catch.Error
			if errRegex == "*" {
				errRegex = ".*"
			}

			t := time.Now()

			var matched bool
			matched, err = regexp.MatchString(errRegex, cerr.Code)
			if err != nil {
				engine.logToInstance(ctx, t, im.in, "Error catching regex failed to compile: %v", err)
			}

			if matched {

				engine.logToInstance(ctx, t, im.in, "State failed with error '%s': %s", cerr.Code, cerr.Message)
				engine.logToInstance(ctx, t, im.in, "Error caught by error definition %d: %s", i, catch.Error)

				transition = &stateTransition{
					Transform: "",
					NextState: catch.Transition,
				}

				breaker++

				code = cerr.Code

				goto next

			}

		}

	}

	engine.CrashInstance(ctx, im, err)

}

func (engine *engine) transformState(ctx context.Context, im *instanceMemory, transition *stateTransition) error {

	if transition == nil || transition.Transform == nil {
		return nil
	}

	if s, ok := transition.Transform.(string); ok && (s == "" || s == ".") {
		return nil
	}

	engine.logToInstance(ctx, time.Now(), im.in, "Transforming state data.")

	x, err := jqObject(im.data, transition.Transform)
	if err != nil {
		return WrapCatchableError("unable to apply transform: %v", err)
	}

	im.data = x

	return nil

}

func (engine *engine) transitionState(ctx context.Context, im *instanceMemory, transition *stateTransition, errCode string) {

	if transition == nil || transition.NextState == "" {

		// TODO

		// ns, err := engine.InstanceNamespace(ctx, im)
		// if err != nil {
		// 	engine.CrashInstance(ctx, im, err)
		// 	return
		// }
		//
		// for i := range im.eventQueue {
		// 	we.server.flushEvent(im.eventQueue[i], ns.ID, true)
		// }

	}

	if transition == nil {
		engine.InstanceYield(im)
		return
	}

	engine.metricsCompleteState(ctx, im, transition.NextState, errCode, false)

	if transition.NextState != "" {
		engine.sugar.Debugf("Instance transitioning to next state: %s -> %s", im.ID().String(), transition.NextState)
		engine.logToInstance(ctx, time.Now(), im.in, "Transitioning to next state: %s (%d).", transition.NextState, im.Step()+1)
		go engine.Transition(ctx, im, transition.NextState, 0)
		return
	}

	status := StatusComplete
	if im.ErrorCode() != "" {
		status = StatusFailed
		engine.sugar.Debugf("Instance failed: %s", im.ID().String())
		engine.logToInstance(ctx, time.Now(), im.in, "Workflow failed with error '%s': %s", im.ErrorCode(), im.in.ErrorMessage)
	}

	engine.sugar.Debugf("Instance terminated: %s", im.ID().String())

	engine.metricsCompleteInstance(ctx, im)

	var err error

	rt := im.in.Edges.Runtime
	rte := rt.Edges
	rt, err = rt.Update().SetOutput(im.MarshalData()).Save(ctx)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return
	}
	rt.Edges = rte
	im.in.Edges.Runtime = rt

	in := im.in
	ine := in.Edges
	in, err = in.Update().SetStatus(status).Save(ctx)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return
	}
	in.Edges = ine
	im.in = in

	engine.pubsub.NotifyInstance(im.in)

	engine.logToInstance(ctx, time.Now(), im.in, "Workflow completed.")

	engine.TerminateInstance(ctx, im)

}
