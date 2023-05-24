package flow

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/functions"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"github.com/senseyeio/duration"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type engine struct {
	*server
	cancellers     map[string]func()
	cancellersLock sync.Mutex
}

func initEngine(srv *server) (*engine, error) {
	engine := new(engine)

	engine.server = srv

	engine.cancellers = make(map[string]func())

	return engine, nil
}

func (engine *engine) Close() error {
	return nil
}

type newInstanceArgs struct {
	Namespace   string
	Path        string
	Ref         string
	Input       []byte
	Caller      string
	CallerData  string
	CallPath    string
	CallerState string
	Iterator    string
}

type subflowCaller struct {
	InstanceID  uuid.UUID
	State       string
	Step        int
	Depth       int
	As          string
	CallPath    string
	CallerState string
	Iterator    string
}

const (
	apiCaller = "api"
)

func unmarshalInstanceInputData(input []byte) interface{} {
	var inputData, stateData interface{}

	err := json.Unmarshal(input, &inputData)
	if err != nil {
		inputData = base64.StdEncoding.EncodeToString(input)
	}

	if _, ok := inputData.(map[string]interface{}); ok {
		stateData = inputData
	} else {
		stateData = map[string]interface{}{
			"input": inputData,
		}
	}

	return stateData
}

func marshalInstanceInputData(input []byte) string {
	x := unmarshalInstanceInputData(input)

	data, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (engine *engine) NewInstance(ctx context.Context, args *newInstanceArgs) (*instanceMemory, error) {
	tctx, tx, err := engine.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := engine.mux(tctx, args.Namespace, args.Path, args.Ref)
	if err != nil {
		engine.sugar.Debugf("Failed to create new instance: %v", err)
		engine.logger.Errorf(ctx, engine.flow.ID, engine.flow.GetAttributes(), "Failed to receive workflow %s", args.Path)
		if derrors.IsNotFound(err) {
			return nil, derrors.NewUncatchableError("direktiv.workflow.notfound", "workflow not found: %v", err.Error())
		}
		return nil, err
	}

	var wf model.Workflow
	err = wf.Load(cached.Revision.Data)
	if err != nil {
		engine.logger.Errorf(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Cannot parse workflow source %s", args.Path)
		return nil, derrors.NewUncatchableError("direktiv.workflow.invalid", "cannot parse workflow '%s': %v", args.Path, err)
	}

	if len(wf.GetStartDefinition().GetEvents()) > 0 {
		if strings.ToLower(args.Caller) == apiCaller {
			engine.logger.Errorf(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Cannot manually invoke event-based workflow %s", args.Path)
			return nil, derrors.NewUncatchableError("direktiv.workflow.invoke", "cannot manually invoke event-based workflow")
		}
		if strings.HasPrefix(args.Caller, "instance") {
			engine.logger.Errorf(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Cannot invoke event-based workflow %s as a subflow", args.Path)
			return nil, derrors.NewUncatchableError("direktiv.workflow.invoke", "cannot invoke event-based workflow as a subflow")
		}
	}

	as := args.Path
	if args.Ref != "" {
		as += ":" + args.Ref
	}
	callerInstanceID := ""
	if strings.HasPrefix(args.Caller, "instance:") {
		callerInstanceID = strings.Split(args.Caller, ":")[1]
	}
	callpath := internallogger.AppendInstanceID(args.CallPath, callerInstanceID)
	data := marshalInstanceInputData(args.Input)

	clients := engine.edb.Clients(tctx)

	// TODO: alan, need to add logToEvents value here
	rt, err := clients.InstanceRuntime.Create().SetInput(args.Input).SetData(data).SetMemory("null").SetCallerData(args.CallerData).Save(tctx)
	if err != nil {
		return nil, err
	}

	inst, err := clients.Instance.Create().SetNamespaceID(cached.Namespace.ID).SetWorkflowID(cached.File.ID).SetRevisionID(cached.Revision.ID).SetRuntime(rt).SetStatus(util.InstanceStatusPending).SetInvoker(args.Caller).SetAs(util.SanitizeAsField(as)).SetCallpath(callpath).SetInvokerState(args.CallerState).Save(tctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	runtime := &database.InstanceRuntime{
		ID:              rt.ID,
		Input:           rt.Input,
		Data:            rt.Data,
		Controller:      rt.Controller,
		Memory:          rt.Memory,
		Flow:            rt.Flow,
		Output:          rt.Output,
		StateBeginTime:  rt.StateBeginTime,
		Deadline:        rt.Deadline,
		Attempts:        rt.Attempts,
		CallerData:      rt.CallerData,
		InstanceContext: rt.InstanceContext,
		StateContext:    rt.StateContext,
		Metadata:        rt.Metadata,
		LogToEvents:     rt.LogToEvents,
	}

	cached.Instance = &database.Instance{
		ID:           inst.ID,
		CreatedAt:    inst.CreatedAt,
		UpdatedAt:    inst.UpdatedAt,
		EndAt:        inst.EndAt,
		Status:       inst.Status,
		As:           inst.As,
		ErrorCode:    inst.ErrorCode,
		ErrorMessage: inst.ErrorMessage,
		Invoker:      inst.Invoker,
		Namespace:    cached.Namespace.ID,
		Workflow:     cached.File.ID,
		Revision:     cached.Revision.ID,
		Runtime:      runtime.ID,
		CallPath:     inst.Callpath,
		InvokerState: inst.InvokerState,
	}

	im := new(instanceMemory)
	im.engine = engine
	im.cached = cached
	im.runtime = runtime

	err = im.engine.database.FlushInstance(ctx, im.cached.Instance)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(im.runtime.Data), &im.data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(im.runtime.Memory), &im.memory)
	if err != nil {
		return nil, err
	}

	// im.tx = nil

	ctx, err = traceFullAddWorkflowInstance(ctx, im)
	if err != nil {
		return nil, err
	}
	im.AddAttribute("loop-index", args.Iterator)

	engine.pubsub.NotifyInstances(cached.Namespace)
	engine.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Workflow '%s' has been triggered by %s.", args.Path, args.Caller)
	engine.logger.Infof(ctx, im.cached.File.ID, im.cached.GetAttributes(recipient.Workflow), "Instance '%s' created by %s.", im.ID().String(), args.Caller)
	engine.logger.Debugf(ctx, im.cached.Instance.ID, im.GetAttributes(), "Preparing workflow triggered by %s.", args.Caller)

	// Broadcast Event
	err = engine.flow.BroadcastInstance(BroadcastEventTypeInstanceStarted, ctx,
		broadcastInstanceInput{
			WorkflowPath: args.Path,
			InstanceID:   im.ID().String(),
			Caller:       args.Caller,
		}, cached)
	if err != nil {
		return nil, err
	}
	return im, nil
}

func (engine *engine) start(im *instanceMemory) {
	ctx, err := engine.InstanceLock(im, defaultLockWait)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	engine.sugar.Debugf("Starting workflow %v", im.ID().String())
	engine.logger.Infof(ctx, im.cached.Namespace.ID, im.cached.GetAttributes(recipient.Namespace), "Starting workflow %v", database.GetWorkflow(im.cached.Instance.As))
	engine.logger.Infof(ctx, im.cached.File.ID, im.cached.GetAttributes(recipient.Workflow), "Starting workflow %v", database.GetWorkflow(im.cached.Instance.As))
	engine.logger.Debugf(ctx, im.cached.Instance.ID, im.GetAttributes(), "Starting workflow %v.", database.GetWorkflow(im.cached.Instance.As))

	workflow, err := im.Model()
	if err != nil {
		engine.CrashInstance(ctx, im, derrors.NewUncatchableError(ErrCodeWorkflowUnparsable, "failed to parse workflow YAML: %v", err))
		engine.logger.Errorf(ctx, im.cached.Namespace.ID, im.cached.GetAttributes(recipient.Namespace), "failed to parse workflow YAML")
		return
	}

	start := workflow.GetStartState()

	engine.Transition(ctx, im, start.GetID(), 0)
}

func (engine *engine) loadStateLogic(im *instanceMemory, stateID string) error {
	workflow, err := im.Model()
	if err != nil {
		return err
	}

	wfstates := workflow.GetStatesMap()
	state, exists := wfstates[stateID]
	if !exists {
		return fmt.Errorf("workflow %s cannot resolve state: %s", database.GetWorkflow(im.cached.Instance.As), stateID)
	}

	im.logic, err = states.StateLogic(im, state)
	if err != nil {
		return err
	}

	return nil
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

	if nextState == "" {
		panic("don't call this function with an empty nextState")
	}

	err = engine.loadStateLogic(im, nextState)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return
	}

	flow := append(im.Flow(), nextState)
	deadline := im.logic.Deadline(ctx)

	err = engine.SetMemory(ctx, im, nil)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return
	}

	ctx, cleanup, err := traceStateGenericBegin(ctx, im)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return
	}
	defer cleanup()

	t := time.Now()
	data := im.MarshalData()
	memory := im.MarshalMemory()
	updater := im.getRuntimeUpdater()
	updater = updater.SetFlow(flow).
		SetController(engine.pubsub.Hostname).
		SetAttempts(attempt).
		SetDeadline(deadline).
		SetStateBeginTime(t).
		SetData(data).
		SetMemory(memory)
	im.runtime.Flow = flow
	im.runtime.Controller = engine.pubsub.Hostname
	im.runtime.Attempts = attempt
	im.runtime.Deadline = deadline
	im.runtime.StateBeginTime = t
	im.runtime.Data = data
	im.runtime.Memory = memory
	im.runtimeUpdater = updater

	err = im.flushUpdates(ctx)
	if err != nil {
		engine.sugar.Errorf("Failed to update database record: %v", err) // TODO: how often does this happens? And what are the consequences when we continue running?
		return
	}

	engine.ScheduleSoftTimeout(im, oldController, deadline)

	engine.runState(ctx, im, nil, nil)
}

func (engine *engine) CrashInstance(ctx context.Context, im *instanceMemory, err error) {
	cerr := new(derrors.CatchableError)
	uerr := new(derrors.UncatchableError)

	if errors.As(err, &cerr) {
		engine.reportInstanceCrashed(ctx, im, "catchable", cerr.Code, err)
	} else if errors.As(err, &uerr) && uerr.Code != "" {
		engine.reportInstanceCrashed(ctx, im, "uncatchable", uerr.Code, err)
	} else {
		_, file, line, _ := runtime.Caller(1)
		engine.reportInstanceCrashed(ctx, im, "unknown", fmt.Sprintf("thrown by %s:%d", file, line), err)
	}

	err = engine.SetInstanceFailed(ctx, im, err)
	if err != nil {
		engine.sugar.Error(err)
	}

	broadcastErr := engine.flow.BroadcastInstance(BroadcastEventTypeInstanceFailed, ctx, broadcastInstanceInput{
		WorkflowPath: GetInodePath(im.cached.Instance.As),
		InstanceID:   im.cached.Instance.ID.String(),
	}, im.cached)
	if broadcastErr != nil {
		engine.sugar.Errorf("Failed to broadcast: %v", broadcastErr)
	}

	engine.TerminateInstance(ctx, im)
}

func (engine *engine) setEndAt(im *instanceMemory) {
	t := time.Now()
	updater := im.getInstanceUpdater()
	updater = updater.SetEndAt(t)
	im.cached.Instance.EndAt = t
	im.instanceUpdater = updater
}

func (engine *engine) TerminateInstance(ctx context.Context, im *instanceMemory) {
	engine.setEndAt(im)

	err := im.flushUpdates(ctx)
	if err != nil {
		engine.sugar.Errorf("Failed to update database record: %v", err)
		return
	}

	if im.logic != nil {
		engine.metricsCompleteState(ctx, im, "", im.ErrorCode(), false)
	}

	engine.metricsCompleteInstance(ctx, im)
	engine.FreeInstanceMemory(im)
	engine.WakeInstanceCaller(ctx, im)
}

func (engine *engine) runState(ctx context.Context, im *instanceMemory, wakedata []byte, err error) {
	defer func() {
		e := im.flushUpdates(ctx)
		if e != nil {
			err = e
		}
	}()

	engine.logRunState(ctx, im, wakedata, err)

	var code string
	var transition *states.Transition

	ctx, cleanup, e2 := traceStateGenericLogicThread(ctx, im)
	if e2 != nil {
		err = e2
		goto failure
	}
	defer cleanup()

	if err != nil {
		goto failure
	}

	if lq := im.logic.GetLog(); im.GetMemory() == nil && len(wakedata) == 0 && lq != nil {
		var object interface{}
		object, err = jqOne(im.data, lq)
		if err != nil {
			goto failure
		}

		var data []byte
		data, err = json.MarshalIndent(object, "", "  ")
		if err != nil {
			err = derrors.NewInternalError(fmt.Errorf("failed to marshal state data: %w", err))
			goto failure
		}

		engine.UserLog(ctx, im, string(data))
	}

	if md := im.logic.GetMetadata(); im.GetMemory() == nil && len(wakedata) == 0 && md != nil {
		var object interface{}
		object, err = jqOne(im.data, md)
		if err != nil {
			goto failure
		}

		var data []byte
		data, err = json.MarshalIndent(object, "", "  ")
		if err != nil {
			err = derrors.NewInternalError(fmt.Errorf("failed to marshal state data: %w", err))
			goto failure
		}

		engine.StoreMetadata(ctx, im, string(data))
	}

	transition, err = im.logic.Run(ctx, wakedata)
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

	traceStateError(ctx, err)

	var breaker int

	if breaker > 10 {
		err = derrors.NewInternalError(errors.New("somehow ended up in a catchable error loop"))
	}

	engine.CancelInstanceChildren(ctx, im)

	cerr := new(derrors.CatchableError)

	if errors.As(err, &cerr) {
		_ = im.StoreData("error", cerr)

		for i, catch := range im.logic.ErrorDefinitions() {
			errRegex := catch.Error
			if errRegex == "*" {
				errRegex = ".*"
			}

			matched, regErr := regexp.MatchString(errRegex, cerr.Code)
			if regErr != nil {
				engine.logger.Errorf(ctx, im.GetInstanceID(), im.GetAttributes(), "Error catching regex failed to compile: %v", regErr)
			}

			if matched {
				engine.logger.Errorf(ctx, im.GetInstanceID(), im.GetAttributes(), "State failed with error '%s': %s", cerr.Code, cerr.Message)
				engine.logger.Errorf(ctx, im.GetInstanceID(), im.GetAttributes(), "Error caught by error definition %d: %s", i, catch.Error)

				transition = &states.Transition{
					Transform: "",
					NextState: catch.Transition,
				}

				// breaker++

				code = cerr.Code

				goto next
			}
		}
	}

	engine.CrashInstance(ctx, im, err)
}

func (engine *engine) transformState(ctx context.Context, im *instanceMemory, transition *states.Transition) error {
	if transition == nil || transition.Transform == nil {
		return nil
	}

	if s, ok := transition.Transform.(string); ok && (s == "" || s == ".") {
		return nil
	}

	engine.logger.Debugf(ctx, im.GetInstanceID(), im.GetAttributes(), "Transforming state data.")

	x, err := jqObject(im.data, transition.Transform)
	if err != nil {
		return derrors.WrapCatchableError("unable to apply transform: %v", err)
	}

	im.data = x

	return nil
}

func (engine *engine) transitionState(ctx context.Context, im *instanceMemory, transition *states.Transition, errCode string) {
	e := im.flushUpdates(ctx)
	if e != nil {
		engine.sugar.Errorf("Failed to flush updates: %v", e)
	}

	if transition == nil {
		engine.InstanceYield(im)
		return
	}

	if transition.NextState != "" {
		engine.metricsCompleteState(ctx, im, transition.NextState, errCode, false)
		engine.sugar.Debugf("Instance transitioning to next state: %s -> %s", im.ID().String(), transition.NextState)
		engine.logger.Debugf(ctx, im.GetInstanceID(), im.GetAttributes(), "Transitioning to next state: %s (%d).", transition.NextState, im.Step()+1)
		go engine.Transition(ctx, im, transition.NextState, 0)
		return
	}
	status := util.InstanceStatusComplete
	if im.ErrorCode() != "" {
		status = util.InstanceStatusFailed
		engine.sugar.Debugf("Instance failed: %s", im.ID().String())
		engine.logger.Errorf(ctx, im.cached.Namespace.ID, im.cached.GetAttributes(recipient.Namespace), "Workflow failed with error '%s': %s", im.ErrorCode(), im.cached.Instance.ErrorMessage)
	}

	engine.sugar.Debugf("Instance terminated: %s", im.ID().String())

	rtUpdater := im.getRuntimeUpdater()
	output := im.MarshalData()
	rtUpdater = rtUpdater.SetOutput(output)
	im.runtime.Output = output
	im.runtimeUpdater = rtUpdater

	updater := im.getInstanceUpdater()
	updater = updater.SetStatus(status)
	im.cached.Instance.Status = status
	im.instanceUpdater = updater

	// engine.pubsub.NotifyInstance(im.cached.Instance)

	engine.logger.Infof(ctx, im.GetInstanceID(), im.GetAttributes(), "Workflow %s completed.", database.GetWorkflow(im.cached.Instance.As))
	engine.logger.Infof(ctx, im.cached.Namespace.ID, im.cached.GetAttributes(recipient.Namespace), "Workflow %s completed.", database.GetWorkflow(im.cached.Instance.As))

	engine.pubsub.NotifyInstances(im.cached.Namespace)
	broadcastErr := engine.flow.BroadcastInstance(BroadcastEventTypeInstanceSuccess, ctx, broadcastInstanceInput{
		WorkflowPath: GetInodePath(im.cached.Instance.As),
		InstanceID:   im.cached.Instance.ID.String(),
	}, im.cached)
	if broadcastErr != nil {
		engine.sugar.Errorf("Failed to broadcast: %v", broadcastErr)
	}

	engine.TerminateInstance(ctx, im)
}

func (engine *engine) subflowInvoke(ctx context.Context, caller *subflowCaller, cached *database.CacheData, name string, input []byte) (*instanceMemory, error) {
	var err error

	elems := strings.SplitN(name, ":", 2)

	args := new(newInstanceArgs)
	args.Namespace = cached.Namespace.Name
	args.Path = elems[0]
	if len(elems) == 2 {
		args.Ref = elems[1]
	}
	args.Iterator = caller.Iterator
	args.Input = input
	args.Caller = fmt.Sprintf("instance:%v", caller.InstanceID)
	args.CallPath = caller.CallPath
	callerData, err := json.Marshal(caller)
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}

	args.CallerData = string(callerData)
	args.CallerState = caller.CallerState
	pcached := new(database.CacheData)

	err = engine.database.Instance(ctx, pcached, caller.InstanceID)
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}

	fStore, _, _, rollback, err := engine.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, revision, err := fStore.GetRevision(ctx, pcached.Instance.Revision)
	if err != nil {
		return nil, err
	}

	pcached.File = file
	pcached.Revision = revision

	threadVars, err := engine.database.ThreadVariables(ctx, pcached.Instance.ID)
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}

	if !filepath.IsAbs(args.Path) {
		dir, _ := filepath.Split(caller.As)
		if dir == "" {
			dir = "/"
		}
		args.Path = filepath.Join(dir, args.Path)
	}

	im, err := engine.NewInstance(ctx, args)
	if err != nil {
		return nil, err
	}
	clients := engine.edb.Clients(context.Background())

	for _, tv := range threadVars {
		err = clients.VarRef.Create().SetBehaviour("thread").SetInstanceID(im.cached.Instance.ID).SetName(tv.Name).SetVardataID(tv.VarData).Exec(ctx)
		if err != nil {
			return nil, derrors.NewInternalError(err)
		}
	}
	im.AddAttribute("loop-index", caller.Iterator)
	traceSubflowInvoke(ctx, args.Path, im.ID().String())

	return im, nil
}

type retryMessage struct {
	InstanceID string
	// State      string
	Step int
	Data []byte
}

const retryWakeupFunction = "retryWakeup"

func (engine *engine) scheduleRetry(id, state string, step int, t time.Time, data []byte) error {
	data, err := json.Marshal(&retryMessage{
		InstanceID: id,
		// State:      state,
		Step: step,
		Data: data,
	})
	if err != nil {
		panic(err)
	}

	if d := time.Until(t); d < time.Second*5 {
		go func() {
			time.Sleep(d)
			/* #nosec */
			engine.retryWakeup(data)
		}()
		return nil
	}

	err = engine.timers.addOneShot(id, retryWakeupFunction, t, data)
	if err != nil {
		return derrors.NewInternalError(err)
	}

	return nil
}

func (engine *engine) retryWakeup(data []byte) {
	msg := new(retryMessage)

	err := json.Unmarshal(data, msg)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	ctx, im, err := engine.loadInstanceMemory(msg.InstanceID, msg.Step)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	engine.logger.Infof(ctx, im.GetInstanceID(), im.GetAttributes(), "Waking up to retry.")

	engine.sugar.Debugf("Handling retry wakeup: %s", this())

	go engine.runState(ctx, im, msg.Data, nil)
}

type actionResultPayload struct {
	ActionID     string
	ErrorCode    string
	ErrorMessage string
	Output       []byte
}

type actionResultMessage struct {
	InstanceID string
	State      string
	Step       int
	Payload    actionResultPayload
}

func (engine *engine) doActionRequest(ctx context.Context, ar *functionRequest) error {
	if ar.Workflow.Timeout == 0 {
		ar.Workflow.Timeout = 5 * 60 // 5 mins default, knative's default
	}

	// Log warning if timeout exceeds max allowed timeout
	if actionTimeout := time.Duration(ar.Workflow.Timeout) * time.Second; actionTimeout > engine.conf.GetFunctionsTimeout() {
		_, err := engine.internal.ActionLog(context.Background(), &grpc.ActionLogRequest{
			InstanceId: ar.Workflow.InstanceID, Msg: []string{fmt.Sprintf("Warning: Action timeout '%v' is longer than max allowed duariton '%v'", actionTimeout, engine.conf.GetFunctionsTimeout())},
		})
		if err != nil {
			engine.sugar.Errorf("Failed to log: %v.", err)
		}
	}

	switch ar.Container.Type {
	case model.DefaultFunctionType:
		fallthrough
	case model.NamespacedKnativeFunctionType:
		fallthrough
	case model.ReusableContainerFunctionType:
		go engine.doKnativeHTTPRequest(ctx, ar)
	}

	return nil
}

func (engine *engine) doKnativeHTTPRequest(ctx context.Context,
	ar *functionRequest,
) {
	var err error

	tr := engine.createTransport()

	// configured namespace for workflows
	ns := os.Getenv(util.DirektivServiceNamespace)

	// set service name if namespace
	// otherwise generate baes on action request
	svn := ar.Container.Service

	if ar.Container.Type == model.ReusableContainerFunctionType {
		scale := int32(ar.Container.Scale)
		size := int32(ar.Container.Size)
		svn, _, _ = functions.GenerateServiceName(&igrpc.BaseInfo{
			Name:          &ar.Container.ID,
			Namespace:     &ar.Workflow.NamespaceID,
			Workflow:      &ar.Workflow.WorkflowID,
			Revision:      &ar.Workflow.Revision,
			NamespaceName: &ar.Workflow.NamespaceName,
			Cmd:           &ar.Container.Cmd,
			Image:         &ar.Container.Image,
			MinScale:      &scale,
			Size:          &size,
			Envs:          make(map[string]string),
		})
		if err != nil {
			engine.sugar.Errorf("can not create service name: %v", err)
			engine.reportError(ar, err)
		}
	}

	addr := fmt.Sprintf("http://%s.%s", svn, ns)
	engine.sugar.Debugf("function request for image %s name %s addr %v:", ar.Container.Image, ar.Container.ID, addr)
	engine.logger.Debugf(ctx, engine.flow.ID, engine.flow.GetAttributes(), "function request for image %s name %s", ar.Container.Image, ar.Container.ID)

	deadline := time.Now().Add(time.Duration(ar.Workflow.Timeout) * time.Second)
	rctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	engine.sugar.Debugf("deadline for request: %v", time.Until(deadline))

	req, err := http.NewRequestWithContext(rctx, http.MethodPost, addr,
		bytes.NewReader(ar.Container.Data))
	if err != nil {
		engine.reportError(ar, err)
		return
	}

	// add headers
	req.Header.Add(DirektivDeadlineHeader, deadline.Format(time.RFC3339))
	req.Header.Add(DirektivNamespaceHeader, ar.Workflow.NamespaceName)
	req.Header.Add(DirektivActionIDHeader, ar.ActionID)
	req.Header.Add(DirektivInstanceIDHeader, ar.Workflow.InstanceID)
	req.Header.Add(DirektivStepHeader, fmt.Sprintf("%d",
		int64(ar.Workflow.Step)))
	req.Header.Add(DirektivIteratorHeader, fmt.Sprintf("%d",
		int64(ar.Iterator)))
	for i := range ar.Container.Files {
		f := &ar.Container.Files[i]
		data, err := json.Marshal(f)
		if err != nil {
			panic(err)
		}
		str := base64.StdEncoding.EncodeToString(data)
		req.Header.Add(DirektivFileHeader, str)
	}

	client := &http.Client{
		Transport: tr,
	}

	var resp *http.Response

	// potentially dns error for a brand new service
	// we just loop and see if we can recreate the service
	// one minute wait max
	cleanup := util.TraceHTTPRequest(ctx, req)
	defer cleanup()
	errorReusableContainerMissingRepoted := false
	for i := 0; i < 180; i++ {
		engine.sugar.Debugf("functions request (%d): %v", i, addr)
		resp, err = client.Do(req)
		if err != nil {
			if ctxErr := rctx.Err(); ctxErr != nil {
				engine.sugar.Debugf("context error in knative call")
				return
			}
			engine.logger.Debugf(ctx, engine.flow.ID, engine.flow.GetAttributes(), "function request for image %s name %s returned an error: %v", ar.Container.Image, ar.Container.ID, err)
			dnsErr := new(net.DNSError)
			if errors.As(err, &dnsErr) {
				// recreate if the service does not exist
				if ar.Container.Type == model.ReusableContainerFunctionType &&
					!engine.isKnativeFunction(engine.actions.client, ar) {
					engine.sugar.Debugf("creating KnativeFunction %s %s", ar.Container.Image, ar.Container.ID)
					err := createKnativeFunction(engine.actions.client, ar)
					if err != nil && !strings.Contains(err.Error(), "already exists") {
						engine.sugar.Errorf("can not create knative function: %v", err)
						engine.reportError(ar, err)
						return
					}
				}

				// recreate if the service if it exists in the database but not knative
				if (ar.Container.Type == model.NamespacedKnativeFunctionType) &&
					!engine.isScopedKnativeFunction(engine.actions.client, ar.Container.Service) {
					err := reconstructScopedKnativeFunction(engine.actions.client, ar.Container.Service)
					if err != nil {
						if stErr, ok := status.FromError(err); ok && stErr.Code() == codes.NotFound {
							engine.sugar.Errorf("knative function: '%s' does not exist", ar.Container.Service)
							engine.reportError(ar, fmt.Errorf("knative function: '%s' does not exist, image %s name %s", ar.Container.Service, ar.Container.Image, ar.Container.ID))
							return
						}

						engine.sugar.Errorf("can not create scoped knative function: %v", err)
						engine.reportError(ar, err)
						return
					}
				}
				if errorReusableContainerMissingRepoted && i > 18 && ar.Container.Type == model.ReusableContainerFunctionType {
					err := fmt.Errorf("reusable container image %s is probably missing", ar.Container.Image)
					engine.sugar.Errorf("reusable knative function is missing: %v", err)
					engine.reportError(ar, err)
					errorReusableContainerMissingRepoted = true
				}
				time.Sleep(1000 * time.Millisecond)
				continue
			}

			time.Sleep(1000 * time.Millisecond)
		} else {
			engine.sugar.Debugf("successfully created function with image %s name %s", ar.Container.Image, ar.Container.ID, err)
			break
		}
	}

	if err != nil {
		err := fmt.Errorf("failed creating function with image %s name %s with error: %w", ar.Container.Image, ar.Container.ID, err)
		engine.reportError(ar, err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		engine.reportError(ar, fmt.Errorf("action error status: %d",
			resp.StatusCode))
	}

	engine.sugar.Debugf("function request done")
}

func (engine *engine) reportError(ar *functionRequest, err error) {
	ec := ""
	em := err.Error()
	step := int32(ar.Workflow.Step)
	r := &grpc.ReportActionResultsRequest{
		InstanceId:   ar.Workflow.InstanceID,
		Step:         step,
		ActionId:     ar.ActionID,
		ErrorCode:    ec,
		ErrorMessage: em,
		Iterator:     int32(ar.Iterator),
	}

	_, err = engine.internal.ReportActionResults(context.Background(), r)
	if err != nil {
		engine.sugar.Errorf("can not respond to flow: %v", err)
	}
}

func (engine *engine) createTransport() *http.Transport {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return tr
}

func (engine *engine) wakeEventsWaiter(signature []byte, events []*cloudevents.Event) {
	sig := new(eventsWaiterSignature)
	err := json.Unmarshal(signature, sig)
	if err != nil {
		err = derrors.NewInternalError(err)
		engine.sugar.Error(err)
		return
	}

	ctx, im, err := engine.loadInstanceMemory(sig.InstanceID, sig.Step)
	if err != nil {
		err = fmt.Errorf("cannot load workflow logic instance: %w", err)
		engine.sugar.Error(err)
		return
	}

	wakedata, err := json.Marshal(events)
	if err != nil {
		err = fmt.Errorf("cannot marshal the action results payload: %w", err)
		engine.CrashInstance(ctx, im, err)
		return
	}

	engine.sugar.Debugf("Handling events wakeup: %s", this())

	ctx, cleanup, err := traceStateGenericBegin(ctx, im)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return
	}
	defer cleanup()

	engine.runState(ctx, im, wakedata, nil)
}

func (engine *engine) EventsInvoke(workflowID string, events ...*cloudevents.Event) {
	ctx := context.Background()

	id, err := uuid.Parse(workflowID)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	fStore, _, _, rollback, err := engine.flow.beginSqlTx(ctx)
	if err != nil {
		engine.sugar.Error(err)
		return
	}
	defer rollback()

	file, err := fStore.GetFile(ctx, id)
	if err != nil {
		engine.sugar.Error(err)
		return
	}
	rollback()

	cached := new(database.CacheData)

	ns, err := engine.edb.Namespace(ctx, file.RootID)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	cached.Namespace = ns
	cached.File = file

	var input []byte
	m := make(map[string]interface{})
	for _, event := range events {
		if event == nil {
			continue
		}

		m[event.Type()] = event
	}

	input, err = json.Marshal(m)
	if err != nil {
		engine.sugar.Errorf("Internal error on EventsInvoke: %v", err)
		return
	}

	args := new(newInstanceArgs)
	args.Namespace = cached.Namespace.Name
	args.Path = cached.File.Path

	args.Input = input
	args.Caller = "cloudevent"

	im, err := engine.NewInstance(ctx, args)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	engine.queue(im)
}

func (engine *engine) SetMemory(ctx context.Context, im *instanceMemory, x interface{}) error {
	im.setMemory(x)

	data, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	s := string(data)

	updater := im.getRuntimeUpdater()
	updater = updater.SetMemory(s)
	im.runtime.Memory = s
	im.runtimeUpdater = updater

	return nil
}

func (engine *engine) reportInstanceCrashed(ctx context.Context, im *instanceMemory, typ, code string, err error) {
	engine.sugar.Errorf("Instance failed with %s error '%s': %v", typ, code, err)
	engine.logger.Errorf(ctx, im.GetInstanceID(), im.GetAttributes(), "Instance failed with %s error '%s': %s", typ, code, err.Error())
	engine.logger.Errorf(ctx, im.cached.Instance.Namespace, im.cached.GetAttributes(recipient.Namespace), "Workflow failed %s Instance %s crashed with %s error '%s': %s", database.GetWorkflow(im.cached.Instance.As), im.GetInstanceID(), typ, code, err.Error())
}

func rollback(tx database.Transaction) {
	err := tx.Rollback()
	if err != nil && !strings.Contains(err.Error(), "already been") {
		fmt.Fprintf(os.Stderr, "failed to rollback transaction: %v\n", err)
	}
}

// GetInodePath returns the exact path to a inode.
func GetInodePath(path string) string {
	path = strings.TrimSuffix(path, "/")
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	path = filepath.Clean(path)
	return path
}
