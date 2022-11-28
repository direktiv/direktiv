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
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/senseyeio/duration"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/functions"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/util"

	entvardata "github.com/direktiv/direktiv/pkg/flow/ent/vardata"
	entvar "github.com/direktiv/direktiv/pkg/flow/ent/varref"
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
	As         string
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
		if derrors.IsNotFound(err) {
			return nil, derrors.NewUncatchableError("direktiv.workflow.notfound", "workflow not found: %v", err.Error())
		}
		return nil, err
	}

	var wf model.Workflow
	err = wf.Load(d.rev().Source)
	if err != nil {
		return nil, derrors.NewUncatchableError("direktiv.workflow.invalid", "cannot parse workflow '%s': %v", args.Path, err)
	}

	if len(wf.GetStartDefinition().GetEvents()) > 0 {
		if strings.ToLower(args.Caller) == "api" {
			return nil, derrors.NewUncatchableError("direktiv.workflow.invoke", "cannot manually invoke event-based workflow")
		}
		if strings.HasPrefix(args.Caller, "instance") {
			return nil, derrors.NewUncatchableError("direktiv.workflow.invoke", "cannot invoke event-based workflow as a subflow")
		}
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

	in, err := inc.Create().SetNamespace(d.ns()).SetWorkflow(d.wf).SetRevision(d.rev()).SetRuntime(rt).SetStatus(util.InstanceStatusPending).SetInvoker(args.Caller).SetAs(util.SanitizeAsField(as)).Save(ctx)
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
	im.engine = engine
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

	ctx, err = traceFullAddWorkflowInstance(ctx, d, im)
	if err != nil {
		return nil, err
	}

	t := time.Now()
	engine.pubsub.NotifyInstances(d.ns())
	engine.logToNamespace(ctx, t, d.ns(), "Workflow '%s' has been triggered by %s.", args.Path, args.Caller)
	engine.logToWorkflow(ctx, t, d.wfData, "Instance '%s' created by %s.", im.ID().String(), args.Caller)
	engine.logToInstance(ctx, t, in, "Preparing workflow triggered by %s.", args.Caller)

	// Broadcast Event
	err = engine.flow.BroadcastInstance(BroadcastEventTypeInstanceStarted, ctx,
		broadcastInstanceInput{
			WorkflowPath: args.Path,
			InstanceID:   im.ID().String(),
			Caller:       args.Caller,
		}, d.ns())
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

	workflow, err := im.Model()
	if err != nil {
		engine.CrashInstance(ctx, im, derrors.NewUncatchableError(ErrCodeWorkflowUnparsable, "failed to parse workflow YAML: %v", err))
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
		return fmt.Errorf("workflow cannot resolve state: %s", stateID)
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

	rt := im.in.Edges.Runtime
	edges := rt.Edges

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

	if cerr, catchable := err.(*derrors.CatchableError); catchable {
		engine.sugar.Errorf("Instance failed with error '%s': %v", cerr.Code, err)
		engine.logToInstance(ctx, time.Now(), im.in, "Instance failed with error '%s': %s", cerr.Code, err.Error())
	} else if uerr, uncatchable := err.(*derrors.UncatchableError); uncatchable && uerr.Code != "" {
		engine.sugar.Errorf("Instance failed with uncatchable error '%s': %v", uerr.Code, err)
		engine.logToInstance(ctx, time.Now(), im.in, "Instance failed with uncatchable error '%s': %s", uerr.Code, err.Error())
	} else {
		engine.sugar.Errorf("Instance failed with uncatchable error: %v", err)
		engine.logToInstance(ctx, time.Now(), im.in, "Instance failed with uncatchable error: %s", err.Error())
	}

	err = engine.SetInstanceFailed(ctx, im, err)
	if err != nil {
		engine.sugar.Error(err)
	}

	if ns, err := im.in.Namespace(ctx); err == nil {
		broadcastErr := engine.flow.BroadcastInstance(BroadcastEventTypeInstanceFailed, ctx, broadcastInstanceInput{
			WorkflowPath: GetInodePath(im.in.As),
			InstanceID:   im.in.ID.String(),
		}, ns)
		if broadcastErr != nil {
			engine.sugar.Errorf("Failed to broadcast: %v", broadcastErr)
		}
	} else {
		engine.sugar.Errorf("Failed to start broadcast: %v", err)
	}

	engine.TerminateInstance(ctx, im)

}

func (engine *engine) setEndAt(im *instanceMemory) {

	ctx := context.Background()

	in, err := im.in.Update().SetEndAt(time.Now()).Save(ctx)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	in.Edges = im.in.Edges
	im.in = in

}

func (engine *engine) TerminateInstance(ctx context.Context, im *instanceMemory) {

	engine.setEndAt(im)

	if im.logic != nil {
		engine.metricsCompleteState(ctx, im, "", im.ErrorCode(), false)
	}
	engine.metricsCompleteInstance(ctx, im)
	engine.FreeInstanceMemory(im)
	engine.WakeInstanceCaller(ctx, im)

}

func (engine *engine) runState(ctx context.Context, im *instanceMemory, wakedata []byte, err error) {

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

	if cerr, ok := err.(*derrors.CatchableError); ok {

		_ = im.StoreData("error", cerr)

		for i, catch := range im.logic.ErrorDefinitions() {

			errRegex := catch.Error
			if errRegex == "*" {
				errRegex = ".*"
			}

			t := time.Now()

			matched, regErr := regexp.MatchString(errRegex, cerr.Code)
			if regErr != nil {
				engine.logToInstance(ctx, t, im.in, "Error catching regex failed to compile: %v", regErr)
			}

			if matched {

				engine.logToInstance(ctx, t, im.in, "State failed with error '%s': %s", cerr.Code, cerr.Message)
				engine.logToInstance(ctx, t, im.in, "Error caught by error definition %d: %s", i, catch.Error)

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

	engine.logToInstance(ctx, time.Now(), im.in, "Transforming state data.")

	x, err := jqObject(im.data, transition.Transform)
	if err != nil {
		return derrors.WrapCatchableError("unable to apply transform: %v", err)
	}

	im.data = x

	return nil

}

func (engine *engine) transitionState(ctx context.Context, im *instanceMemory, transition *states.Transition, errCode string) {

	if transition == nil {
		engine.InstanceYield(im)
		return
	}

	if transition.NextState != "" {
		engine.metricsCompleteState(ctx, im, transition.NextState, errCode, false)
		engine.sugar.Debugf("Instance transitioning to next state: %s -> %s", im.ID().String(), transition.NextState)
		engine.logToInstance(ctx, time.Now(), im.in, "Transitioning to next state: %s (%d).", transition.NextState, im.Step()+1)
		go engine.Transition(ctx, im, transition.NextState, 0)
		return
	}

	status := util.InstanceStatusComplete
	if im.ErrorCode() != "" {
		status = util.InstanceStatusFailed
		engine.sugar.Debugf("Instance failed: %s", im.ID().String())
		engine.logToInstance(ctx, time.Now(), im.in, "Workflow failed with error '%s': %s", im.ErrorCode(), im.in.ErrorMessage)
	}

	engine.sugar.Debugf("Instance terminated: %s", im.ID().String())

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

	if ns, err := im.in.Namespace(ctx); err == nil {
		engine.pubsub.NotifyInstances(ns)
		broadcastErr := engine.flow.BroadcastInstance(BroadcastEventTypeInstanceSuccess, ctx, broadcastInstanceInput{
			WorkflowPath: GetInodePath(im.in.As),
			InstanceID:   im.in.ID.String(),
		}, ns)
		if broadcastErr != nil {
			engine.sugar.Errorf("Failed to broadcast: %v", broadcastErr)
		}
	} else {
		engine.sugar.Errorf("Failed to start broadcast: %v", err)
	}

	engine.TerminateInstance(ctx, im)

}

func (engine *engine) subflowInvoke(ctx context.Context, caller *subflowCaller, ns *ent.Namespace, name string, input []byte) (*instanceMemory, error) {

	var err error

	elems := strings.SplitN(name, ":", 2)

	args := new(newInstanceArgs)
	args.Namespace = ns.Name
	args.Path = elems[0]
	if len(elems) == 2 {
		args.Ref = elems[1]
	}

	args.Input = input
	args.Caller = fmt.Sprintf("instance:%v", caller.InstanceID) // TODO: human readable

	var threadVars []*ent.VarRef

	callerData, err := json.Marshal(caller)
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}
	args.CallerData = string(callerData)

	parent, err := engine.getInstance(ctx, engine.db.Namespace, ns.Name, caller.InstanceID, false)
	if err != nil {
		return nil, derrors.NewInternalError(err)
	}

	threadVars, err = parent.in.QueryVars().Where(entvar.BehaviourEQ("thread")).All(ctx)
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

	for _, tv := range threadVars {
		vd, err := tv.QueryVardata().Select(entvardata.FieldID).Only(ctx)
		if err != nil {
			return nil, derrors.NewInternalError(err)
		}
		err = engine.db.VarRef.Create().SetBehaviour("thread").SetInstance(im.in).SetName(tv.Name).SetVardataID(vd.ID).Exec(ctx)
		if err != nil {
			return nil, derrors.NewInternalError(err)
		}
	}

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

	engine.logToInstance(ctx, time.Now(), im.in, "Waking up to retry.")

	engine.sugar.Debugf("Handling retry wakeup: %s", this())

	go engine.runState(ctx, im, []byte(msg.Data), nil)

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
	if actionTimeout := (time.Duration(ar.Workflow.Timeout) * time.Second); actionTimeout > engine.conf.GetFunctionsTimeout() {
		_, err := engine.internal.ActionLog(context.Background(), &grpc.ActionLogRequest{
			InstanceId: ar.Workflow.InstanceID, Msg: []string{fmt.Sprintf("Warning: Action timeout '%v' is longer than max allowed duariton '%v'", actionTimeout, engine.conf.GetFunctionsTimeout())},
		})
		if err != nil {
			engine.sugar.Errorf("Failed to log: %v.", err)
		}
	}

	// TODO: should this ctx be modified with a shorter deadline?
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
	ar *functionRequest) {

	var (
		err error
	)

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
	engine.sugar.Debugf("function request: %v", addr)

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

	var (
		resp *http.Response
	)

	// potentially dns error for a brand new service
	// we just loop and see if we can recreate the service
	// one minute wait max
	cleanup := util.TraceHTTPRequest(ctx, req)
	defer cleanup()
	for i := 0; i < 180; i++ {
		engine.sugar.Debugf("functions request (%d): %v", i, addr)
		resp, err = client.Do(req)
		if err != nil {
			if ctxErr := rctx.Err(); ctxErr != nil {
				engine.sugar.Debugf("context error in knative call")
				return
			}
			engine.sugar.Debugf("error in request: %v", err)
			if err, ok := err.(*url.Error); ok {
				if err, ok := err.Err.(*net.OpError); ok {
					if _, ok := err.Err.(*net.DNSError); ok {

						// recreate if the service does not exist
						if ar.Container.Type == model.ReusableContainerFunctionType &&
							!engine.isKnativeFunction(engine.actions.client, ar) {
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
									engine.reportError(ar, fmt.Errorf("knative function: '%s' does not exist", ar.Container.Service))
									return
								}

								engine.sugar.Errorf("can not create scoped knative function: %v", err)
								engine.reportError(ar, err)
								return
							}
						}

						time.Sleep(1000 * time.Millisecond)
						continue
					}
				}
			}

			time.Sleep(1000 * time.Millisecond)

		} else {
			break
		}
	}

	if err != nil {
		engine.reportError(ar, err)
		return
	}

	if resp.StatusCode != 200 {
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
		err = fmt.Errorf("cannot load workflow logic instance: %v", err)
		engine.sugar.Error(err)
		return
	}

	wakedata, err := json.Marshal(events)
	if err != nil {
		err = fmt.Errorf("cannot marshal the action results payload: %v", err)
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

	d, err := engine.reverseTraverseToWorkflow(ctx, workflowID)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

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
	args.Namespace = d.namespace()
	args.Path = d.path

	args.Input = input
	args.Caller = "cloudevent" // TODO: human readable

	// TODO: TRACE traceEventsInvoked

	im, err := engine.NewInstance(ctx, args)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	engine.queue(im)

}
