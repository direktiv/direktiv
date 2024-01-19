package flow

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
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
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/model"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/google/uuid"
	"github.com/senseyeio/duration"
	"go.opentelemetry.io/otel/trace"
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
	ID            uuid.UUID
	Namespace     *database.Namespace
	CalledAs      string
	Input         []byte
	Invoker       string
	DescentInfo   *enginerefactor.InstanceDescentInfo
	TelemetryInfo *enginerefactor.InstanceTelemetryInfo
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

func trim(s string) string {
	return strings.TrimPrefix(s, "/")
}

func (engine *engine) NewInstance(ctx context.Context, args *newInstanceArgs) (*instanceMemory, error) {
	file, data, err := engine.mux(ctx, args.Namespace, args.CalledAs)
	if err != nil {
		engine.sugar.Debugf("Failed to create new instance: %v", err)
		engine.logger.Errorf(ctx, engine.flow.ID, engine.flow.GetAttributes(), "Failed to receive workflow %s", args.CalledAs)
		if derrors.IsNotFound(err) {
			return nil, derrors.NewUncatchableError("direktiv.workflow.notfound", "workflow not found: %v", err.Error())
		}
		return nil, err
	}

	var wf model.Workflow
	err = wf.Load(data)
	if err != nil {
		return nil, derrors.NewUncatchableError("direktiv.workflow.invalid", "cannot parse workflow '%s': %v", trim(file.Path), err)
	}

	if len(wf.GetStartDefinition().GetEvents()) > 0 {
		if strings.ToLower(args.Invoker) == apiCaller {
			return nil, derrors.NewUncatchableError("direktiv.workflow.invoke", "cannot manually invoke event-based workflow")
		}
		if strings.HasPrefix(args.Invoker, "instance") {
			return nil, derrors.NewUncatchableError("direktiv.workflow.invoke", "cannot invoke event-based workflow as a subflow")
		}
	}

	/*
		as := args.Path
		if args.Ref != "" {
			as += ":" + args.Ref
		}
		callerInstanceID := ""
		if strings.HasPrefix(args.Caller, "instance:") {
			callerInstanceID = strings.Split(args.Caller, ":")[1]
		}
		TODO: alan, put this somewhere else
		callpath := internallogger.AppendInstanceID(args.CallPath, callerInstanceID)
	*/

	root := args.ID
	iterator := 0
	if args.DescentInfo != nil && len(args.DescentInfo.Descent) > 0 {
		root = args.DescentInfo.Descent[0].ID
		iterator = args.DescentInfo.Descent[len(args.DescentInfo.Descent)-1].Branch
	}

	descentInfo, err := args.DescentInfo.MarshalJSON()
	if err != nil {
		panic(err)
	}

	args.TelemetryInfo.NamespaceName = args.Namespace.Name
	telemetryInfo, err := args.TelemetryInfo.MarshalJSON()
	if err != nil {
		panic(err)
	}

	settings := &enginerefactor.InstanceSettings{
		// TODO: alan, LogToEvents:
		NamespaceConfig: []byte(args.Namespace.Config),
	}
	settingsData, err := settings.MarshalJSON()
	if err != nil {
		panic(err)
	}

	liveData := marshalInstanceInputData(args.Input)

	ri := &enginerefactor.InstanceRuntimeInfo{}
	riData, err := ri.MarshalJSON()
	if err != nil {
		panic(err)
	}

	ci := &enginerefactor.InstanceChildrenInfo{}
	ciData, err := ci.MarshalJSON()
	if err != nil {
		panic(err)
	}

	tx, err := engine.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	idata, err := tx.InstanceStore().CreateInstanceData(ctx, &instancestore.CreateInstanceDataArgs{
		ID:             args.ID,
		NamespaceID:    args.Namespace.ID,
		Namespace:      args.Namespace.Name,
		RootInstanceID: root,
		Invoker:        args.Invoker,
		WorkflowPath:   file.Path,
		Definition:     data,
		Input:          args.Input,
		LiveData:       []byte(liveData),
		TelemetryInfo:  telemetryInfo,
		Settings:       settingsData,
		DescentInfo:    descentInfo,
		RuntimeInfo:    riData,
		ChildrenInfo:   ciData,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	instance, err := enginerefactor.ParseInstanceData(idata)
	if err != nil {
		panic(err)
	}

	im := new(instanceMemory)
	im.engine = engine
	im.instance = instance
	im.updateArgs = new(instancestore.UpdateInstanceDataArgs)

	err = json.Unmarshal(im.instance.Instance.LiveData, &im.data)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(im.instance.Instance.StateMemory, &im.memory)
	if err != nil {
		panic(err)
	}

	ctx, err = traceFullAddWorkflowInstance(ctx, im)
	if err != nil {
		return nil, fmt.Errorf("failed to traceFullAddWorkflowInstance: %w", err)
	}
	im.AddAttribute("loop-index", fmt.Sprintf("%d", iterator))

	engine.pubsub.NotifyInstances(im.Namespace())
	engine.logger.Infof(ctx, instance.Instance.NamespaceID, instance.GetAttributes(recipient.Namespace), "Workflow '%s' has been triggered by %s.", args.CalledAs, args.Invoker)
	// TODO: alex, do we need to restore workflow logs?
	// engine.logger.Infof(ctx, im.instance.Instance.WorkflowID, im.instance.GetAttributes(recipient.Workflow), "Instance '%s' created by %s.", im.ID().String(), args.Invoker)
	engine.logger.Debugf(ctx, im.instance.Instance.ID, im.GetAttributes(), "Preparing workflow triggered by %s.", args.Invoker)

	// Broadcast Event
	err = engine.flow.BroadcastInstance(BroadcastEventTypeInstanceStarted, ctx,
		broadcastInstanceInput{
			WorkflowPath: args.CalledAs,
			InstanceID:   im.ID().String(),
			Caller:       args.Invoker,
		}, im.instance)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast instance: %w", err)
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
	engine.logger.Infof(ctx, im.instance.Instance.NamespaceID, im.instance.GetAttributes(recipient.Namespace), "Starting workflow %v", database.GetWorkflow(im.instance.Instance.WorkflowPath))
	// TODO: alex, do we need to restore workflow logs?
	// engine.logger.Infof(ctx, im.instance.Instance.WorkflowID, im.instance.GetAttributes(recipient.Workflow), "Starting workflow %v", database.GetWorkflow(im.instance.Instance.CalledAs))
	engine.logger.Debugf(ctx, im.instance.Instance.ID, im.GetAttributes(), "Starting workflow %v.", database.GetWorkflow(im.instance.Instance.WorkflowPath))

	workflow, err := im.Model()
	if err != nil {
		engine.CrashInstance(ctx, im, derrors.NewUncatchableError(ErrCodeWorkflowUnparsable, "failed to parse workflow YAML: %v", err))
		engine.logger.Errorf(ctx, im.instance.Instance.NamespaceID, im.instance.GetAttributes(recipient.Namespace), "failed to parse workflow YAML")
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
		return fmt.Errorf("workflow %s cannot resolve state: %s", database.GetWorkflow(im.instance.Instance.WorkflowPath), stateID)
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
		t := time.Now().UTC()
		tSoft := time.Now().UTC().Add(time.Minute * 15)
		tHard := time.Now().UTC().Add(time.Minute * 20)

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

	t := time.Now().UTC()

	im.instance.RuntimeInfo.Flow = flow
	im.instance.RuntimeInfo.Controller = engine.pubsub.Hostname
	im.instance.RuntimeInfo.Attempts = attempt
	im.instance.RuntimeInfo.StateBeginTime = t
	rtData, err := im.instance.RuntimeInfo.MarshalJSON()
	if err != nil {
		panic(err)
	}
	im.updateArgs.RuntimeInfo = &rtData

	im.instance.Instance.Deadline = &deadline
	im.updateArgs.Deadline = im.instance.Instance.Deadline

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
		WorkflowPath: GetInodePath(im.instance.Instance.WorkflowPath),
		InstanceID:   im.instance.Instance.ID.String(),
	}, im.instance)
	if broadcastErr != nil {
		engine.sugar.Errorf("Failed to broadcast: %v", broadcastErr)
	}

	engine.TerminateInstance(ctx, im)
}

func (engine *engine) setEndAt(im *instanceMemory) {
	t := time.Now().UTC()
	im.instance.Instance.EndedAt = &t
	im.updateArgs.EndedAt = im.instance.Instance.EndedAt
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

	im.replaceData(x)

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

	status := instancestore.InstanceStatusComplete
	if im.ErrorCode() != "" {
		status = instancestore.InstanceStatusFailed
		engine.sugar.Debugf("Instance failed: %s", im.ID().String())
		engine.logger.Errorf(ctx, im.instance.Instance.NamespaceID, im.instance.GetAttributes(recipient.Namespace), "Workflow failed with error '%s': %s", im.ErrorCode(), im.instance.Instance.ErrorMessage)
	}

	engine.sugar.Debugf("Instance terminated: %s", im.ID().String())

	output := im.MarshalData()
	im.instance.Instance.Output = []byte(output)
	im.updateArgs.Output = &im.instance.Instance.Output
	im.instance.Instance.Status = status
	im.updateArgs.Status = &im.instance.Instance.Status

	engine.logger.Infof(ctx, im.GetInstanceID(), im.GetAttributes(), "Workflow %s completed.", database.GetWorkflow(im.instance.Instance.WorkflowPath))
	engine.logger.Infof(ctx, im.instance.Instance.NamespaceID, im.instance.GetAttributes(recipient.Namespace), "Workflow %s completed.", database.GetWorkflow(im.instance.Instance.WorkflowPath))

	defer engine.pubsub.NotifyInstance(im.instance.Instance.ID)
	defer engine.pubsub.NotifyInstances(im.Namespace())

	broadcastErr := engine.flow.BroadcastInstance(BroadcastEventTypeInstanceSuccess, ctx, broadcastInstanceInput{
		WorkflowPath: GetInodePath(im.instance.Instance.WorkflowPath),
		InstanceID:   im.instance.Instance.ID.String(),
	}, im.instance)
	if broadcastErr != nil {
		engine.sugar.Errorf("Failed to broadcast: %v", broadcastErr)
	}

	engine.TerminateInstance(ctx, im)
}

func (engine *engine) subflowInvoke(ctx context.Context, pi *enginerefactor.ParentInfo, instance *enginerefactor.Instance, name string, input []byte) (*instanceMemory, error) {
	var err error

	di := &enginerefactor.InstanceDescentInfo{
		Descent: append(instance.DescentInfo.Descent, *pi),
	}

	span := trace.SpanFromContext(ctx)

	args := &newInstanceArgs{
		ID: uuid.New(),
		Namespace: &database.Namespace{
			ID:     instance.Instance.NamespaceID,
			Name:   instance.TelemetryInfo.NamespaceName,
			Config: string(instance.Settings.NamespaceConfig),
		},
		CalledAs:    name,
		Input:       input,
		Invoker:     fmt.Sprintf("instance:%v", pi.ID),
		DescentInfo: di,
		TelemetryInfo: &enginerefactor.InstanceTelemetryInfo{
			TraceID: span.SpanContext().TraceID().String(),
			SpanID:  span.SpanContext().SpanID().String(),
			// TODO: alan, CallPath: ,
			NamespaceName: instance.TelemetryInfo.NamespaceName,
		},
	}

	tx, err := engine.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if !filepath.IsAbs(args.CalledAs) {
		dir, _ := filepath.Split(instance.Instance.WorkflowPath)
		if dir == "" {
			dir = "/"
		}
		args.CalledAs = filepath.Join(dir, args.CalledAs)
	}

	im, err := engine.NewInstance(ctx, args)
	if err != nil {
		return nil, err
	}

	im.AddAttribute("loop-index", fmt.Sprintf("%d", pi.Branch))
	traceSubflowInvoke(ctx, args.CalledAs, im.ID().String())

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

func (engine *engine) wakeEventsWaiter(instance uuid.UUID, step int, events []*cloudevents.Event) {
	ctx, im, err := engine.loadInstanceMemory(instance.String(), step)
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

func (engine *engine) EventsInvoke(workflowID uuid.UUID, events ...*cloudevents.Event) {
	ctx := context.Background()

	tx, err := engine.flow.beginSqlTx(ctx)
	if err != nil {
		engine.sugar.Error(err)
		return
	}
	defer tx.Rollback()

	file, err := tx.FileStore().GetFileByID(ctx, workflowID)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	root, err := tx.FileStore().GetRoot(ctx, file.RootID)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, root.Namespace)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	tx.Rollback()

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

	span := trace.SpanFromContext(ctx)

	args := &newInstanceArgs{
		ID:        uuid.New(),
		Namespace: ns,
		CalledAs:  file.Path,
		Input:     input,
		Invoker:   "cloudevent",
		TelemetryInfo: &enginerefactor.InstanceTelemetryInfo{
			TraceID: span.SpanContext().TraceID().String(),
			SpanID:  span.SpanContext().SpanID().String(),
			// TODO: alan, CallPath: ,
			NamespaceName: ns.Name,
		},
	}

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

	im.instance.Instance.StateMemory = data
	im.updateArgs.StateMemory = &data

	return nil
}

func (engine *engine) reportInstanceCrashed(ctx context.Context, im *instanceMemory, typ, code string, err error) {
	engine.sugar.Errorf("Instance failed with %s error '%s': %v", typ, code, err)
	engine.logger.Errorf(ctx, im.GetInstanceID(), im.GetAttributes(), "Instance failed with %s error '%s': %s", typ, code, err.Error())
	engine.logger.Errorf(ctx, im.instance.Instance.NamespaceID, im.instance.GetAttributes(recipient.Namespace), "Workflow failed %s Instance %s crashed with %s error '%s': %s", database.GetWorkflow(im.instance.Instance.WorkflowPath), im.GetInstanceID(), typ, code, err.Error())
}

func (engine *engine) UserLog(ctx context.Context, im *instanceMemory, msg string, a ...interface{}) {
	engine.logger.Infof(ctx, im.GetInstanceID(), im.GetAttributes(), msg, a...)

	if attr := im.instance.Settings.LogToEvents; attr != "" {
		s := fmt.Sprintf(msg, a...)
		event := cloudevents.NewEvent()
		event.SetID(uuid.New().String())
		event.SetSource(im.instance.Instance.WorkflowPath)
		event.SetType("direktiv.instanceLog")
		event.SetExtension("logger", attr)
		event.SetDataContentType("application/json")
		err := event.SetData("application/json", s)
		if err != nil {
			engine.sugar.Errorf("Failed to create CloudEvent: %v.", err)
		}

		err = engine.events.BroadcastCloudevent(ctx, im.Namespace(), &event, 0)
		if err != nil {
			engine.sugar.Errorf("Failed to broadcast CloudEvent: %v.", err)
			return
		}
	}
}

func (engine *engine) logRunState(ctx context.Context, im *instanceMemory, wakedata []byte, err error) {
	engine.sugar.Debugf("Running state logic -- %s:%v (%s) (%v)", im.ID().String(), im.Step(), im.logic.GetID(), time.Now().UTC())
	if im.GetMemory() == nil && len(wakedata) == 0 && err == nil {
		engine.logger.Infof(ctx, im.GetInstanceID(), im.GetAttributes(), "Running state logic (step:%v) -- %s", im.Step(), im.logic.GetID())
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
