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
	"regexp"
	"strings"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/senseyeio/duration"

	"github.com/vorteil/direktiv/pkg/flow/ent"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"github.com/vorteil/direktiv/pkg/functions"
	"github.com/vorteil/direktiv/pkg/model"
	"github.com/vorteil/direktiv/pkg/util"
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
		model.StateTypeNoop:          initNoopStateLogic,
		model.StateTypeAction:        initActionStateLogic,
		model.StateTypeConsumeEvent:  initConsumeEventStateLogic,
		model.StateTypeDelay:         initDelayStateLogic,
		model.StateTypeError:         initErrorStateLogic,
		model.StateTypeEventsAnd:     initEventsAndStateLogic,
		model.StateTypeEventsXor:     initEventsXorStateLogic,
		model.StateTypeForEach:       initForEachStateLogic,
		model.StateTypeGenerateEvent: initGenerateEventStateLogic,
		model.StateTypeParallel:      initParallelStateLogic,
		model.StateTypeSwitch:        initSwitchStateLogic,
		model.StateTypeValidate:      initValidateStateLogic,
		model.StateTypeGetter:        initGetterStateLogic,
		model.StateTypeSetter:        initSetterStateLogic,
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

	ctx, err := engine.InstanceLock(im, defaultLockWait)
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

func (engine *engine) loadStateLogic(im *instanceMemory, stateID string) error {

	workflow, err := im.Model()
	if err != nil {
		return err
	}

	states := workflow.GetStatesMap()
	state, exists := states[stateID]
	if !exists {
		return fmt.Errorf("workflow cannot resolve state: %s", stateID)
	}

	init, exists := engine.stateLogics[state.GetType()]
	if !exists {
		return fmt.Errorf("engine cannot resolve state type: %s", state.GetType().String())
	}

	stateLogic, err := init(workflow, state)
	if err != nil {
		return fmt.Errorf("cannot initialize state logic: %v", err)
	}
	im.logic = stateLogic

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

	// TODO: I don't think this is actually possible?
	// if len(wli.rec.Flow) != wli.step {
	// 	err := errors.New("workflow logic instance aborted for being tardy")
	// 	engine.sugar.Error(err)
	// 	wli.Close()
	// 	return
	// }

	if nextState == "" {
		panic("don't call this function with an empty nextState")
	}

	err = engine.loadStateLogic(im, nextState)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return
	}

	flow := append(im.Flow(), nextState)
	deadline := im.logic.Deadline(ctx, engine, im)

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

	engine.sugar.Errorf("Instance failed with uncatchable error: %v", err)

	engine.logToInstance(ctx, time.Now(), im.in, "Instance failed with uncatchable error: %s", err.Error())

	err = engine.SetInstanceFailed(ctx, im, err)
	if err != nil {
		engine.sugar.Error(err)
	}

	engine.TerminateInstance(ctx, im)

}

func (engine *engine) TerminateInstance(ctx context.Context, im *instanceMemory) {

	engine.WakeInstanceCaller(ctx, im)
	engine.metricsCompleteState(ctx, im, "", im.ErrorCode(), false)
	engine.metricsCompleteInstance(ctx, im)
	engine.FreeInstanceMemory(im)

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

const maxSubflowDepth = 5

func (engine *engine) subflowInvoke(ctx context.Context, caller *subflowCaller, ns *ent.Namespace, name string, input []byte) (string, error) {

	var err error

	elems := strings.SplitN(name, ":", 2)

	args := new(newInstanceArgs)
	args.Namespace = ns.Name
	args.Path = elems[0]
	if len(elems) == 2 {
		args.Ref = elems[1]
	}

	args.Input = input
	args.Caller = "workflow" // TODO: human readable

	if caller != nil {

		callerData, err := json.Marshal(caller)
		if err != nil {
			return "", NewInternalError(err)
		}
		args.CallerData = string(callerData)

	}

	im, err := engine.NewInstance(ctx, args)
	if err != nil {
		engine.sugar.Debugf("Error returned to gRPC request %s: %v", this(), err)
		return "", err
	}

	engine.queue(im)

	return im.ID().String(), nil

}

type retryMessage struct {
	InstanceID string
	State      string
	Step       int
	Data       []byte
}

const retryWakeupFunction = "retryWakeup"

func (engine *engine) scheduleRetry(id, state string, step int, t time.Time, data []byte) error {

	data, _ = json.Marshal(&retryMessage{
		InstanceID: id,
		State:      state,
		Step:       step,
		Data:       data,
	})

	if d := t.Sub(time.Now()); d < time.Second*5 {
		go func() {
			time.Sleep(d)
			/* #nosec */
			_ = engine.retryWakeup(data)
		}()
		return nil
	}

	err := engine.timers.addOneShot(id, retryWakeupFunction, t, data)
	if err != nil {
		return NewInternalError(err)
	}

	return nil

}

func (engine *engine) retryWakeup(data []byte) error {

	msg := new(retryMessage)

	err := json.Unmarshal(data, msg)
	if err != nil {
		engine.sugar.Error(err)
		return nil
	}

	ctx, im, err := engine.loadInstanceMemory(msg.InstanceID, msg.Step)
	if err != nil {
		engine.sugar.Error(err)
		return nil
	}

	engine.logToInstance(ctx, time.Now(), im.in, "Waking up to retry.")

	engine.sugar.Debugf("Handling retry wakeup: %s", this())

	go engine.runState(ctx, im, []byte(msg.Data), nil)

	return nil

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

	// TODO: should this ctx be modified with a shorter deadline?
	switch ar.Container.Type {
	case model.IsolatedContainerFunctionType:
		hostname, ip, err := engine.addPodFunction(ctx, engine.actions.client, ar)
		if err != nil {
			return NewInternalError(err)
		}

		go func(ar *functionRequest) {
			// post data
			engine.doPodHTTPRequest(ctx, ar, hostname, ip)
		}(ar)

	case model.DefaultFunctionType:
		fallthrough
	case model.NamespacedKnativeFunctionType:
		fallthrough
	case model.GlobalKnativeFunctionType:
		fallthrough
	case model.ReusableContainerFunctionType:
		go engine.doKnativeHTTPRequest(ctx, ar)
	}

	return nil

}

func (engine *engine) doPodHTTPRequest(ctx context.Context,
	ar *functionRequest, hostname, ip string) {

	// useTLS := engine.conf.FunctionsProtocol == "https"
	//
	// tr := engine.createTransport(useTLS)

	// configured namespace for workflows
	addr := fmt.Sprintf("http://%s:8890", ip)
	// if useTLS {
	// 	addr = fmt.Sprintf("%s://%s:8890", engine.conf.FunctionsProtocol, hostname)
	// }

	engine.sugar.Debugf("function request: %v", addr)

	now := time.Now()
	deadline := now.Add(time.Duration(ar.Workflow.Timeout) * time.Second)
	rctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	engine.sugar.Debugf("deadline for pod request: %v", deadline.Sub(now))

	req, err := http.NewRequestWithContext(rctx, http.MethodPost, addr,
		bytes.NewReader(ar.Container.Data))
	if err != nil {
		engine.reportError(ar, err)
		return
	}

	for i := range ar.Container.Files {
		f := &ar.Container.Files[i]
		data, err := json.Marshal(f)
		if err != nil {
			engine.reportError(ar, err)
		}
		str := base64.StdEncoding.EncodeToString(data)
		req.Header.Add(DirektivFileHeader, str)
	}

	client := &http.Client{}

	var (
		resp *http.Response
	)

	resp, err = client.Do(req)

	if err != nil {
		if ctxErr := rctx.Err(); ctxErr != nil {
			engine.sugar.Debugf("context error in pod call")
			return
		}
		engine.reportError(ar, err)
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

func (engine *engine) doKnativeHTTPRequest(ctx context.Context,
	ar *functionRequest) {

	var (
		err error
	)

	tr := engine.createTransport()

	// configured namespace for workflows
	ns := os.Getenv(util.DirektivServiceNamespace)

	// set service name if global/namespace
	// otherwise generate baes on action request
	svn := ar.Container.Service
	if ar.Container.Type == model.ReusableContainerFunctionType {
		svn, _, err = functions.GenerateServiceName(ar.Workflow.Namespace,
			ar.Workflow.ID, ar.Container.ID)
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

	engine.sugar.Debugf("deadline for request: %v", deadline.Sub(time.Now()))

	req, err := http.NewRequestWithContext(rctx, http.MethodPost, addr,
		bytes.NewReader(ar.Container.Data))
	if err != nil {
		engine.reportError(ar, err)
		return
	}

	// add headers
	req.Header.Add(DirektivDeadlineHeader, deadline.Format(time.RFC3339))
	req.Header.Add(DirektivNamespaceHeader, ar.Workflow.Namespace)
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
	for i := 0; i < 60; i++ {
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

						// we can recreate the function if it is a workflow scope function
						// if not we can bail right here
						if ar.Container.Type != model.ReusableContainerFunctionType {
							engine.reportError(ar,
								fmt.Errorf("function %s does not exist on scope %v",
									ar.Container.ID, ar.Container.Type))
							return
						}

						// recreate if the service does not exist
						if ar.Container.Type == model.ReusableContainerFunctionType &&
							!engine.isKnativeFunction(engine.actions.client, ar.Container.ID,
								ar.Workflow.Namespace, ar.Workflow.ID) {
							err := createKnativeFunction(engine.actions.client, ar)
							if err != nil && !strings.Contains(err.Error(), "already exists") {
								engine.sugar.Errorf("can not create knative function: %v", err)
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

const eventsWakeupFunction = "eventsWakeup"

func (engine *engine) wakeEventsWaiter(signature []byte, events []*cloudevents.Event) {

	sig := new(eventsWaiterSignature)
	err := json.Unmarshal(signature, sig)
	if err != nil {
		err = NewInternalError(err)
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

	go engine.runState(ctx, im, wakedata, nil)

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

	im, err := engine.NewInstance(ctx, args)
	if err != nil {
		engine.sugar.Error(err)
		return
	}

	engine.queue(im)

}
