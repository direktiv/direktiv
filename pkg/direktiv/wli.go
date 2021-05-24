package direktiv

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/itchyny/gojq"
	hashstructure "github.com/mitchellh/hashstructure/v2"
	"github.com/senseyeio/duration"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/pkg/dlog"
	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/direktiv/pkg/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type workflowLogicInstance struct {
	engine    *workflowEngine
	data      interface{}
	startData []byte
	wf        *model.Workflow
	rec       *ent.WorkflowInstance
	step      int

	namespace   string
	id          string
	logToEvents string
	lockConn    *sql.Conn
	logic       stateLogic
	logger      dlog.Logger
}

func (we *workflowEngine) newWorkflowLogicInstance(ctx context.Context, namespace, name string, input []byte) (*workflowLogicInstance, error) {

	var err error
	var inputData, stateData interface{}

	err = json.Unmarshal(input, &inputData)
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

	rec, err := we.db.getNamespaceWorkflow(ctx, name, namespace)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, NewUncatchableError("direktiv.subflow.notExist", "workflow '%s' does not exist", name)
		}
		return nil, NewInternalError(err)
	}

	if !rec.Active {
		return nil, grpc.Errorf(codes.InvalidArgument, "workflow is inactive")
	}

	wf := new(model.Workflow)
	err = wf.Load(rec.Workflow)
	if err != nil {
		return nil, NewInternalError(err)
	}

	wli := new(workflowLogicInstance)
	wli.namespace = namespace
	wli.engine = we
	wli.wf = wf
	wli.data = stateData
	wli.logToEvents = rec.LogToEvents

	wli.id = fmt.Sprintf("%s/%s/%s", namespace, name, randSeq(6))
	wli.startData, err = json.MarshalIndent(wli.data, "", "  ")
	if err != nil {
		return nil, NewInternalError(err)
	}

	wli.logger, err = (*we.instanceLogger).LoggerFunc(namespace, wli.id)
	if err != nil {
		return nil, NewInternalError(err)
	}

	return wli, nil

}

func (wli *workflowLogicInstance) start() {

	ctx, err := wli.lock(time.Second * 5)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debugf("Starting workflow %v", wli.id)
	start := wli.wf.GetStartState()
	wli.Transition(ctx, start.GetID(), 0)

}

func (we *workflowEngine) loadWorkflowLogicInstance(id string, step int) (context.Context, *workflowLogicInstance, error) {

	wli := new(workflowLogicInstance)
	wli.id = id
	wli.engine = we

	var success bool

	defer func() {
		if !success {
			wli.unlock()
		}
	}()

	ctx, err := wli.lock(time.Second * defaultLockWait)
	if err != nil {
		return ctx, nil, NewInternalError(fmt.Errorf("cannot assume control of workflow instance lock: %v", err))
	}

	rec, err := we.db.getWorkflowInstance(ctx, id)
	if err != nil {
		wli.unlock()
		return nil, nil, NewInternalError(err)
	}
	wli.rec = rec

	qwf := wli.rec.Edges.Workflow
	qns := qwf.Edges.Namespace

	wli.namespace = qns.ID

	wli.logger, err = (*we.instanceLogger).LoggerFunc(qns.ID, wli.id)
	if err != nil {
		wli.unlock()
		return ctx, nil, NewInternalError(fmt.Errorf("cannot initialize instance logger: %v", err))
	}

	err = json.Unmarshal([]byte(rec.StateData), &wli.data)
	if err != nil {
		wli.unlock()
		return ctx, nil, NewInternalError(fmt.Errorf("cannot load saved workflow state data: %v", err))
	}

	wli.wf = new(model.Workflow)
	wli.logToEvents = qwf.LogToEvents

	err = wli.wf.Load(qwf.Workflow)
	if err != nil {
		wli.unlock()
		return ctx, nil, NewInternalError(fmt.Errorf("cannot load saved workflow definition: %v", err))
	}

	if rec.Status != "pending" && rec.Status != "running" {
		wli.unlock()
		return ctx, nil, NewInternalError(fmt.Errorf("aborting workflow logic: database records instance terminated"))
	}

	wli.step = step
	if len(rec.Flow) != wli.step {
		wli.unlock()
		return ctx, nil, NewInternalError(fmt.Errorf("aborting workflow logic: steps out of sync (expect/actual - %d/%d)", step, len(rec.Flow)))
	}

	state := rec.Flow[step-1]
	states := wli.wf.GetStatesMap()
	stateObject, exists := states[state]
	if !exists {
		wli.unlock()
		return ctx, nil, NewInternalError(fmt.Errorf("workflow cannot resolve state: %s", state))
	}

	init, exists := wli.engine.stateLogics[stateObject.GetType()]
	if !exists {
		wli.unlock()
		return ctx, nil, NewInternalError(fmt.Errorf("engine cannot resolve state type: %s", stateObject.GetType().String()))
	}

	stateLogic, err := init(wli.wf, stateObject)
	if err != nil {
		wli.unlock()
		return ctx, nil, NewInternalError(fmt.Errorf("cannot initialize state logic: %v", err))
	}
	wli.logic = stateLogic

	success = true

	return ctx, wli, nil

}

func (wli *workflowLogicInstance) Close() error {

	if wli.lockConn != nil {
		wli.unlock()
	}

	if wli.logger != nil {
		err := wli.logger.Close()
		if err != nil {
			return err
		}
	}

	return nil

}

func (wli *workflowLogicInstance) Raise(ctx context.Context, cerr *CatchableError) error {

	var err error

	if wli.rec.ErrorCode == "" {
		wf := wli.rec.Edges.Workflow
		wli.rec, err = wli.rec.Update().
			SetStatus("failed").
			SetErrorCode(cerr.Code).
			SetErrorMessage(cerr.Message).
			Save(ctx)
		wli.rec.Edges.Workflow = wf
		if err != nil {
			return NewInternalError(err)
		}
	} else {
		return NewCatchableError(ErrCodeMultipleErrors, "the workflow instance tried to throw multiple errors")
	}

	return nil

}

func (wli *workflowLogicInstance) setStatus(ctx context.Context, status, code, message string) error {

	var err error

	if status == "crashed" {
		code = "direktiv.internal.error"
	}

	wli.engine.completeState(ctx, wli.rec, "", code, false)

	wf := wli.rec.Edges.Workflow

	if wli.rec.ErrorCode == "" {
		wli.rec, err = wli.rec.Update().
			SetStatus(status).
			SetEndTime(time.Now()).
			SetErrorCode(code).
			SetErrorMessage(message).
			Save(ctx)
		wli.rec.Edges.Workflow = wf
		return err
	}

	wli.rec, err = wli.rec.Update().
		SetEndTime(time.Now()).
		Save(ctx)
	wli.rec.Edges.Workflow = wf

	return err

}

func (wli *workflowLogicInstance) wakeCaller(ctx context.Context, data []byte) {

	// wake API call if there is a waiter
	go publishToAPI(wli.engine.server.dbManager, wli.id)

	if wli.rec.InvokedBy != "" {

		// wakeup caller
		caller := new(subflowCaller)
		err := json.Unmarshal([]byte(wli.rec.InvokedBy), caller)
		if err != nil {
			log.Error(err)
			return
		}

		msg := &actionResultMessage{
			InstanceID: caller.InstanceID,
			State:      caller.State,
			Step:       caller.Step,
			Payload: actionResultPayload{
				ActionID:     wli.id,
				ErrorCode:    wli.rec.ErrorCode,
				ErrorMessage: wli.rec.ErrorMessage,
				Output:       data,
			},
		}

		wli.Log("Reporting results to calling workflow.")

		err = wli.engine.wakeCaller(ctx, msg)
		if err != nil {
			log.Error(err)
			return
		}

	}

}

func (wli *workflowLogicInstance) lock(timeout time.Duration) (context.Context, error) {

	hash, err := hashstructure.Hash(wli.id, hashstructure.FormatV2, nil)
	if err != nil {
		return nil, NewInternalError(err)
	}

	wait := int(timeout.Seconds())
	conn, err := wli.engine.db.lockDB(hash, wait)
	if err != nil {
		return nil, NewInternalError(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wli.engine.cancelsLock.Lock()
	wli.lockConn = conn
	wli.engine.cancels[wli.id] = cancel
	wli.engine.cancelsLock.Unlock()

	return ctx, nil

}

func (wli *workflowLogicInstance) unlock() {

	if wli.lockConn == nil {
		return
	}

	hash, err := hashstructure.Hash(wli.id, hashstructure.FormatV2, nil)
	if err != nil {
		log.Error(NewInternalError(err))
		return
	}

	wli.engine.cancelsLock.Lock()
	cancel := wli.engine.cancels[wli.id]
	delete(wli.engine.cancels, wli.id)
	cancel()

	err = wli.engine.db.unlockDB(hash, wli.lockConn)
	wli.lockConn = nil
	wli.engine.cancelsLock.Unlock()

	if err != nil {
		log.Error(NewInternalError(fmt.Errorf("Failed to unlock database mutex: %v", err)))
		return
	}

	return

}

func jq(input interface{}, command string) ([]interface{}, error) {

	data, err := json.Marshal(input)
	if err != nil {
		return nil, NewInternalError(err)
	}

	var x interface{}

	err = json.Unmarshal(data, &x)
	if err != nil {
		return nil, NewInternalError(err)
	}

	query, err := gojq.Parse(command)
	if err != nil {
		return nil, NewCatchableError(ErrCodeJQBadQuery, err.Error())
	}

	var output []interface{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	iter := query.RunWithContext(ctx, x)

	for i := 0; ; i++ {

		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			return nil, NewUncatchableError("direktiv.jq.badCommand", err.Error())
		}

		output = append(output, v)

	}

	return output, nil

}

func jqOne(input interface{}, command string) (interface{}, error) {

	output, err := jq(input, command)
	if err != nil {
		return nil, err
	}

	if len(output) != 1 {
		return nil, NewCatchableError(ErrCodeJQNotObject, "the `jq` command produced multiple outputs")
	}

	return output[0], nil

}

func jqObject(input interface{}, command string) (map[string]interface{}, error) {

	x, err := jqOne(input, command)
	if err != nil {
		return nil, err
	}

	m, ok := x.(map[string]interface{})
	if !ok {
		return nil, NewCatchableError(ErrCodeJQNotObject, "the `jq` command produced a non-object output")
	}

	return m, nil

}

func (wli *workflowLogicInstance) UserLog(ctx context.Context, msg string, a ...interface{}) {

	s := fmt.Sprintf(msg, a...)

	wli.logger.Info(s)

	// TODO: detect content type and handle base64 data

	if attr := wli.logToEvents; attr != "" {
		event := cloudevents.NewEvent()
		event.SetID(uuid.New().String())
		event.SetSource(wli.wf.ID)
		event.SetType("direktiv.instanceLog")
		event.SetExtension("logger", attr)
		event.SetData("application/json", s)
		data, err := event.MarshalJSON()
		if err != nil {
			log.Errorf("failed to marshal UserLog cloudevent: %v", err)
			return
		}
		_, err = wli.engine.ingressClient.BroadcastEvent(ctx, &ingress.BroadcastEventRequest{
			Namespace:  &wli.namespace,
			Cloudevent: data,
		})
		if err != nil {
			log.Errorf("failed to broadcast cloudevent: %v", err)
			return
		}
	}

}

func (wli *workflowLogicInstance) Log(msg string, a ...interface{}) {

	s := fmt.Sprintf(msg, a...)

	wli.logger.Info(s)

}

func (wli *workflowLogicInstance) Save(ctx context.Context, data []byte) error {
	var err error

	str := base64.StdEncoding.EncodeToString(data)

	wli.rec, err = wli.rec.Update().SetMemory(str).Save(ctx)
	if err != nil {
		return NewInternalError(err)
	}
	return nil
}

func (wli *workflowLogicInstance) StoreData(key string, val interface{}) error {

	m, ok := wli.data.(map[string]interface{})
	if !ok {
		return NewInternalError(errors.New("unable to store data because state data isn't a valid JSON object"))
	}

	m[key] = val

	return nil

}

func (wli *workflowLogicInstance) Transform(transform string) error {

	x, err := jqObject(wli.data, transform)
	if err != nil {
		return WrapCatchableError("unable to apply transform: %v", err)
	}

	wli.data = x
	return nil

}

func (wli *workflowLogicInstance) Retry(ctx context.Context, delayString string, multiplier float64, errCode string) error {

	var err error
	var x interface{}

	err = json.Unmarshal([]byte(wli.rec.StateData), &x)
	if err != nil {
		return NewInternalError(err)
	}

	wli.data = x

	nextState := wli.rec.Flow[len(wli.rec.Flow)-1]

	wli.engine.completeState(ctx, wli.rec, nextState, errCode, true)

	attempt := wli.rec.Attempts + 1
	if multiplier == 0 {
		multiplier = 1.0
	}

	// Set a default delay string if none is provided to parse delay properly.
	if delayString == "" {
		delayString = "PT1S"
	}

	delay, err := duration.ParseISO8601(delayString)
	if err != nil {
		return NewInternalError(err)
	}

	multiplier = math.Pow(multiplier, float64(attempt))

	now := time.Now()
	t := delay.Shift(now)
	duration := t.Sub(now)
	duration = time.Duration(float64(duration) * multiplier)

	schedule := now.Add(duration)
	deadline := schedule.Add(time.Second * 5)
	duration = wli.logic.Deadline().Sub(now)
	deadline = deadline.Add(duration)

	oldController := wli.rec.Controller
	newController := wli.engine.server.hostname

	wf := wli.rec.Edges.Workflow

	var rec *ent.WorkflowInstance
	rec, err = wli.rec.Update().
		SetAttempts(attempt).
		SetDeadline(deadline).
		SetController(newController).
		Save(ctx)
	if err != nil {
		return err
	}
	wli.rec = rec
	wli.rec.Edges.Workflow = wf

	wli.ScheduleSoftTimeout(oldController, deadline)

	if duration < time.Second*5 {
		go func() {
			time.Sleep(duration)
			wli.Log("Retrying failed workflow state.")
			wli.Transition(ctx, nextState, attempt)
		}()
	} else {
		wli.Log("Scheduling a retry for the failed workflow state at approximate time: %s.", schedule.UTC().String())
		err = wli.engine.scheduleRetry(wli.id, nextState, wli.step, schedule)
		if err != nil {
			return err
		}
		wli.Close()
	}

	return nil

}

func (wli *workflowLogicInstance) scheduleTimeout(oldController string, t time.Time, soft bool) {

	var err error
	deadline := t

	prefixes := []string{"soft", "hard"}
	prefix := prefixes[1]
	if soft {
		prefix = prefixes[0]
	}

	oldId := fmt.Sprintf("timeout:%s:%s:%d", wli.id, prefix, wli.step-1)
	id := fmt.Sprintf("timeout:%s:%s:%d", wli.id, prefix, wli.step)
	if wli.step == 0 {
		id = fmt.Sprintf("timeout:%s:%s", wli.id, prefix)
	}

	// cancel existing timeouts

	wli.engine.timer.deleteTimerByName(oldController, wli.engine.server.hostname, oldId)
	wli.engine.timer.deleteTimerByName(oldController, wli.engine.server.hostname, id)

	// schedule timeout

	args := &timeoutArgs{
		InstanceId: wli.id,
		Step:       wli.step,
		Soft:       false,
	}

	data, err := json.Marshal(args)
	if err != nil {
		log.Error(err)
	}

	err = wli.engine.timer.addOneShot(id, timeoutFunction, deadline, data)
	if err != nil {
		log.Error(err)
	}

}

func (wli *workflowLogicInstance) ScheduleHardTimeout(oldController string, t time.Time) {
	wli.scheduleTimeout(oldController, t, false)
}

func (wli *workflowLogicInstance) ScheduleSoftTimeout(oldController string, t time.Time) {
	wli.scheduleTimeout(oldController, t, true)
}

func (wli *workflowLogicInstance) Transition(ctx context.Context, nextState string, attempt int) {

	oldController := wli.rec.Controller

	if wli.step == 0 {
		t := time.Now()
		tSoft := time.Now().Add(time.Minute * 15)
		tHard := time.Now().Add(time.Minute * 20)
		if wli.wf.Timeouts != nil {
			s := wli.wf.Timeouts.Interrupt
			if s != "" {
				d, err := duration.ParseISO8601(s)
				if err != nil {
					log.Error(err)
					wli.Close()
					return
				}
				tSoft = d.Shift(t)
				tHard = tSoft.Add(time.Minute * 5)
			}
			s = wli.wf.Timeouts.Kill
			if s != "" {
				d, err := duration.ParseISO8601(s)
				if err != nil {
					log.Error(err)
					wli.Close()
					return
				}
				tHard = d.Shift(t)
			}
		}
		wli.ScheduleSoftTimeout(oldController, tSoft)
		wli.ScheduleHardTimeout(oldController, tHard)
	}

	if len(wli.rec.Flow) != wli.step {
		err := errors.New("workflow logic instance aborted for being tardy")
		log.Error(err)
		wli.Close()
		return
	}

	data, err := json.Marshal(wli.data)
	if err != nil {
		err = fmt.Errorf("engine cannot marshal state data for storage: %v", err)
		log.Error(err)
		wli.Close()
		return
	}

	if nextState == "" {
		panic("don't call this function with an empty nextState")
	}

	states := wli.wf.GetStatesMap()
	state, exists := states[nextState]
	if !exists {
		err = fmt.Errorf("workflow cannot resolve transition: %s", nextState)
		log.Error(err)
		wli.Close()
		return
	}

	init, exists := wli.engine.stateLogics[state.GetType()]
	if !exists {
		err = fmt.Errorf("engine cannot resolve state type: %s", state.GetType().String())
		log.Error(err)
		wli.Close()
		return
	}

	stateLogic, err := init(wli.wf, state)
	if err != nil {
		err = fmt.Errorf("cannot initialize state logic: %v", err)
		log.Error(err)
		wli.Close()
		return
	}
	wli.logic = stateLogic

	flow := append(wli.rec.Flow, nextState)
	wli.step++
	deadline := stateLogic.Deadline()

	t := time.Now()

	wf := wli.rec.Edges.Workflow

	var rec *ent.WorkflowInstance
	rec, err = wli.rec.Update().
		SetDeadline(deadline).
		SetController(wli.engine.server.hostname).
		SetStateBeginTime(t).
		SetNillableMemory(nil).
		SetAttempts(attempt).
		SetFlow(flow).
		SetStateData(string(data)).
		Save(ctx)
	if err != nil {
		log.Error(err)
		wli.Close()
		return
	}
	wli.rec = rec
	wli.rec.Edges.Workflow = wf

	wli.ScheduleSoftTimeout(oldController, deadline)

	wli.engine.runState(ctx, wli, nil, nil)

}

func InstanceMemory(rec *ent.WorkflowInstance) ([]byte, error) {

	if rec.Memory == "" {
		return nil, nil
	}

	savedata, err := base64.StdEncoding.DecodeString(rec.Memory)
	if err != nil {
		err = fmt.Errorf("cannot decode the savedata: %v", err)
		log.Error(err)
		return nil, err
	}

	return savedata, nil

}
