package flow

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	enginerefactor "github.com/direktiv/direktiv/pkg/engine"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/nohome"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/instancestore"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/direktiv/direktiv/pkg/telemetry"
	"github.com/google/uuid"
	"github.com/senseyeio/duration"
	"go.opentelemetry.io/otel/attribute"
)

type engine struct {
	*server

	scheduled sync.Map
}

func initEngine(srv *server) *engine {
	engine := new(engine)

	engine.server = srv

	engine.Bus.Subscribe(&pubsub.InstanceMessageEvent{}, engine.instanceMessagesChannelHandler)

	go engine.instanceKicker()

	return engine
}

func (engine *engine) instanceKicker() {
	<-time.After(1 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	for {
		<-ticker.C
		go engine.kickWaitingInstances()
	}
}

func (engine *engine) kickWaitingInstances() {
	ctx := context.Background()

	slog.Debug("starting to kick waiting (homeless) instances")
	tx, err := engine.beginSQLTx(ctx)
	if err != nil {
		slog.Error("failed to begin SQL transaction in kickWaitingInstances", "error", err)
		return
	}
	defer tx.Rollback()

	instances, err := tx.InstanceStore().GetHomelessInstances(ctx, time.Now().UTC().Add(-engineSchedulingTimeout))
	if err != nil {
		slog.Error("failed to list homeless instances in kickWaitingInstances. Some instances may remain unprocessed", "error", err)
		return
	}
	if len(instances) == 0 {
		slog.Debug("no homeless instances found to kick")
		return
	}
	slog.Info("processing homeless instances", "count", len(instances))

	for idx := range instances {
		instance := instances[idx]
		slog.Debug("kicking instance", "instance", instance.ID)

		data, err := json.Marshal(&instanceMessageChannelData{
			InstanceID:        instance.ID,
			LastKnownServer:   instance.Server,
			LastKnownUpdateAt: instance.UpdatedAt,
		})
		if err != nil {
			slog.Error("failed to marshal instance data in kickWaitingInstances", "instance", instance.ID, "error", err)
		}

		engine.instanceMessagesChannelHandler(string(data))
	}
}

type newInstanceArgs struct {
	tx            *database.DB
	ID            uuid.UUID
	Namespace     *datastore.Namespace
	CalledAs      string
	Input         []byte
	Invoker       string
	DescentInfo   *enginerefactor.InstanceDescentInfo
	TelemetryInfo *enginerefactor.InstanceTelemetryInfo
	SyncHash      *string
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
	ctx, span := enginerefactor.TraceReconstruct(ctx, args.TelemetryInfo, "new-instance")
	defer span.End()

	if args.TelemetryInfo.CallPath == "" {
		args.TelemetryInfo.CallPath = "/" + args.ID.String() + "/"
	} else {
		args.TelemetryInfo.CallPath = args.TelemetryInfo.CallPath + args.ID.String() + "/"
	}

	traceParent := telemetry.TraceParent(ctx)
	args.TelemetryInfo.TraceParent = traceParent

	file, data, err := engine.mux(ctx, args.Namespace, args.CalledAs)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "id",
			Value: attribute.StringValue(args.ID.String()),
		},
		attribute.KeyValue{
			Key:   "namespace",
			Value: attribute.StringValue(args.Namespace.Name),
		},
		attribute.KeyValue{
			Key:   "path",
			Value: attribute.StringValue(file.Path),
		},
	)

	ctx = telemetry.LogInitCtx(ctx, telemetry.LogObject{
		Namespace: args.Namespace.Name,
		ID:        args.ID.String(),
		Scope:     telemetry.LogScopeInstance,
		InstanceInfo: telemetry.InstanceInfo{
			Invoker:  args.Invoker,
			Status:   core.LogRunningStatus,
			State:    "new-instance",
			Path:     file.Path,
			CallPath: args.TelemetryInfo.CallPath,
		},
	})
	telemetry.LogInstance(ctx, telemetry.LogLevelDebug, "creating new instance")

	var wf model.Workflow
	err = wf.Load(data)
	if err != nil {
		telemetry.ReportError(span, err)
		telemetry.LogInstanceError(ctx,
			"failed to parse workflow definition", err)
		return nil, derrors.NewUncatchableError("direktiv.workflow.invalid", "cannot parse workflow '%s': %v", trim(file.Path), err)
	}

	if len(wf.GetStartDefinition().GetEvents()) > 0 {
		if strings.ToLower(args.Invoker) == apiCaller {
			telemetry.ReportError(span, fmt.Errorf("cannot manually invoke event-based workflow"))
			return nil, derrors.NewUncatchableError("direktiv.workflow.invoke", "cannot manually invoke event-based workflow")
		}
		if strings.HasPrefix(args.Invoker, "instance") {
			telemetry.ReportError(span, fmt.Errorf("cannot invoke event-based workflow as a subflow"))
			return nil, derrors.NewUncatchableError("direktiv.workflow.invoke", "cannot invoke event-based workflow as a subflow")
		}
	}

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

	args.TelemetryInfo.TraceParent = traceParent
	telemetryInfo, err := args.TelemetryInfo.MarshalJSON()
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
	tx := args.tx

	if tx == nil {
		tx, err = engine.flow.beginSQLTx(ctx)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()
	}
	telemetry.LogNamespace(telemetry.LogLevelDebug, args.Namespace.Name,
		"preparing to commit new instance transaction")

	idata, err := tx.InstanceStore().CreateInstanceData(ctx, &instancestore.CreateInstanceDataArgs{
		ID:             args.ID,
		NamespaceID:    args.Namespace.ID,
		Namespace:      args.Namespace.Name,
		RootInstanceID: root,
		Server:         engine.ID,
		Invoker:        args.Invoker,
		WorkflowPath:   file.Path,
		Definition:     data,
		Input:          args.Input,
		LiveData:       []byte(liveData),
		TelemetryInfo:  telemetryInfo,
		DescentInfo:    descentInfo,
		RuntimeInfo:    riData,
		ChildrenInfo:   ciData,
		SyncHash:       args.SyncHash,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	telemetry.LogNamespace(telemetry.LogLevelDebug, args.Namespace.Name,
		"new instance transaction committed")

	instance, err := enginerefactor.ParseInstanceData(idata)
	if err != nil {
		panic(err)
	}

	im := new(instanceMemory)
	im.engine = engine
	im.instance = instance
	im.updateArgs = new(instancestore.UpdateInstanceDataArgs)
	im.updateArgs.Server = engine.ID

	err = json.Unmarshal(im.instance.Instance.LiveData, &im.data)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(im.instance.Instance.StateMemory, &im.memory)
	if err != nil {
		panic(err)
	}

	im.AddAttribute("loop-index", fmt.Sprintf("%d", iterator))

	engine.pubsub.NotifyInstances(im.Namespace())

	if engine.config.OtelBackend != "" {
		telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
			fmt.Sprintf("tracing id %s", span.SpanContext().TraceID()))
	}

	telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
		"workflow has been started")

	telemetry.LogNamespace(telemetry.LogLevelInfo, im.Namespace().Name, "workflow has been started")

	return im, nil
}

func (engine *engine) loadStateLogic(im *instanceMemory, stateID string) error {
	workflow, err := im.Model()
	if err != nil {
		return err
	}

	var state model.State

	if stateID == "" {
		state = workflow.GetStartState()
	} else {
		wfstates := workflow.GetStatesMap()
		var exists bool
		state, exists = wfstates[stateID]
		if !exists {
			return fmt.Errorf("workflow %s cannot resolve state: %s", nohome.GetWorkflow(im.instance.Instance.WorkflowPath), stateID)
		}
	}

	im.logic, err = states.StateLogic(im, state)
	if err != nil {
		return err
	}

	return nil
}

func (engine *engine) Transition(ctx context.Context, im *instanceMemory, nextState string, attempt int) *states.Transition {
	// prepare log context
	ctx = im.Context(ctx)

	ctx, span := enginerefactor.TraceReconstruct(ctx, im.instance.TelemetryInfo, fmt.Sprintf("state-%s", nextState))
	defer span.End()

	workflow, err := im.Model()
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return nil
	}

	oldController := im.Controller()

	if im.Step() == 0 { //nolint:nestif
		t := time.Now().UTC()
		tSoft := time.Now().UTC().Add(time.Minute * 15)
		tHard := time.Now().UTC().Add(time.Minute * 20)

		if workflow.Timeouts != nil {
			s := workflow.Timeouts.Interrupt

			if s != "" {
				d, err := duration.ParseISO8601(s)
				if err != nil {
					engine.CrashInstance(ctx, im, err)
					return nil
				}
				tSoft = d.Shift(t)
				tHard = tSoft.Add(time.Minute * 5)
			}

			s = workflow.Timeouts.Kill

			if s != "" {
				d, err := duration.ParseISO8601(s)
				if err != nil {
					engine.CrashInstance(ctx, im, err)
					return nil
				}
				tHard = d.Shift(t)
			}
		}

		engine.ScheduleSoftTimeout(ctx, im, oldController, tSoft)
		engine.ScheduleHardTimeout(ctx, im, oldController, tHard)
	}

	if nextState == "" {
		panic("don't call this function with an empty nextState")
	}

	err = engine.loadStateLogic(im, nextState)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return nil
	}

	flow := append(im.Flow(), nextState)
	deadline := im.logic.Deadline(ctx)

	err = engine.SetMemory(ctx, im, nil)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return nil
	}

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

	im.instance.TelemetryInfo = &enginerefactor.InstanceTelemetryInfo{
		TraceParent: telemetry.TraceParent(ctx),
		CallPath:    im.instance.TelemetryInfo.CallPath,
	}

	err = im.flushUpdates(ctx)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return nil
	}

	engine.ScheduleSoftTimeout(ctx, im, oldController, deadline)

	return engine.runState(ctx, im, nil, nil)
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

	engine.SetInstanceFailed(ctx, im, err)
	engine.TerminateInstance(ctx, im)
}

func (engine *engine) setEndAt(im *instanceMemory) {
	t := time.Now().UTC()
	im.instance.Instance.EndedAt = &t
	im.updateArgs.EndedAt = im.instance.Instance.EndedAt
}

type noCancelCtx struct {
	//nolint:containedctx
	ctx context.Context
}

func (c noCancelCtx) Deadline() (time.Time, bool)       { return time.Time{}, false }
func (c noCancelCtx) Done() <-chan struct{}             { return nil }
func (c noCancelCtx) Err() error                        { return nil }
func (c noCancelCtx) Value(key interface{}) interface{} { return c.ctx.Value(key) }

// WithoutCancel returns a context that is never canceled.
func NoCancelContext(ctx context.Context) context.Context {
	return noCancelCtx{ctx: ctx}
}

func (engine *engine) TerminateInstance(ctx context.Context, im *instanceMemory) {
	if engine.GetIsInstanceCrashed(im) {
		ctx = NoCancelContext(ctx)
	}

	engine.setEndAt(im)

	engine.freeArtefacts(im) //nolint:contextcheck
	err := engine.freeMemory(ctx, im)
	if err != nil {
		if !engine.GetIsInstanceCrashed(im) {
			engine.CrashInstance(ctx, im, err)
			return
		}

		engine.forceFreeCriticalMemory(ctx, im)
		slog.Debug("failed to free memory during a crash", "error", err)
	}

	engine.WakeInstanceCaller(ctx, im)
}

//nolint:gocognit
func (engine *engine) runState(ctx context.Context, im *instanceMemory, wakedata []byte, err error) *states.Transition {
	// prepare log context
	ctx = im.Context(ctx)

	telemetry.LogInstance(ctx, telemetry.LogLevelDebug, "starting state execution")

	var transition *states.Transition

	if err != nil {
		telemetry.LogInstanceError(ctx,
			"error before state execution", err)

		goto failure
	}

	if lq := im.logic.GetLog(); im.GetMemory() == nil && len(wakedata) == 0 && lq != nil {
		var object interface{}
		object, err = jqOne(im.data, lq) //nolint:contextcheck
		if err != nil {
			telemetry.LogInstanceError(ctx,
				fmt.Sprintf("failed to process jq query on state data, %v", lq), err)

			goto failure
		}

		var data []byte
		data, err = json.MarshalIndent(object, "", "  ")
		if err != nil {
			err = derrors.NewInternalError(fmt.Errorf("failed to marshal state data: %w", err))
			telemetry.LogInstanceError(ctx,
				"failed to marshal jq query result for logging", err)

			goto failure
		}

		engine.UserLog(ctx, im, string(data))
	}

	if md := im.logic.GetMetadata(); im.GetMemory() == nil && len(wakedata) == 0 && md != nil {
		var object interface{}
		object, err = jqOne(im.data, md) //nolint:contextcheck
		if err != nil {
			telemetry.LogInstanceError(ctx,
				fmt.Sprintf("failed to execute jq query for metadata, %v", md), err)

			goto failure
		}

		var data []byte
		data, err = json.MarshalIndent(object, "", "  ")
		if err != nil {
			err = derrors.NewInternalError(fmt.Errorf("failed to marshal state data: %w", err))
			telemetry.LogInstanceError(ctx,
				"failed to marshal metadata", err)

			goto failure
		}

		engine.StoreMetadata(ctx, im, string(data))
	}
	telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
		fmt.Sprintf("running state logic %s", im.GetState()))

	// set new parent during transition
	transition, err = im.logic.Run(ctx, wakedata)
	if err != nil {
		telemetry.LogInstanceError(ctx,
			"state logic execution failed", err)

		goto failure
	}
	telemetry.LogInstance(ctx, telemetry.LogLevelDebug,
		"applying state transformation based on logic run")

	err = engine.transformState(ctx, im, transition)
	if err != nil {
		telemetry.LogInstanceError(ctx,
			"state transformation failed", err)

		goto failure
	}

next:
	telemetry.LogInstance(ctx, telemetry.LogLevelDebug,
		"processing post execution actions")

	return engine.transitionState(ctx, im, transition)

failure:
	telemetry.LogInstanceError(ctx, "state execution failed", err)

	var breaker int

	if breaker > 10 {
		err = derrors.NewInternalError(errors.New("somehow ended up in a catchable error loop"))
		telemetry.LogInstanceError(ctx, "possible error loop detected", err)
	}

	err1 := engine.CancelInstanceChildren(ctx, im)
	if err1 != nil {
		telemetry.LogInstanceError(ctx, "canceling instance children failed", err)
	}
	cerr := new(derrors.CatchableError)

	if errors.As(err, &cerr) {
		err2 := im.StoreData("error", cerr)
		if err2 != nil {
			telemetry.LogInstanceError(ctx, "failed to store error data", err)
		}

		for _, catch := range im.logic.ErrorDefinitions() {
			errRegex := catch.Error
			if errRegex == "*" {
				errRegex = ".*"
			}

			matched, regErr := regexp.MatchString(errRegex, cerr.Code)
			if regErr != nil {
				telemetry.LogInstanceError(ctx, "regex compilation failed for error catch definition", err)
			}

			if matched {
				telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
					"executing defined transition with machting catchable error")
				telemetry.LogInstanceError(ctx, fmt.Sprintf("state failed with an error '%s': %s", cerr.Code, cerr.Message), err)

				transition = &states.Transition{
					Transform: "",
					NextState: catch.Transition,
				}

				goto next
			}
		}
	}

	telemetry.LogInstanceError(ctx, "unrecoverable error encountered; initiating instance crash", err)
	engine.CrashInstance(ctx, im, err)

	return nil
}

func (engine *engine) transformState(ctx context.Context, im *instanceMemory, transition *states.Transition) error {
	if transition == nil || transition.Transform == nil {
		return nil
	}

	if s, ok := transition.Transform.(string); ok && (s == "" || s == ".") {
		return nil
	}

	ctx = im.Context(ctx)
	telemetry.LogInstance(ctx, telemetry.LogLevelDebug,
		"transforming state data")

	x, err := jqObject(im.data, transition.Transform) //nolint:contextcheck
	if err != nil {
		telemetry.LogInstanceError(ctx, "failed to apply jq to transform", err)

		return derrors.WrapCatchableError("unable to apply transform: %v", err)
	}

	im.replaceData(x)
	telemetry.LogInstance(ctx, telemetry.LogLevelDebug,
		"transformed state data")

	return nil
}

func (engine *engine) transitionState(ctx context.Context, im *instanceMemory, transition *states.Transition) *states.Transition {
	ctx = im.Context(ctx)
	ctx, _ = enginerefactor.TraceGet(ctx, im.instance.TelemetryInfo)

	e := im.flushUpdates(ctx)
	if e != nil {
		telemetry.LogInstanceError(ctx, fmt.Sprintf("failed to flush updates for instance %v (%v)", im.ID(), im.Namespace().Name), e)
		engine.CrashInstance(ctx, im, e)

		return nil
	}

	if transition == nil {
		engine.InstanceYield(ctx, im)

		return nil
	}

	if transition.NextState != "" {
		telemetry.LogInstance(ctx, telemetry.LogLevelDebug,
			"transitioning to next state")

		return transition
	}

	status := instancestore.InstanceStatusComplete
	if im.ErrorCode() != "" {
		status = instancestore.InstanceStatusFailed
		if im.ErrorCode() == "direktiv.cancels.parent" || im.ErrorCode() == "direktiv.cancels.api" {
			status = instancestore.InstanceStatusCancelled
		}
		telemetry.LogInstanceError(ctx, "workflow failed with error", fmt.Errorf("'%s': %s", im.ErrorCode(), im.instance.Instance.ErrorMessage))
	}

	telemetry.LogInstance(ctx, telemetry.LogLevelDebug,
		fmt.Sprintf("instance %v terminated", im.ID().String()))

	output := im.MarshalData()
	im.instance.Instance.Output = []byte(output)
	im.updateArgs.Output = &im.instance.Instance.Output
	im.instance.Instance.Status = status
	im.updateArgs.Status = &im.instance.Instance.Status

	telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
		"workflow completed")
	telemetry.LogNamespace(telemetry.LogLevelInfo, im.Namespace().Name, "workflow completed")

	defer engine.pubsub.NotifyInstance(im.instance.Instance.ID)
	defer engine.pubsub.NotifyInstances(im.Namespace())

	engine.TerminateInstance(ctx, im)

	return nil
}

func (engine *engine) subflowInvoke(ctx context.Context, pi *enginerefactor.ParentInfo, instance *enginerefactor.Instance, name string, input []byte) (*instanceMemory, error) {
	var err error

	ctx, span := enginerefactor.TraceReconstruct(ctx, instance.TelemetryInfo, "executing-subflow")
	defer span.End()

	di := &enginerefactor.InstanceDescentInfo{
		Descent: append(instance.DescentInfo.Descent, *pi),
	}

	telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
		fmt.Sprintf("invoking subflow %s", name))

	args := &newInstanceArgs{
		ID: uuid.New(),
		Namespace: &datastore.Namespace{
			ID:   instance.Instance.NamespaceID,
			Name: instance.Instance.Namespace,
		},
		CalledAs:    name,
		Input:       input,
		Invoker:     fmt.Sprintf("instance:%v", pi.ID),
		DescentInfo: di,
		TelemetryInfo: &enginerefactor.InstanceTelemetryInfo{
			TraceParent: telemetry.TraceParent(ctx),
			CallPath:    instance.TelemetryInfo.CallPath,
		},
	}

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

	return im, nil
}

type retryMessage struct {
	InstanceID string
	Data       []byte
}

const retryWakeupFunction = "retryWakeup"

func (engine *engine) scheduleRetry(id string, t time.Time, data []byte) error {
	data, err := json.Marshal(&retryMessage{
		InstanceID: id,
		Data:       data,
	})
	if err != nil {
		panic(err) // TODO ?
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
		slog.Error("failed to unmarshal retryMessage", "error", err)
		return
	}

	uid, err := uuid.Parse(msg.InstanceID)
	if err != nil {
		slog.Error("failed to parse instance ID in retryMessage", "error", err)
		return
	}

	ctx := context.Background()

	err = engine.enqueueInstanceMessage(ctx, uid, "wake", msg)
	if err != nil {
		slog.Error("failed to enqueue instance message for retryWakeup", "error", err)
		return
	}
}

type actionResultPayload struct {
	ActionID     string
	ErrorCode    string
	ErrorMessage string
	Output       []byte
}

type actionResultMessage struct {
	InstanceID string
	enginerefactor.ActionContext
	Payload actionResultPayload
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

func (engine *engine) WakeEventsWaiter(instance uuid.UUID, events []*cloudevents.Event) {
	ctx := context.Background()

	ctx, span := telemetry.Tracer.Start(ctx, "event-wake")
	defer span.End()

	err := engine.enqueueInstanceMessage(ctx, instance, "event", events)
	if err != nil {
		slog.Error("failed to enqueue instance message for wakeEventsWaiter", "error", err)
		return
	}
}

func (engine *engine) EventsInvoke(ctx context.Context, workflowID uuid.UUID, events ...*cloudevents.Event) {
	ctx, span := telemetry.Tracer.Start(ctx, "event-invoke")
	defer span.End()

	tx, err := engine.flow.beginSQLTx(context.Background())
	if err != nil {
		slog.Error("failed to begin SQL transaction in EventsInvoke", "error", err)
		return
	}
	defer tx.Rollback()

	file, err := tx.FileStore().GetFileByID(context.Background(), workflowID)
	if err != nil {
		slog.Error("failed to fetch file from database", "workflowID", workflowID, "error", err)
		return
	}

	root, err := tx.FileStore().GetRoot(context.Background(), file.RootID)
	if err != nil {
		slog.Error("failed to fetch Root from database", "workflowID", workflowID, "error", err)
		return
	}

	ns, err := tx.DataStore().Namespaces().GetByName(context.Background(), root.Namespace)
	if err != nil {
		slog.Error("failed to fetch namespace from database", "namespace", root.Namespace, "error", err)
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
		slog.Error("failed to marshal event data in EventsInvoke", "error", err)
		return
	}

	args := &newInstanceArgs{
		ID:        uuid.New(),
		Namespace: ns,
		CalledAs:  file.Path,
		Input:     input,
		Invoker:   "cloudevent",
		TelemetryInfo: &enginerefactor.InstanceTelemetryInfo{
			TraceParent: telemetry.TraceParent(ctx),
		},
	}

	// needs a new context because it will be cancelled by the original call
	// it is getting passed with TraceParent
	im, err := engine.NewInstance(context.Background(), args)
	if err != nil {
		slog.Error("new instance", "error", err)
		return
	}
	slog.Debug("invoked new workflow instance", "instanceID", im.ID().String(), "workflowPath", file.Path)

	go engine.start(im)
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
	ctx, span := enginerefactor.TraceReconstruct(ctx, im.instance.TelemetryInfo, "error")
	defer span.End()
	telemetry.ReportError(span, err)

	ctx = im.Context(ctx)

	msg := fmt.Sprintf("workflow failed with code = %v, type = %v, error = %v", typ, code, err.Error())
	telemetry.LogInstanceError(ctx, msg, err)
	telemetry.LogNamespaceError(im.Namespace().Name, msg, err)
}

func (engine *engine) UserLog(ctx context.Context, im *instanceMemory, msg string) {
	ctx, span := enginerefactor.TraceGet(ctx, im.instance.TelemetryInfo)
	span.AddEvent(msg)

	ctx = im.Context(ctx)
	telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
		strings.Trim(msg, "\""))
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
