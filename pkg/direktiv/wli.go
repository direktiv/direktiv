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

	"github.com/itchyny/gojq"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/senseyeio/duration"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/ent"
	"github.com/vorteil/direktiv/pkg/dlog"
	"github.com/vorteil/direktiv/pkg/model"
)

type workflowLogicInstance struct {
	engine    *workflowEngine
	data      interface{}
	startData []byte
	wf        *model.Workflow
	rec       *ent.WorkflowInstance
	step      int

	namespace string
	id        string
	lockConn  *sql.Conn
	logic     stateLogic
	logger    dlog.Logger
}

func (wli *workflowLogicInstance) Close() error {
	return wli.logger.Close()
}

func (wli *workflowLogicInstance) Raise(ctx context.Context, cerr *CatchableError) error {

	var err error

	if wli.rec.ErrorCode != "" {
		wli.rec, err = wli.rec.Update().
			SetStatus("failed").
			SetErrorCode(cerr.Code).
			SetErrorMessage(cerr.Message).
			Save(ctx)
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

	if wli.rec.ErrorCode == "" {
		wli.rec, err = wli.rec.Update().
			SetStatus(status).
			SetEndTime(time.Now()).
			SetErrorCode(code).
			SetErrorMessage(message).
			Save(ctx)
		return err
	}

	return nil

}

func (wli *workflowLogicInstance) wakeCaller(data []byte) {

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

		err = wli.engine.wakeCaller(msg)
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

	return output, nil

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

func (wli *workflowLogicInstance) Log(msg string, a ...interface{}) {
	wli.logger.Info(fmt.Sprintf(msg, a...))
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

func (wli *workflowLogicInstance) Retry(ctx context.Context, delayString string, multiplier float64) error {

	var err error
	var x interface{}

	err = json.Unmarshal([]byte(wli.rec.StateData), &x)
	if err != nil {
		return NewInternalError(err)
	}

	wli.data = x

	nextState := wli.rec.Flow[len(wli.rec.Flow)-1]

	attempt := wli.rec.Attempts + 1
	if multiplier == 0 {
		multiplier = 1.0
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

	var rec *ent.WorkflowInstance
	rec, err = wli.rec.Update().SetDeadline(deadline).Save(ctx)
	if err != nil {
		return err
	}
	wli.rec = rec
	wli.ScheduleSoftTimeout(deadline)

	if duration < time.Second*5 {
		time.Sleep(duration)
		wli.Log("Retrying failed workflow state.")
		go wli.Transition(nextState, attempt)
	} else {
		wli.Log("Scheduling a retry for the failed workflow state at approximate time: %s.", schedule.UTC().String())
		err = wli.engine.scheduleRetry(wli.id, nextState, wli.step, schedule)
		if err != nil {
			return err
		}
	}

	return nil

}

func (wli *workflowLogicInstance) scheduleTimeout(t time.Time, soft bool) {

	var err error
	deadline := t
	oldId := fmt.Sprintf("timeout:%s:%d", wli.id, wli.step-1)
	id := fmt.Sprintf("timeout:%s:%d", wli.id, wli.step)

	if wli.step == 0 {
		id = fmt.Sprintf("timeout:%s", wli.id)
	}

	// cancel existing timeouts

	wli.engine.timer.actionTimerByName(oldId, deleteTimerAction)
	wli.engine.timer.actionTimerByName(id, deleteTimerAction)

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

	_, err = wli.engine.timer.addOneShot(id, timeoutFunction, deadline, data)
	if err != nil {
		log.Error(err)
	}

}

func (wli *workflowLogicInstance) ScheduleHardTimeout(t time.Time) {
	wli.scheduleTimeout(t, false)
}

func (wli *workflowLogicInstance) ScheduleSoftTimeout(t time.Time) {
	wli.scheduleTimeout(t, true)
}

func (wli *workflowLogicInstance) Transition(nextState string, attempt int) {

	ctx, err := wli.lock(time.Second * 5)
	if err != nil {
		log.Error(err)
		return
	}

	defer wli.unlock()

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
					return
				}
				tHard = d.Shift(t)
			}
		}
		wli.ScheduleSoftTimeout(tSoft)
		wli.ScheduleHardTimeout(tHard)
	}

	if len(wli.rec.Flow) != wli.step {
		err = errors.New("workflow logic instance aborted for being tardy")
		log.Error(err)
		return
	}

	data, err := json.Marshal(wli.data)
	if err != nil {
		err = fmt.Errorf("engine cannot marshal state data for storage: %v", err)
		log.Error(err)
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
		return
	}

	init, exists := wli.engine.stateLogics[state.GetType()]
	if !exists {
		err = fmt.Errorf("engine cannot resolve state type: %s", state.GetType().String())
		log.Error(err)
		return
	}

	stateLogic, err := init(wli.wf, state)
	if err != nil {
		err = fmt.Errorf("cannot initialize state logic: %v", err)
		log.Error(err)
		return
	}
	wli.logic = stateLogic

	flow := append(wli.rec.Flow, nextState)
	wli.step++
	deadline := stateLogic.Deadline()

	var rec *ent.WorkflowInstance
	rec, err = wli.rec.Update().
		SetDeadline(deadline).
		SetNillableMemory(nil).
		SetAttempts(attempt).
		SetFlow(flow).
		SetStateData(string(data)).
		Save(ctx)
	if err != nil {
		log.Error(err)
		return
	}
	wli.rec = rec
	wli.ScheduleSoftTimeout(deadline)

	go func(we *workflowEngine, id, state string, step int) {
		ctx, wli, err := we.loadWorkflowLogicInstance(wli.id, wli.step)
		if err != nil {
			log.Errorf("cannot load workflow logic instance: %v", err)
			return
		}
		go wli.engine.runState(ctx, wli, nil, nil)
	}(wli.engine, wli.id, nextState, wli.step)

	return

}
