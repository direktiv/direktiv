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

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/nohome"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/google/uuid"
	"github.com/senseyeio/duration"
	"go.opentelemetry.io/otel/trace"
)

type engine struct {
	*server

	scheduled sync.Map
}

func initEngine(srv *server) (*engine, error) {
	engine := new(engine)

	engine.server = srv

	engine.pBus.Subscribe(engine.instanceMessagesChannelHandler, engineInstanceMessagesChannel)

	go engine.instanceKicker()

	return engine, nil
}

func (engine *engine) Close() error {
	return nil
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

	slog.Debug("Starting to kick waiting (homeless) instances.")
	tx, err := engine.beginSqlTx(ctx)
	if err != nil {
		slog.Error("Failed to begin SQL transaction in kickWaitingInstances.", "error", err)
		return
	}
	defer tx.Rollback()

	instances, err := tx.InstanceStore().GetHomelessInstances(ctx, time.Now().UTC().Add(-engineSchedulingTimeout))
	if err != nil {
		slog.Error("Failed to list homeless instances in kickWaitingInstances. Some instances may remain unprocessed.", "error", err)
		return
	}
	if len(instances) == 0 {
		slog.Debug("No homeless instances found to kick.")
		return
	}
	slog.Info("Processing homeless instances.", "count", len(instances))

	for idx := range instances {
		instance := instances[idx]
		slog.Debug("Kicking instance.", "instance_id", instance.ID)

		data, err := json.Marshal(&instanceMessageChannelData{
			InstanceID:        instance.ID,
			LastKnownServer:   instance.Server,
			LastKnownUpdateAt: instance.UpdatedAt,
		})
		if err != nil {
			slog.Error("Failed to marshal instance data in kickWaitingInstances.", "instance", instance.ID, "error", err)
		}

		engine.instanceMessagesChannelHandler(string(data))
	}
}

type newInstanceArgs struct {
	tx            *database.SQLStore
	ID            uuid.UUID
	Namespace     *datastore.Namespace
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
	slog.Debug("Initializing new instance creation.", "namespace", args.Namespace.Name, "workflow", args.CalledAs, "invoker", args.Invoker)
	file, data, err := engine.mux(ctx, args.Namespace, args.CalledAs)
	loggingCtx := args.Namespace.WithTags(ctx)
	loggingCtx = enginerefactor.AddTag(loggingCtx, "calledAs", args.CalledAs)
	loggingCtx = enginerefactor.AddTag(loggingCtx, "invoker", args.Invoker)
	loggingCtx = enginerefactor.WithTrack(loggingCtx, enginerefactor.BuildNamespaceTrack(args.Namespace.Name))

	if err != nil {
		slog.Debug("Failed to retrieve workflow for new instance.", enginerefactor.GetSlogAttributesWithError(loggingCtx, err)...)
		if derrors.IsNotFound(err) {
			return nil, derrors.NewUncatchableError("direktiv.workflow.notfound", "workflow not found: %v", err.Error())
		}
		return nil, err
	}

	var wf model.Workflow
	err = wf.Load(data)
	if err != nil {
		slog.Error("Failed to parse workflow definition.", enginerefactor.GetSlogAttributesWithError(loggingCtx, err)...)
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

	settings := &enginerefactor.InstanceSettings{}
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

	tx := args.tx

	if tx == nil {
		tx, err = engine.flow.beginSqlTx(ctx)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()
	}
	slog.Debug("Preparing to commit new instance transaction.", "instance", args.ID.String(), "namespace", args.Namespace)

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
	slog.Info("New instance transaction committed successfully.", "instance", args.ID.String(), "namespace", args.Namespace)

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

	_, err = traceFullAddWorkflowInstance(ctx, im) // TODO.
	if err != nil {
		return nil, fmt.Errorf("failed to traceFullAddWorkflowInstance: %w", err)
	}
	im.AddAttribute("loop-index", fmt.Sprintf("%d", iterator))

	engine.pubsub.NotifyInstances(im.Namespace())
	namespaceTrackCtx := enginerefactor.WithTrack(loggingCtx, enginerefactor.BuildNamespaceTrack(im.instance.Instance.Namespace))
	slog.Info("Workflow has been triggered", enginerefactor.GetSlogAttributesWithStatus(namespaceTrackCtx, core.LogRunningStatus)...)

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
	workflow, err := im.Model()
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return nil
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

	ctx, cleanup, err := traceStateGenericBegin(ctx, im)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return nil
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

	engine.freeArtefacts(im)
	err := engine.freeMemory(ctx, im)
	if err != nil {
		if !engine.GetIsInstanceCrashed(im) {
			engine.CrashInstance(ctx, im, err)
			return
		}

		engine.forceFreeCriticalMemory(ctx, im)
		slog.Debug("Failed to free memory during a crash", "error", err)
	}

	if im.logic != nil {
		engine.metricsCompleteState(im, "", im.ErrorCode(), false)
	}

	engine.metricsCompleteInstance(im)

	engine.WakeInstanceCaller(ctx, im)
}

func (engine *engine) runState(ctx context.Context, im *instanceMemory, wakedata []byte, err error) *states.Transition {
	loggingCtx := im.Namespace().WithTags(ctx)
	instanceTrackCtx := enginerefactor.WithTrack(loggingCtx, enginerefactor.BuildInstanceTrack(im.instance))
	slog.Debug("Starting state execution.", enginerefactor.GetSlogAttributesWithStatus(ctx, core.LogRunningStatus)...)

	engine.logRunState(ctx, im, wakedata, err)
	slog.Debug("State logic executed. Processing post-execution actions.", enginerefactor.GetSlogAttributesWithStatus(instanceTrackCtx, core.LogRunningStatus)...)
	slog.Debug("Processing post-execution actions.", enginerefactor.GetSlogAttributesWithStatus(instanceTrackCtx, core.LogRunningStatus)...)

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
			slog.Error("Failed to process jq query on state data.", enginerefactor.GetSlogAttributesWithError(instanceTrackCtx, fmt.Errorf("query failed %v, err: %w", lq, err))...)

			goto failure
		}

		var data []byte
		data, err = json.MarshalIndent(object, "", "  ")
		if err != nil {
			err = derrors.NewInternalError(fmt.Errorf("failed to marshal state data: %w", err))
			slog.Error("Failed to marshal jq query result for logging.", enginerefactor.GetSlogAttributesWithError(instanceTrackCtx, fmt.Errorf("failed to marshal state data: %w", err))...)

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
	slog.Debug("Executing state logic.", enginerefactor.GetSlogAttributesWithStatus(instanceTrackCtx, core.LogRunningStatus)...)

	transition, err = im.logic.Run(ctx, wakedata)
	if err != nil {
		goto failure
	}
	slog.Debug("Applying state transformation based on logic run.", enginerefactor.GetSlogAttributesWithStatus(instanceTrackCtx, core.LogRunningStatus)...)

	err = engine.transformState(ctx, im, transition)
	if err != nil {
		goto failure
	}

next:
	slog.Debug("Initiating state logic run.", enginerefactor.GetSlogAttributesWithStatus(instanceTrackCtx, core.LogRunningStatus)...)
	return engine.transitionState(ctx, im, transition, code)

failure:

	traceStateError(ctx, err)

	var breaker int

	if breaker > 10 {
		err = derrors.NewInternalError(errors.New("somehow ended up in a catchable error loop"))
	}

	err1 := engine.CancelInstanceChildren(ctx, im)
	if err1 != nil {
		slog.Error("Canceling Instance's chrildren failed", enginerefactor.GetSlogAttributesWithError(instanceTrackCtx, err1)...)
	}
	cerr := new(derrors.CatchableError)

	if errors.As(err, &cerr) {
		_ = im.StoreData("error", cerr)

		for _, catch := range im.logic.ErrorDefinitions() {
			errRegex := catch.Error
			if errRegex == "*" {
				errRegex = ".*"
			}

			matched, regErr := regexp.MatchString(errRegex, cerr.Code)
			if regErr != nil {
				slog.Error("Regex compilation failed for error catch definition.", enginerefactor.GetSlogAttributesWithError(instanceTrackCtx, regErr)...)
			}

			if matched {
				slog.Info("Catchable error matched; executing defined transition.", enginerefactor.GetSlogAttributesWithStatus(instanceTrackCtx, core.LogErrStatus)...)
				slog.Error("State failed with an error", enginerefactor.GetSlogAttributesWithError(instanceTrackCtx, fmt.Errorf("State failed with an error '%s': %s", cerr.Code, cerr.Message))...)

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
	slog.Error("Unrecoverable error encountered; initiating instance crash.", enginerefactor.GetSlogAttributesWithError(instanceTrackCtx, err)...)
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
	loggingCtx := im.Namespace().WithTags(ctx)
	slog.Debug("Transforming state data.",
		enginerefactor.GetSlogAttributesWithStatus(
			enginerefactor.WithTrack(im.WithTags(loggingCtx),
				enginerefactor.BuildInstanceTrack(im.instance)), core.LogRunningStatus,
		)...)

	x, err := jqObject(im.data, transition.Transform)
	if err != nil {
		return derrors.WrapCatchableError("unable to apply transform: %v", err)
		// return derrors.WrapCatchableError("Failed to apply jq transformation on state data. Transformation: '%v', Error: %v", transition.Transform, err)
	}

	im.replaceData(x)
	slog.Debug("Successfully transformed state data.",
		enginerefactor.GetSlogAttributesWithStatus(
			enginerefactor.WithTrack(im.WithTags(loggingCtx),
				enginerefactor.BuildInstanceTrack(im.instance)), core.LogRunningStatus,
		)...)

	return nil
}

func (engine *engine) transitionState(ctx context.Context, im *instanceMemory, transition *states.Transition, errCode string) *states.Transition {
	e := im.flushUpdates(ctx)
	if e != nil {
		slog.Error("Failed to flush updates for instance.", "instance", im.ID(), "namespace", im.Namespace(), "error", e)
		engine.CrashInstance(ctx, im, e)
		return nil
	}

	if transition == nil {
		engine.InstanceYield(ctx, im)
		return nil
	}
	loggingCtx := im.Namespace().WithTags(ctx)
	instanceTrackCtx := enginerefactor.WithTrack(im.WithTags(loggingCtx), enginerefactor.BuildInstanceTrack(im.instance))

	if transition.NextState != "" {
		engine.metricsCompleteState(im, transition.NextState, errCode, false)
		slog.Debug("Transitioning to next state.",
			enginerefactor.GetSlogAttributesWithStatus(instanceTrackCtx, core.LogRunningStatus)...)

		return transition
	}

	status := instancestore.InstanceStatusComplete
	if im.ErrorCode() != "" {
		status = instancestore.InstanceStatusFailed
		if im.ErrorCode() == "direktiv.cancels.parent" || im.ErrorCode() == "direktiv.cancels.api" {
			status = instancestore.InstanceStatusCancelled
		}
		slog.Debug("Workflow failed with an error.", enginerefactor.GetSlogAttributesWithError(instanceTrackCtx, fmt.Errorf("'%s': %s", im.ErrorCode(), im.instance.Instance.ErrorMessage))...)
	}

	slog.Debug("Instance terminated", "instance", im.ID().String(), "namespace", im.Namespace())

	output := im.MarshalData()
	im.instance.Instance.Output = []byte(output)
	im.updateArgs.Output = &im.instance.Instance.Output
	im.instance.Instance.Status = status
	im.updateArgs.Status = &im.instance.Instance.Status

	slog.Info("Workflow completed.", enginerefactor.GetSlogAttributesWithStatus(instanceTrackCtx, core.LogCompletedStatus)...)

	defer engine.pubsub.NotifyInstance(im.instance.Instance.ID)
	defer engine.pubsub.NotifyInstances(im.Namespace())

	engine.TerminateInstance(ctx, im)

	return nil
}

func (engine *engine) subflowInvoke(ctx context.Context, pi *enginerefactor.ParentInfo, instance *enginerefactor.Instance, name string, input []byte) (*instanceMemory, error) {
	var err error

	di := &enginerefactor.InstanceDescentInfo{
		Descent: append(instance.DescentInfo.Descent, *pi),
	}

	span := trace.SpanFromContext(ctx)

	args := &newInstanceArgs{
		ID: uuid.New(),
		Namespace: &datastore.Namespace{
			ID:   instance.Instance.NamespaceID,
			Name: instance.TelemetryInfo.NamespaceName,
		},
		CalledAs:    name,
		Input:       input,
		Invoker:     fmt.Sprintf("instance:%v", pi.ID),
		DescentInfo: di,
		TelemetryInfo: &enginerefactor.InstanceTelemetryInfo{
			TraceID:       span.SpanContext().TraceID().String(),
			SpanID:        span.SpanContext().SpanID().String(),
			NamespaceName: instance.TelemetryInfo.NamespaceName,
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
	traceSubflowInvoke(ctx, args.CalledAs, im.ID().String())

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

func (engine *engine) wakeEventsWaiter(instance uuid.UUID, events []*cloudevents.Event) {
	ctx := context.Background()

	err := engine.enqueueInstanceMessage(ctx, instance, "event", events)
	if err != nil {
		slog.Error("failed to enqueue instance message for wakeEventsWaiter", "error", err)
		return
	}
}

func (engine *engine) EventsInvoke(workflowID uuid.UUID, events ...*cloudevents.Event) {
	ctx := context.Background()

	tx, err := engine.flow.beginSqlTx(ctx)
	if err != nil {
		slog.Error("Failed to begin SQL transaction in EventsInvoke.", "error", err)
		return
	}
	defer tx.Rollback()

	file, err := tx.FileStore().GetFileByID(ctx, workflowID)
	if err != nil {
		slog.Error("Failed to fetch file from database.", "workflowID", workflowID, "error", err)
		return
	}

	root, err := tx.FileStore().GetRoot(ctx, file.RootID)
	if err != nil {
		slog.Error("Failed to fetch Root from database.", "workflowID", workflowID, "error", err)
		return
	}

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, root.Namespace)
	if err != nil {
		slog.Error("Failed to fetch namespace from database.", "namespace", root.Namespace, "error", err)
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
		slog.Error("Failed to marshal event data in EventsInvoke.", "error", err)
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
			TraceID:       span.SpanContext().TraceID().String(),
			SpanID:        span.SpanContext().SpanID().String(),
			NamespaceName: ns.Name,
		},
	}

	im, err := engine.NewInstance(ctx, args)
	if err != nil {
		slog.Error("new instance", "error", err)
		return
	}
	slog.Debug("Successfully invoked new workflow instance.", "instanceID", im.ID().String(), "workflowPath", file.Path)

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
	loggingCtx := im.Namespace().WithTags(ctx)

	instanceTrackCtx := enginerefactor.WithTrack(im.WithTags(loggingCtx), enginerefactor.BuildInstanceTrack(im.instance))
	namespaceTrackCtx := enginerefactor.WithTrack(im.WithTags(loggingCtx), enginerefactor.BuildNamespaceTrack(im.Namespace().Name))
	msg := fmt.Sprintf("Workflow failed with code = %v, error = %v", typ, code)
	slog.Error(msg, enginerefactor.GetSlogAttributesWithError(instanceTrackCtx, err)...)
	slog.Error(msg, enginerefactor.GetSlogAttributesWithError(namespaceTrackCtx, err)...)
}

func (engine *engine) UserLog(ctx context.Context, im *instanceMemory, msg string, a ...interface{}) {
	loggingCtx := im.Namespace().WithTags(ctx)
	instanceTrackCtx := enginerefactor.WithTrack(im.WithTags(loggingCtx), enginerefactor.BuildInstanceTrack(im.instance))

	slog.Info(fmt.Sprintf(msg, a...), enginerefactor.GetSlogAttributesWithStatus(instanceTrackCtx, core.LogUnknownStatus)...)

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
			slog.Error("failed to create cloudevent", "error", err.Error())
		}

		err = engine.events.BroadcastCloudevent(ctx, im.Namespace(), &event, 0)
		if err != nil {
			slog.Error("failed to broadcast cloudevent", "error", err)
			return
		}
	}
}

func (engine *engine) logRunState(ctx context.Context, im *instanceMemory, wakedata []byte, err error) {
	slog.Debug("Running state logic", "instance", im.ID().String(), "step", im.Step(), "logic", im.logic.GetID())
	if im.GetMemory() == nil && len(wakedata) == 0 && err == nil {
		loggingCtx := im.Namespace().WithTags(ctx)
		instanceTrackCtx := enginerefactor.WithTrack(im.WithTags(loggingCtx), enginerefactor.BuildInstanceTrack(im.instance))

		slog.Info("Running state logic", enginerefactor.GetSlogAttributesWithStatus(instanceTrackCtx, core.LogRunningStatus)...)
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
