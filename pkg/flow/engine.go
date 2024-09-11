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
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/google/uuid"
	"github.com/senseyeio/duration"
)

type engine struct {
	*server

	scheduled sync.Map
}

func initEngine(srv *server) *engine {
	engine := new(engine)

	engine.server = srv

	engine.pBus.Subscribe(&pubsub.InstanceMessageEvent{}, engine.instanceMessagesChannelHandler)

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

	slog.Debug("Starting to kick waiting (homeless) instances.")
	tx, err := engine.beginSQLTx(ctx)
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
		slog.Debug("Kicking instance.", "instance", instance.ID)

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
	ctx = tracing.AddInstanceAttr(ctx, tracing.InstanceAttributes{
		Namespace:    args.Namespace.Name,
		InstanceID:   args.ID.String(),
		Invoker:      args.Invoker,
		Callpath:     args.TelemetryInfo.CallPath,
		WorkflowPath: args.CalledAs,
		Status:       core.LogRunningStatus,
	})
	ctx, cleanup, err2 := tracing.NewSpan(ctx, "creating a new Instance: "+args.ID.String()+", workflow: "+args.CalledAs)
	if err2 != nil {
		slog.Debug("failed in new instance", "error", err2)
	}
	defer cleanup()
	slog.DebugContext(ctx, "Initializing new instance creation.")
	file, data, err := engine.mux(ctx, args.Namespace, args.CalledAs)
	if err != nil {
		return nil, err
	}

	var wf model.Workflow
	err = wf.Load(data)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to parse workflow definition.", "error", err)
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
	traceParent, err := tracing.ExtractTraceParent(ctx)
	if err != nil {
		slog.Debug("NewInstance telemetry failed", "error", err)
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
	slog.DebugContext(ctx, "Preparing to commit new instance transaction.")

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
	slog.DebugContext(ctx, "New instance transaction committed successfully.")

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
	slog.InfoContext(ctx, "Workflow has been triggered")

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
	ctx, cleanup, err := tracing.NewSpan(ctx, "engine transitions: "+nextState)
	if err != nil {
		slog.Debug("transition failed to init telemetry", "error", err)
	}
	defer cleanup()
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
	ctx = tracing.AddInstanceMemoryAttr(ctx,
		tracing.InstanceAttributes{
			Namespace:    im.Namespace().Name,
			InstanceID:   im.GetInstanceID().String(),
			Invoker:      im.instance.Instance.Invoker,
			Callpath:     tracing.CreateCallpath(im.instance),
			WorkflowPath: im.instance.Instance.WorkflowPath,
			Status:       core.LogUnknownStatus,
		}, im.GetState())
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

	engine.freeArtefacts(im) //nolint:contextcheck
	err := engine.freeMemory(ctx, im)
	if err != nil {
		if !engine.GetIsInstanceCrashed(im) {
			engine.CrashInstance(ctx, im, err)
			return
		}

		engine.forceFreeCriticalMemory(ctx, im)
		slog.Debug("Failed to free memory during a crash", "error", err)
	}

	engine.WakeInstanceCaller(ctx, im)
}

//nolint:gocognit
func (engine *engine) runState(ctx context.Context, im *instanceMemory, wakedata []byte, err error) *states.Transition {
	ctx = tracing.AddNamespace(ctx, im.Namespace().Name)
	ctx = tracing.WithTrack(ctx, tracing.BuildInstanceTrack(im.instance))
	ctx, cleanup, err3 := tracing.NewSpan(ctx, "preparing instance for state execution")
	if err != nil {
		slog.Debug("failed to init telemery in runstate", "error", err3)
	}
	defer cleanup()

	slog.DebugContext(ctx, "Starting state execution.")

	var transition *states.Transition

	if err != nil {
		slog.ErrorContext(ctx, "Error before state execution.", "error", err)

		goto failure
	}

	if lq := im.logic.GetLog(); im.GetMemory() == nil && len(wakedata) == 0 && lq != nil {
		var object interface{}
		object, err = jqOne(im.data, lq) //nolint:contextcheck
		if err != nil {
			slog.ErrorContext(ctx, "Failed to process jq query on state data.", "error", fmt.Errorf("query failed %v, err: %w", lq, err))

			goto failure
		}

		var data []byte
		data, err = json.MarshalIndent(object, "", "  ")
		if err != nil {
			err = derrors.NewInternalError(fmt.Errorf("failed to marshal state data: %w", err))
			slog.ErrorContext(ctx, "Failed to marshal jq query result for logging.", "error", fmt.Errorf("failed to marshal state data: %w", err))

			goto failure
		}

		engine.UserLog(ctx, im, string(data))
	}

	if md := im.logic.GetMetadata(); im.GetMemory() == nil && len(wakedata) == 0 && md != nil {
		var object interface{}
		object, err = jqOne(im.data, md) //nolint:contextcheck
		if err != nil {
			slog.ErrorContext(ctx, "Failed to execute jq query for metadata.", "error", err)

			goto failure
		}

		var data []byte
		data, err = json.MarshalIndent(object, "", "  ")
		if err != nil {
			err = derrors.NewInternalError(fmt.Errorf("failed to marshal state data: %w", err))
			slog.ErrorContext(ctx, "Failed to marshal metadata.", "error", err)

			goto failure
		}

		engine.StoreMetadata(ctx, im, string(data))
	}
	ctx = tracing.AddInstanceMemoryAttr(ctx, tracing.InstanceAttributes{
		Namespace:    im.Namespace().Name,
		InstanceID:   im.GetInstanceID().String(),
		Invoker:      im.instance.Instance.Invoker,
		Callpath:     tracing.CreateCallpath(im.instance),
		WorkflowPath: im.instance.Instance.WorkflowPath,
		Status:       core.LogUnknownStatus,
	}, im.GetState())
	slog.InfoContext(ctx, "Running state logic.")

	transition, err = im.logic.Run(ctx, wakedata)
	if err != nil {
		slog.ErrorContext(ctx, "State logic execution failed.", "error", err)

		goto failure
	}
	slog.DebugContext(ctx, "Applying state transformation based on logic run.")

	err = engine.transformState(ctx, im, transition)
	if err != nil {
		slog.ErrorContext(ctx, "State transformation failed.", "error", err)

		goto failure
	}

	slog.DebugContext(ctx, "State logic executed. Processing post-execution actions.")

next:
	slog.DebugContext(ctx, "Processing post-execution actions.")

	return engine.transitionState(ctx, im, transition)

failure:
	slog.ErrorContext(ctx, "State execution failed.", "error", err)
	// traceStateError(ctx, err)

	var breaker int

	if breaker > 10 {
		err = derrors.NewInternalError(errors.New("somehow ended up in a catchable error loop"))
		slog.ErrorContext(ctx, "Possible error loop detected.", "error", err)
	}

	err1 := engine.CancelInstanceChildren(ctx, im)
	if err1 != nil {
		slog.ErrorContext(ctx, "Canceling Instance's children failed.", "error", err1)
	}
	cerr := new(derrors.CatchableError)

	if errors.As(err, &cerr) {
		err2 := im.StoreData("error", cerr)
		if err2 != nil {
			slog.ErrorContext(ctx, "Failed to store error data.", "error", err2)
		}

		for _, catch := range im.logic.ErrorDefinitions() {
			errRegex := catch.Error
			if errRegex == "*" {
				errRegex = ".*"
			}

			matched, regErr := regexp.MatchString(errRegex, cerr.Code)
			if regErr != nil {
				slog.ErrorContext(ctx, "Regex compilation failed for error catch definition.", "error", regErr)
			}

			if matched {
				slog.InfoContext(ctx, "Catchable error matched; executing defined transition.")
				slog.ErrorContext(ctx, "State failed with an error", "error", fmt.Errorf("state failed with an error '%s': %s", cerr.Code, cerr.Message))

				transition = &states.Transition{
					Transform: "",
					NextState: catch.Transition,
				}

				goto next
			}
		}
	}
	slog.ErrorContext(ctx, "Unrecoverable error encountered; initiating instance crash.", "error", err)
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
	loggingCtx := tracing.AddInstanceMemoryAttr(ctx, tracing.InstanceAttributes{
		Namespace:    im.instance.Instance.Namespace,
		InstanceID:   im.GetInstanceID().String(),
		Invoker:      im.instance.Instance.Invoker,
		Callpath:     tracing.CreateCallpath(im.instance),
		WorkflowPath: im.instance.Instance.WorkflowPath,
		Status:       core.LogUnknownStatus,
	}, im.GetState())
	loggingCtx = tracing.WithTrack(loggingCtx, tracing.BuildInstanceTrack(im.instance))
	slog.DebugContext(loggingCtx, "Transforming state data.")

	x, err := jqObject(im.data, transition.Transform) //nolint:contextcheck
	if err != nil {
		slog.ErrorContext(loggingCtx, "Failed to apply jq to transform.", "error", err)

		return derrors.WrapCatchableError("unable to apply transform: %v", err)
		// return derrors.WrapCatchableError("Failed to apply jq transformation on state data. Transformation: '%v', Error: %v", transition.Transform, err)
	}

	im.replaceData(x)
	slog.DebugContext(loggingCtx, "Successfully transformed state data.")

	return nil
}

func (engine *engine) transitionState(ctx context.Context, im *instanceMemory, transition *states.Transition) *states.Transition {
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
	loggingCtx := tracing.AddInstanceMemoryAttr(ctx, tracing.InstanceAttributes{
		Namespace:    im.Namespace().Name,
		InstanceID:   im.GetInstanceID().String(),
		Invoker:      im.instance.Instance.Invoker,
		Callpath:     tracing.CreateCallpath(im.instance),
		WorkflowPath: im.instance.Instance.WorkflowPath,
		Status:       core.LogUnknownStatus,
	}, im.GetState())
	instanceTrackCtx := tracing.WithTrack(loggingCtx, tracing.BuildInstanceTrack(im.instance))

	if transition.NextState != "" {
		slog.DebugContext(instanceTrackCtx, "Transitioning to next state.")

		return transition
	}

	status := instancestore.InstanceStatusComplete
	if im.ErrorCode() != "" {
		status = instancestore.InstanceStatusFailed
		if im.ErrorCode() == "direktiv.cancels.parent" || im.ErrorCode() == "direktiv.cancels.api" {
			status = instancestore.InstanceStatusCancelled
		}
		slog.ErrorContext(instanceTrackCtx, "Workflow failed with an error.", "error", fmt.Errorf("'%s': %s", im.ErrorCode(), im.instance.Instance.ErrorMessage))
	}

	slog.DebugContext(instanceTrackCtx, "Instance terminated", "instance", im.ID().String(), "namespace", im.Namespace().Name)

	output := im.MarshalData()
	im.instance.Instance.Output = []byte(output)
	im.updateArgs.Output = &im.instance.Instance.Output
	im.instance.Instance.Status = status
	im.updateArgs.Status = &im.instance.Instance.Status

	slog.InfoContext(instanceTrackCtx, "Workflow completed.")

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

	slog.InfoContext(ctx, "Invoking a subflow")

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
			TraceParent:   instance.TelemetryInfo.TraceParent,
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

	err := engine.enqueueInstanceMessage(ctx, instance, "event", events)
	if err != nil {
		slog.Error("failed to enqueue instance message for wakeEventsWaiter", "error", err)
		return
	}
}

func (engine *engine) EventsInvoke(tctx context.Context, workflowID uuid.UUID, events ...*cloudevents.Event) {
	ctx := context.Background()

	tx, err := engine.flow.beginSQLTx(ctx)
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
	tctx, end, err2 := tracing.NewSpan(tctx, "engine invoked by event")
	if err2 != nil {
		slog.Debug("Failed to tracing.NewSpan.", "error", err)
	}
	defer end()
	traceParent, err2 := tracing.ExtractTraceParent(tctx)
	if err2 != nil {
		slog.Debug("Failed to extract traceParent in EventsInvoke.", "error", err)
	}
	// TODO: tracing

	args := &newInstanceArgs{
		ID:        uuid.New(),
		Namespace: ns,
		CalledAs:  file.Path,
		Input:     input,
		Invoker:   "cloudevent",
		TelemetryInfo: &enginerefactor.InstanceTelemetryInfo{
			TraceParent:   traceParent,
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
	loggingCtx := tracing.AddInstanceMemoryAttr(ctx, tracing.InstanceAttributes{
		Namespace:    im.Namespace().Name,
		InstanceID:   im.GetInstanceID().String(),
		Invoker:      im.instance.Instance.Invoker,
		Callpath:     tracing.CreateCallpath(im.instance),
		WorkflowPath: im.instance.Instance.WorkflowPath,
		Status:       core.LogUnknownStatus,
	}, im.GetState())
	instanceTrackCtx := tracing.WithTrack(loggingCtx, tracing.BuildInstanceTrack(im.instance))

	namespaceTrackCtx := tracing.WithTrack(loggingCtx, tracing.BuildNamespaceTrack(im.Namespace().Name))
	msg := fmt.Sprintf("Workflow failed with code = %v, type = %v, error = %v", typ, code, err.Error())
	slog.ErrorContext(instanceTrackCtx, msg, "error", err)
	slog.ErrorContext(namespaceTrackCtx, msg, "error", err)
}

func (engine *engine) UserLog(ctx context.Context, im *instanceMemory, msg string) {
	loggingCtx := tracing.AddInstanceMemoryAttr(ctx, tracing.InstanceAttributes{
		Namespace:    im.Namespace().Name,
		InstanceID:   im.GetInstanceID().String(),
		Invoker:      im.instance.Instance.Invoker,
		Callpath:     tracing.CreateCallpath(im.instance),
		WorkflowPath: im.instance.Instance.WorkflowPath,
		Status:       core.LogUnknownStatus,
	}, im.GetState())
	instanceTrackCtx := tracing.WithTrack(loggingCtx, tracing.BuildInstanceTrack(im.instance))
	slog.InfoContext(instanceTrackCtx, msg)
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
