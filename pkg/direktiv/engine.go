package direktiv

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/vorteil/direktiv/ent/workflowinstance"
	"github.com/vorteil/direktiv/pkg/dlog/dummy"
	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/metrics"
	secretsgrpc "github.com/vorteil/direktiv/pkg/secrets/grpc"
	"google.golang.org/grpc"

	"github.com/jinzhu/copier"
	"github.com/vorteil/direktiv/pkg/flow"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/pkg/dlog"
	"github.com/vorteil/direktiv/pkg/model"
)

const (
	// WorkflowStateSubscription is the channel that runs workflow states.
	WorkflowStateSubscription = "workflow-state"
)

// internal errors
var (
	ErrCodeJQBadQuery        = "direktiv.jq.badCommand"
	ErrCodeJQNotObject       = "direktiv.jq.notObject"
	ErrCodeMultipleErrors    = "direktiv.workflow.multipleErrors"
	ErrCodeAllBranchesFailed = "direktiv.parallel.allFailed"
)

type workflowEngine struct {
	db             *dbManager
	timer          *timerManager
	instanceLogger *dlog.Log
	stateLogics    map[model.StateType]func(*model.Workflow, model.State) (stateLogic, error)
	server         *WorkflowServer

	cancels     map[string]func()
	cancelsLock sync.Mutex

	flowClient flow.DirektivFlowClient

	secretsClient secretsgrpc.SecretsServiceClient
	ingressClient ingress.DirektivIngressClient
	grpcConns     []*grpc.ClientConn

	metricsClient *metrics.Client
}

func newWorkflowEngine(s *WorkflowServer) (*workflowEngine, error) {

	var err error

	we := new(workflowEngine)
	we.server = s
	we.db = s.dbManager
	we.timer = s.tmManager
	we.instanceLogger = &s.instanceLogger
	we.cancels = make(map[string]func())

	we.stateLogics = map[model.StateType]func(*model.Workflow, model.State) (stateLogic, error){
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

	err = we.timer.registerFunction(sleepWakeupFunction, we.sleepWakeup)
	if err != nil {
		return nil, err
	}

	err = we.timer.registerFunction(retryWakeupFunction, we.retryWakeup)
	if err != nil {
		return nil, err
	}

	err = we.timer.registerFunction(timeoutFunction, we.timeoutHandler)
	if err != nil {
		return nil, err
	}

	err = we.timer.registerFunction(wfCron, we.wfCronHandler)
	if err != nil {
		return nil, err
	}

	// get flow client
	conn, err := GetEndpointTLS(s.config.FlowAPI.Endpoint, true)
	if err != nil {
		return nil, err
	}
	we.grpcConns = append(we.grpcConns, conn)

	we.flowClient = flow.NewDirektivFlowClient(conn)

	// get secrets client
	conn, err = GetEndpointTLS(secretsEndpoint, false)
	if err != nil {
		return nil, err
	}
	we.grpcConns = append(we.grpcConns, conn)
	we.secretsClient = secretsgrpc.NewSecretsServiceClient(conn)

	// get ingress client
	conn, err = GetEndpointTLS(s.config.IngressAPI.Endpoint, true)
	if err != nil {
		return nil, err
	}
	we.grpcConns = append(we.grpcConns, conn)
	we.ingressClient = ingress.NewDirektivIngressClient(conn)

	// setup metrics client
	we.metricsClient, err = metrics.NewClient()
	if err != nil {
		return nil, err
	}

	go we.checkTimeoutInstances()

	return we, nil

}

// checks if deadlines have exceeded. That can happen if a node dies
// this function picks them up and starts a retry.
// if that fails instance is getting cancelled
func (we *workflowEngine) checkTimeoutInstances() {

	ticker := time.NewTicker(5 * time.Minute)

	for {
		select {
		case <-ticker.C:
			log.Debugf("run expired worklflow thread")
			in, err := we.db.getWorkflowInstanceExpired(context.Background())
			if err != nil {
				log.Errorf("can not get expired workflows: %v", err)
				continue
			}
			for _, i := range in {
				data, _ := json.Marshal(&retryMessage{
					InstanceID: i.InstanceID,
					State:      i.Status,
					Step:       len(i.Flow),
				})

				log.Debugf("rescheduling workflow %v", i.InstanceID)
				err = we.retryWakeup(data)
				if err != nil {
					log.Errorf("can not kickstart workflow: %v", err)
					we.hardCancelInstance(i.InstanceID, "direktiv.cancels.kickstart", "cancelled by failed kickstart")
				}
			}
		}
	}

}

func (we *workflowEngine) localCancel(id string) {

	rec, err := we.db.getWorkflowInstance(we.db.ctx, id)
	if err == nil {
		err = we.timer.deleteTimerByName(rec.Controller, we.server.hostname, id)
		if err != nil {
			log.Error(err)
		}
	}

	we.cancelsLock.Lock()
	cancel, exists := we.cancels[id]
	if exists {
		delete(we.cancels, id)
		defer cancel()
	}
	we.cancelsLock.Unlock()

}

func (we *workflowEngine) finishCancelSubflow(id string) {
	we.localCancel(id)
}

type runStateMessage struct {
	InstanceID string
	State      string
	Step       int
}

func (we *workflowEngine) dispatchState(id, state string, step int) error {

	ctx := context.Background()

	// TODO: timeouts & retries

	var step32 int32
	step32 = int32(step)

	_, err := we.flowClient.Resume(ctx, &flow.ResumeRequest{
		InstanceId: &id,
		Step:       &step32,
	})
	if err != nil {
		return err
	}

	return nil

}

type eventsWaiterSignature struct {
	InstanceID string
	Step       int
}

type eventsResultMessage struct {
	InstanceID string
	State      string
	Step       int
	Payloads   []*cloudevents.Event
}

const eventsWakeupFunction = "eventsWakeup"

func (we *workflowEngine) wakeEventsWaiter(signature []byte, events []*cloudevents.Event) error {

	sig := new(eventsWaiterSignature)
	err := json.Unmarshal(signature, sig)
	if err != nil {
		return NewInternalError(err)
	}

	ctx, wli, err := we.loadWorkflowLogicInstance(sig.InstanceID, sig.Step)
	if err != nil {
		err = fmt.Errorf("cannot load workflow logic instance: %v", err)
		log.Error(err)
		return err
	}

	wakedata, err := json.Marshal(events)
	if err != nil {
		wli.Close()
		err = fmt.Errorf("cannot marshal the action results payload: %v", err)
		log.Error(err)
		return err
	}

	savedata, err := InstanceMemory(wli.rec)
	if err != nil {
		wli.Close()
		return err
	}

	go wli.engine.runState(ctx, wli, savedata, wakedata)

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

func (we *workflowEngine) doActionRequest(ctx context.Context, ar *isolateRequest) error {

	// TODO: should this ctx be modified with a shorter deadline?

	// generate hash name as "url"
	actionHash, err := serviceToHash(ar)

	if err != nil {
		return NewInternalError(err)
	}

	go we.doHTTPRequest(ctx, actionHash, ar)

	return nil

}

func (we *workflowEngine) doHTTPRequest(ctx context.Context,
	ah uint64, ar *isolateRequest) {

	// from here we need to report error as grpc because this is go-routined
	// prepare error here in case
	reportErr := func(err error) {
		ec := ""
		em := err.Error()
		step := int32(ar.Workflow.Step)
		r := &flow.ReportActionResultsRequest{
			InstanceId:   &ar.Workflow.InstanceID,
			Step:         &step,
			ActionId:     &ar.ActionID,
			ErrorCode:    &ec,
			ErrorMessage: &em,
		}

		_, err = we.flowClient.ReportActionResults(context.Background(), r)
		if err != nil {
			log.Errorf("can not respond to flow: %v", err)
		}
	}

	// NOTE: transport copied & modified from http.DefaultTransport
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

	// on https we add the cert to ca
	if we.server.config.FlowAPI.Protocol == "https" {

		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		// Read in the cert file. just in case it is the same being used
		// in the ingress
		certs, err := ioutil.ReadFile(TLSCert)
		if err == nil {
			// Append our cert to the system pool if we have it
			if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
				log.Println("No certs appended, using system certs only")
			}
		}

		// Trust the augmented cert pool in our client
		config := &tls.Config{
			InsecureSkipVerify: true,
			RootCAs:            rootCAs,
		}
		tr.TLSClientConfig = config

	}

	// configured namespace for workflows
	ns := os.Getenv(direktivWorkflowNamespace)

	// calculate address
	addr := fmt.Sprintf("%s://%s-%d.%s",
		we.server.config.FlowAPI.Protocol, ar.Workflow.Namespace, ah, ns)

	log.Debugf("isolate request: %v", addr)

	if ar.Workflow.Timeout == 0 {
		ar.Workflow.Timeout = 15 * 60 // 15 minutes default
	}

	deadline := time.Now().Add(time.Duration(ar.Workflow.Timeout) * time.Second)
	rctx, cancel := context.WithDeadline(context.Background(), deadline.Add(time.Second*60))
	defer cancel()

	req, err := http.NewRequestWithContext(rctx,
		http.MethodPost, addr,
		bytes.NewReader(ar.Container.Data))
	if err != nil {
		reportErr(err)
		return
	}

	// add headers
	req.Header.Add(DirektivNamespaceHeader, ar.Workflow.Namespace)
	req.Header.Add(DirektivActionIDHeader, ar.ActionID)
	req.Header.Add(DirektivInstanceIDHeader, ar.Workflow.InstanceID)
	req.Header.Add(DirektivDeadlineHeader, deadline.Format(time.RFC3339))
	req.Header.Add(DirektivStepHeader, fmt.Sprintf("%d",
		int64(ar.Workflow.Step)))

	client := &http.Client{
		Transport: tr,
	}

	var (
		resp *http.Response
	)

	// potentially dns error for a brand new service
	for i := 0; i < 400; i++ {
		log.Debugf("isolate request (%d): %v", i, addr)
		resp, err = client.Do(req)
		if err != nil {
			if err, ok := err.(*url.Error); ok {
				if err, ok := err.Err.(*net.OpError); ok {
					if _, ok := err.Err.(*net.DNSError); ok {
						// this happens because the function does not exist
						kubeReq.mtx.Lock()
						err := getKnativeFunction(fmt.Sprintf("%s-%d", ar.Workflow.Namespace, ah))

						if err != nil {
							err := addKnativeFunction(ar)
							if err != nil {
								reportErr(fmt.Errorf("can not create knative function %v: %v", addr, err))
								return
							}
						}
						kubeReq.mtx.Unlock()

						time.Sleep(250 * time.Millisecond)
						continue
					}
				}
			}

		} else {
			break
		}
	}

	if err != nil {
		reportErr(err)
		return
	}

	if resp.StatusCode != 200 {
		reportErr(fmt.Errorf("action error status: %d",
			resp.StatusCode))
	}

	log.Debugf("isolate request done")

}

const actionWakeupFunction = "actionWakeup"

func (we *workflowEngine) wakeCaller(ctx context.Context, msg *actionResultMessage) error {

	// TODO: timeouts & retries

	var step int32
	step = int32(msg.Step)

	_, err := we.flowClient.ReportActionResults(ctx, &flow.ReportActionResultsRequest{
		InstanceId:   &msg.InstanceID,
		Step:         &step,
		ActionId:     &msg.Payload.ActionID,
		ErrorCode:    &msg.Payload.ErrorCode,
		ErrorMessage: &msg.Payload.ErrorMessage,
		Output:       msg.Payload.Output,
	})
	if err != nil {
		return err
	}

	return nil

}

const wfCron = "wfcron"

func (we *workflowEngine) wfCronHandler(data []byte) error {

	return we.CronInvoke(string(data))

}

type sleepMessage struct {
	InstanceID string
	State      string
	Step       int
}

const sleepWakeupFunction = "sleepWakeup"
const sleepWakedata = "sleep"

func (we *workflowEngine) sleep(id, state string, step int, t time.Time) error {

	data, _ := json.Marshal(&sleepMessage{
		InstanceID: id,
		State:      state,
		Step:       step,
	})

	err := we.timer.addOneShot(id, sleepWakeupFunction, t, data)
	if err != nil {
		return NewInternalError(err)
	}

	return nil

}

func (we *workflowEngine) sleepWakeup(data []byte) error {

	msg := new(sleepMessage)

	err := json.Unmarshal(data, msg)
	if err != nil {
		log.Errorf("cannot handle sleep wakeup: %v", err)
		return nil
	}

	ctx, wli, err := we.loadWorkflowLogicInstance(msg.InstanceID, msg.Step)
	if err != nil {
		log.Errorf("cannot load workflow logic instance: %v", err)
		return nil
	}

	wli.Log("Waking up from sleep.")

	go wli.engine.runState(ctx, wli, nil, []byte(sleepWakedata))

	return nil

}

func (we *workflowEngine) cancelRecordsChildren(ctx context.Context, rec *ent.WorkflowInstance) error {

	wfrec, err := rec.QueryWorkflow().Only(ctx)
	if err != nil {
		return err
	}

	wf := new(model.Workflow)
	err = wf.Load(wfrec.Workflow)
	if err != nil {
		return err
	}

	step := len(rec.Flow)
	state := rec.Flow[step-1]
	states := wf.GetStatesMap()
	stateObject, exists := states[state]
	if !exists {
		return NewInternalError(fmt.Errorf("workflow cannot resolve state: %s", state))
	}
	init, exists := we.stateLogics[stateObject.GetType()]
	if !exists {
		return NewInternalError(fmt.Errorf("engine cannot resolve state type: %s", stateObject.GetType().String()))
	}
	stateLogic, err := init(wf, stateObject)
	if err != nil {
		return NewInternalError(fmt.Errorf("cannot initialize state logic: %v", err))
	}
	logic := stateLogic

	savedata, err := InstanceMemory(rec)
	if err != nil {
		return err
	}

	we.cancelChildren(logic, savedata)

	return nil

}

func (we *workflowEngine) cancelChildren(logic stateLogic, savedata []byte) {

	if len(savedata) == 0 {
		return
	}

	children := logic.LivingChildren(savedata)
	for _, child := range children {
		switch child.Type {
		case "isolate":
			syncServer(context.Background(), we.db, &we.server.id, child.Id, CancelIsolate)
		case "subflow":
			go func(id string) {
				we.hardCancelInstance(id, "direktiv.cancels.parent", "cancelled by parent workflow")
			}(child.Id)
		default:
			log.Errorf("unrecognized child type: %s", child.Type)
		}
	}

}

func (we *workflowEngine) hardCancelInstance(instanceId, code, message string) error {
	return we.cancelInstance(instanceId, code, message, false)
}

func (we *workflowEngine) softCancelInstance(instanceId string, step int, code, message string) error {
	// TODO: step
	return we.cancelInstance(instanceId, code, message, true)
}

func (we *workflowEngine) clearEventListeners(rec *ent.WorkflowInstance) {
	_ = we.db.deleteWorkflowEventListenerByInstanceID(rec.ID)
}

func (we *workflowEngine) freeResources(rec *ent.WorkflowInstance) {

	err := we.timer.deleteTimersForInstance(rec.InstanceID)
	if err != nil {
		log.Error(err)
	}
	log.Debugf("deleted timers for instance %v", rec.InstanceID)

	we.clearEventListeners(rec)

	var namespace, workflow, instance string
	namespace = rec.Edges.Workflow.Edges.Namespace.ID
	workflow = rec.Edges.Workflow.ID.String()
	instance = rec.InstanceID
	we.server.variableStorage.DeleteAllInScope(context.Background(), namespace, workflow, instance)

}

func (we *workflowEngine) cancelInstance(instanceId, code, message string, soft bool) error {

	killer := make(chan bool)

	ctx := context.Background()

	go func() {

		timer := time.After(time.Millisecond)

		for {

			select {
			case <-timer:
				// broadcast cancel across cluster
				syncServer(ctx, we.db, &we.server.id, instanceId, CancelSubflow)
				// TODO: mark cancelled instances even if not scheduled in
			case <-killer:
				return
			}

		}

	}()

	defer func() {
		close(killer)
	}()

	tx, err := we.db.dbEnt.Tx(ctx)
	if err != nil {
		return err
	}

	rec, err := tx.WorkflowInstance.Query().Where(workflowinstance.InstanceIDEQ(instanceId)).Only(ctx)
	if err != nil {
		return rollback(tx, err)
	}

	if rec.Status != "pending" && rec.Status != "running" {
		return rollback(tx, nil)
	}

	we.completeState(ctx, rec, "", code, false)

	ns, err := rec.QueryWorkflow().QueryNamespace().Only(ctx)
	if err != nil {
		return rollback(tx, err)
	}

	rec, err = rec.Update().
		SetStatus("cancelled").
		SetEndTime(time.Now()).
		SetErrorCode(code).
		SetErrorMessage(message).
		Save(ctx)
	if err != nil {
		return rollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		return rollback(tx, err)
	}

	rec, err = we.db.getWorkflowInstanceByID(ctx, rec.ID)
	if err != nil {
		return err
	}

	err = we.cancelRecordsChildren(ctx, rec)
	if err != nil {
		log.Error(err)
	}

	err = we.timer.deleteTimerByName(rec.Controller, we.server.hostname, instanceId)
	if err != nil {
		log.Error(err)
	}
	// TODO: cancel any other outstanding timers

	logger, err := (*we.instanceLogger).LoggerFunc(ns.ID, instanceId)
	if err != nil {
		dl, _ := dummy.NewLogger()
		logger, _ = dl.LoggerFunc(ns.ID, instanceId)
	}
	defer logger.Close()

	logger.Info(fmt.Sprintf("Workflow %s.", message))

	we.freeResources(rec)

	if rec.InvokedBy != "" {

		// wakeup caller
		caller := new(subflowCaller)
		err = json.Unmarshal([]byte(rec.InvokedBy), caller)
		if err != nil {
			log.Error(err)
			return nil
		}

		msg := &actionResultMessage{
			InstanceID: caller.InstanceID,
			State:      caller.State,
			Step:       caller.Step,
			Payload: actionResultPayload{
				ActionID:     instanceId,
				ErrorCode:    rec.ErrorCode,
				ErrorMessage: rec.ErrorMessage,
			},
		}

		logger.Info(fmt.Sprintf("Reporting failure to calling workflow."))

		err = we.wakeCaller(ctx, msg)
		if err != nil {
			log.Error(err)
			return nil
		}

	}

	return nil

}

type retryMessage struct {
	InstanceID string
	State      string
	Step       int
}

const retryWakeupFunction = "retryWakeup"

func (we *workflowEngine) scheduleRetry(id, state string, step int, t time.Time) error {

	data, _ := json.Marshal(&retryMessage{
		InstanceID: id,
		State:      state,
		Step:       step,
	})

	err := we.timer.addOneShot(id, retryWakeupFunction, t, data)
	if err != nil {
		return NewInternalError(err)
	}

	return nil

}

func (we *workflowEngine) retryWakeup(data []byte) error {

	msg := new(retryMessage)

	err := json.Unmarshal(data, msg)
	if err != nil {
		log.Errorf("cannot handle retry wakeup: %v", err)
		return nil
	}

	ctx, wli, err := we.loadWorkflowLogicInstance(msg.InstanceID, msg.Step)
	if err != nil {
		log.Errorf("cannot load workflow logic instance: %v", err)
		return nil
	}

	wli.Log("Retrying failed state.")

	go wli.engine.runState(ctx, wli, nil, nil)

	return nil

}

const maxWorkflowSteps = 10

func (we *workflowEngine) transformState(wli *workflowLogicInstance, transition *stateTransition) error {

	if transition == nil || transition.Transform == "" || transition.Transform == "." {
		return nil
	}

	wli.Log("Transforming state data.")

	err := wli.Transform(transition.Transform)
	if err != nil {
		return err
	}

	return nil

}

func (we *workflowEngine) completeState(ctx context.Context, rec *ent.WorkflowInstance, nextState, errCode string, retrying bool) {

	if len(rec.Flow) == 0 {
		return
	}

	if rec.Status != "pending" {
		return
	}

	args := new(metrics.InsertRecordArgs)

	wf, err := rec.QueryWorkflow().Only(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	ns, err := wf.QueryNamespace().Only(ctx)
	if err != nil {
		log.Error(err)
		return
	}

	args.Namespace = ns.ID
	args.Workflow = wf.Name
	args.Instance = rec.InstanceID
	args.Invoker = rec.InvokedBy

	args.State = rec.Flow[len(rec.Flow)-1]

	d := time.Now().Sub(rec.StateBeginTime)
	args.WorkflowMilliSeconds = d.Milliseconds()

	args.ErrorCode = errCode
	args.Transition = nextState
	args.Next = metrics.NextTransition
	if nextState == "" {
		args.Next = metrics.NextEnd
	} else if retrying {
		args.Next = metrics.NextRetry
	}

	if len(rec.Flow) == 1 {
		args.Invoker = "start"
	}

	err = we.metricsClient.InsertRecord(args)
	if err != nil {
		log.Error(err)
	}

}

func (we *workflowEngine) transitionState(ctx context.Context, wli *workflowLogicInstance, transition *stateTransition, errCode string) {

	if transition == nil {
		return
	}

	we.completeState(ctx, wli.rec, transition.NextState, errCode, false)

	if transition.NextState != "" {
		wli.Log("Transitioning to next state: %s (%d).", transition.NextState, wli.step)
		go wli.Transition(transition.NextState, 0)
		return
	}

	var rec *ent.WorkflowInstance
	data, err := json.Marshal(wli.data)
	if err != nil {
		err = fmt.Errorf("engine cannot marshal state data for storage: %v", err)
		log.Error(err)
		return
	}

	status := "complete"
	if wli.rec.ErrorCode != "" {
		status = "failed"
		wli.Log("Workflow failed with error '%s': %s", wli.rec.ErrorCode, wli.rec.ErrorMessage)
	}

	wf := wli.rec.Edges.Workflow
	rec, err = wli.rec.Update().SetOutput(string(data)).SetEndTime(time.Now()).SetStatus(status).Save(ctx)
	if err != nil {
		log.Error(err)
		return
	}
	rec.Edges.Workflow = wf

	wli.rec = rec
	log.Debugf("Workflow instance completed: %s", wli.id)
	wli.Log("Workflow completed.")

	wli.engine.freeResources(rec)

	wli.wakeCaller(ctx, data)

}

func (we *workflowEngine) logRunState(wli *workflowLogicInstance, savedata, wakedata []byte) {

	log.Debugf("Running state logic -- %s:%v (%s)", wli.id, wli.step, wli.logic.ID())
	if len(savedata) == 0 && len(wakedata) == 0 {
		wli.Log("Running state logic -- %s:%v (%s)", wli.logic.ID(), wli.step, wli.logic.Type())
	}

}

func (we *workflowEngine) runState(ctx context.Context, wli *workflowLogicInstance, savedata, wakedata []byte) {

	we.logRunState(wli, savedata, wakedata)

	defer wli.unlock()
	defer wli.Close()

	var err error
	var code string
	var transition *stateTransition

	if lq := wli.logic.LogJQ(); len(savedata) == 0 && len(wakedata) == 0 && lq != "" {
		var object interface{}
		object, err = jqObject(wli.data, ".")
		if err != nil {
			goto failure
		}

		var data []byte
		data, err = json.MarshalIndent(object, "", "  ")
		if err != nil {
			err = NewInternalError(fmt.Errorf("failed to marshal state data: %w", err))
			goto failure
		}

		wli.UserLog(ctx, string(data))
	}

	transition, err = wli.logic.Run(ctx, wli, savedata, wakedata)
	if err != nil {
		goto failure
	}

	err = we.transformState(wli, transition)
	if err != nil {
		goto failure
	}

next:
	we.transitionState(ctx, wli, transition, code)
	return

failure:

	var breaker int

	if breaker > 10 {
		err = NewInternalError(errors.New("somehow ended up in a catchable error loop"))
	}

	savedata, err2 := InstanceMemory(wli.rec)
	if err2 == nil {
		wli.engine.cancelChildren(wli.logic, savedata)
	}

	if uerr, ok := err.(*UncatchableError); ok {

		err = wli.setStatus(ctx, "failed", uerr.Code, uerr.Message)
		if err != nil {
			err = NewInternalError(err)
			goto failure
		}

		wli.Log("Workflow failed with uncatchable error: %s", uerr.Message)

		wli.engine.freeResources(wli.rec)
		wli.wakeCaller(ctx, nil)
		return

	} else if cerr, ok := err.(*CatchableError); ok {

		var hasAlreadyFailed bool
		_ = wli.StoreData("error", cerr)

	alreadyFailed:

		for i, catch := range wli.logic.ErrorCatchers() {

			var matched bool

			// NOTE: this error should be checked in model validation
			matched, _ = regexp.MatchString(catch.Error, cerr.Code)

			if matched {

				wli.Log("State failed with error '%s': %s", cerr.Code, cerr.Message)
				wli.Log("Error caught by error definition %d: %s", i, catch.Error)

				transition = &stateTransition{
					Transform: "",
					NextState: catch.Transition,
				}

				breaker++

				code = cerr.Code

				goto next

			}

		}

		if retries := wli.logic.Retries(); retries != nil && !hasAlreadyFailed {
			if wli.rec.Attempts < retries.MaxAttempts {
				err = wli.Retry(ctx, retries.Delay, retries.Multiplier, cerr.Code)
				if err != nil {
					goto failure
				}
				return
			} else {
				wli.Log("Maximum retry attempts exceeded.")
				if retries.Throw != "" {
					cerr = NewCatchableError(retries.Throw, retries.Throw)
					hasAlreadyFailed = true
					goto alreadyFailed
				}
			}
		}

		err = wli.setStatus(ctx, "failed", cerr.Code, cerr.Message)
		if err != nil {
			err = NewInternalError(err)
			goto failure
		}

		wli.Log("Workflow failed with uncaught error '%s': %s", cerr.Code, cerr.Message)
		wli.engine.freeResources(wli.rec)
		wli.wakeCaller(ctx, nil)
		return

	} else if ierr, ok := err.(*InternalError); ok {

		code := ""
		msg := "an internal error occurred"

		if wli != nil && wli.rec != nil {

			var err error
			err = wli.setStatus(ctx, "crashed", code, msg)
			if err == nil {
				log.Errorf("Workflow failed with internal error: %s", ierr.Error())
				wli.Log("Workflow crashed due to an internal error.")
				wli.wakeCaller(ctx, nil)
				return
			}

		}

		log.Errorf("Workflow failed with internal error and the database couldn't be updated: %s", ierr.Error())

		wli.engine.freeResources(wli.rec)

	} else {
		log.Errorf("Unwrapped error detected: %v", err)
	}

	return

}

func (we *workflowEngine) CronInvoke(uid string) error {

	var err error

	ctx := context.Background()

	wf, err := we.db.getWorkflow(uid)
	if err != nil {
		return err
	}

	ns, err := wf.QueryNamespace().Only(ctx)
	if err != nil {
		return nil
	}

	wli, err := we.newWorkflowLogicInstance(ctx, ns.ID, wf.Name, []byte("{}"))
	if err != nil {
		if _, ok := err.(*InternalError); ok {
			log.Errorf("Internal error on CronInvoke: %v", err)
			return errors.New("an internal error occurred")
		} else {
			return err
		}
	}
	defer wli.Close()

	if wli.wf.Start == nil || wli.wf.Start.GetType() != model.StartTypeScheduled {
		return fmt.Errorf("cannot cron invoke workflows with '%s' starts", wli.wf.Start.GetType())
	}

	wli.rec, err = we.db.addWorkflowInstance(ctx, ns.ID, wf.Name, wli.id, string(wli.startData), true)
	if err != nil {
		if strings.Contains(err.Error(), "invoked") || strings.Contains(err.Error(), "transactions") {
			return nil
		}
		return NewInternalError(err)
	}

	start := wli.wf.GetStartState()

	wli.Log("Beginning workflow triggered by API.")

	go wli.Transition(start.GetID(), 0)

	return nil

}

func (we *workflowEngine) DirectInvoke(ctx context.Context, namespace, name string, input []byte) (string, error) {

	var err error

	wli, err := we.newWorkflowLogicInstance(ctx, namespace, name, input)
	if err != nil {
		if _, ok := err.(*InternalError); ok {
			log.Errorf("Internal error on DirectInvoke: %v", err)
			return "", errors.New("an internal error occurred")
		}

		return "", err
	}
	defer wli.Close()

	if wli.wf.Start != nil && wli.wf.Start.GetType() != model.StartTypeDefault {
		return "", fmt.Errorf("cannot directly invoke workflows with '%s' starts", wli.wf.Start.GetType())
	}

	wli.rec, err = we.db.addWorkflowInstance(ctx, namespace, name, wli.id, string(wli.startData), false)
	if err != nil {
		return "", NewInternalError(err)
	}

	start := wli.wf.GetStartState()

	wli.Log("Beginning workflow triggered by API.")

	go wli.Transition(start.GetID(), 0)

	return wli.id, nil

}

func (we *workflowEngine) EventsInvoke(workflowID uuid.UUID, events ...*cloudevents.Event) {

	ctx := context.Background()

	wf, err := we.db.getWorkflowByID(workflowID)
	if err != nil {
		log.Errorf("Internal error on EventsInvoke: %v", err)
		return
	}

	ns, err := wf.QueryNamespace().Only(ctx)
	if err != nil {
		log.Errorf("Internal error on EventsInvoke: %v", err)
		return
	}

	var input []byte
	m := make(map[string]interface{})
	for _, event := range events {

		if event == nil {
			continue
		}

		var x interface{}

		x, err = extractEventPayload(event)
		if err != nil {
			return
		}

		m[event.Type()] = x

	}

	input, err = json.Marshal(m)
	if err != nil {
		log.Errorf("Internal error on EventsInvoke: %v", err)
		return
	}

	namespace := ns.ID
	name := wf.Name

	wli, err := we.newWorkflowLogicInstance(ctx, namespace, name, input)
	if err != nil {
		log.Errorf("Internal error on EventsInvoke: %v", err)
		return
	}
	defer wli.Close()

	var stype model.StartType
	if wli.wf.Start != nil {
		stype = wli.wf.Start.GetType()
	}

	switch stype {
	case model.StartTypeEvent:
	case model.StartTypeEventsAnd:
	case model.StartTypeEventsXor:
	default:
		log.Errorf("cannot event invoke workflows with '%s' starts", stype)
		return
	}

	wli.rec, err = we.db.addWorkflowInstance(ctx, namespace, name, wli.id, string(wli.startData), false)
	if err != nil {
		log.Errorf("Internal error on EventsInvoke: %v", err)
		return
	}

	start := wli.wf.GetStartState()

	if len(events) == 1 {
		wli.Log("Beginning workflow triggered by event: %s", events[0].ID())
	} else {
		var ids = make([]string, len(events))
		for i := range events {
			ids[i] = events[i].ID()
		}
		wli.Log("Beginning workflow triggered by events: %v", ids)
	}

	go wli.Transition(start.GetID(), 0)

}

type subflowCaller struct {
	InstanceID string
	State      string
	Step       int
	Depth      int
}

const maxSubflowDepth = 5

func (we *workflowEngine) subflowInvoke(ctx context.Context, caller *subflowCaller, callersCaller, namespace, name string, input []byte) (string, error) {

	var err error

	if callersCaller != "" {
		cc := new(subflowCaller)
		err = json.Unmarshal([]byte(callersCaller), cc)
		if err != nil {
			log.Errorf("Internal error on subflowInvoke: %v", err)
			return "", errors.New("an internal error occurred")
		}

		caller.Depth = cc.Depth + 1
		if caller.Depth > maxSubflowDepth {
			err = NewUncatchableError("direktiv.limits.depth", "instance aborted for exceeding the maximum subflow depth (%d)", maxSubflowDepth)
			return "", err
		}
	}

	wli, err := we.newWorkflowLogicInstance(ctx, namespace, name, input)
	if err != nil {
		if _, ok := err.(*InternalError); ok {
			log.Errorf("Internal error on subflowInvoke: %v", err)
			return "", errors.New("an internal error occurred")
		} else {
			return "", err
		}
	}
	defer wli.Close()

	if wli.wf.Start != nil && wli.wf.Start.GetType() != model.StartTypeDefault {
		return "", fmt.Errorf("cannot subflow invoke workflows with '%s' starts", wli.wf.Start.GetType())
	}

	wli.rec, err = we.db.addWorkflowInstance(ctx, namespace, name, wli.id, string(wli.startData), false)
	if err != nil {
		return "", NewInternalError(err)
	}

	if caller != nil {

		var data []byte

		data, err = json.Marshal(caller)
		if err != nil {
			return "", NewInternalError(err)
		}

		wli.rec, err = wli.rec.Update().SetInvokedBy(string(data)).Save(ctx)
		if err != nil {
			return "", NewInternalError(err)
		}

	}

	start := wli.wf.GetStartState()

	wli.Log("Beginning workflow triggered as subflow to caller: %s", caller.InstanceID)

	go wli.Transition(start.GetID(), 0)

	return wli.id, nil

}

const timeoutFunction = "timeoutFunction"

type timeoutArgs struct {
	InstanceId string
	Step       int
	Soft       bool
}

func (we *workflowEngine) timeoutHandler(input []byte) error {

	args := new(timeoutArgs)
	err := json.Unmarshal(input, args)
	if err != nil {
		return err
	}

	if args.Soft {
		we.softCancelInstance(args.InstanceId, args.Step, "direktiv.cancels.timeout", "operation timed out")
	} else {
		we.hardCancelInstance(args.InstanceId, "direktiv.cancels.timeout", "workflow timed out")
	}

	return nil

}

func (we *workflowEngine) listenForEvents(ctx context.Context, wli *workflowLogicInstance, events []*model.ConsumeEventDefinition, all bool) error {

	wfid, err := wli.rec.QueryWorkflow().OnlyID(ctx)
	if err != nil {
		return err
	}

	signature, err := json.Marshal(&eventsWaiterSignature{
		InstanceID: wli.id,
		Step:       wli.step,
	})
	if err != nil {
		return err
	}

	var transformedEvents []*model.ConsumeEventDefinition

	for i := range events {

		ev := new(model.ConsumeEventDefinition)
		copier.Copy(ev, events[i])

		for k, v := range events[i].Context {

			str, ok := v.(string)
			if !ok {
				continue
			}

			if strings.HasPrefix(str, "{{") && strings.HasSuffix(str, "}}") {

				query := str[2 : len(str)-2]
				x, err := jqOne(wli.data, query)
				if err != nil {
					return fmt.Errorf("failed to execute jq query for key '%s' on event definition %d: %v", k, i, err)
				}

				switch x.(type) {
				case bool:
				case int:
				case string:
				case []byte:
				case time.Time:
				default:
					return fmt.Errorf("jq query on key '%s' for event definition %d returned an unacceptable type: %v", k, i, reflect.TypeOf(x))
				}

				ev.Context[k] = x

			}

		}

		transformedEvents = append(transformedEvents, ev)

	}

	_, err = we.db.addWorkflowEventListener(wfid, wli.rec.ID,
		transformedEvents, signature, all)
	if err != nil {
		return err
	}

	wli.Log("Registered to receive events.")

	return nil

}
