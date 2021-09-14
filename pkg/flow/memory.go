package flow

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vorteil/direktiv/pkg/flow/ent"
	entinst "github.com/vorteil/direktiv/pkg/flow/ent/instance"
	"github.com/vorteil/direktiv/pkg/model"
)

type instanceMemory struct {
	lock   *sql.Conn
	in     *ent.Instance
	data   interface{}
	memory interface{}
	logic  stateLogic

	// stores the events to be fired on schedule
	eventQueue []string
}

func (im *instanceMemory) ID() uuid.UUID {

	return im.in.ID

}

func (im *instanceMemory) Controller() string {

	return im.in.Edges.Runtime.Controller

}

func (im *instanceMemory) Model() (*model.Workflow, error) {

	data := im.in.Edges.Revision.Source

	workflow := new(model.Workflow)

	err := workflow.Load(data)
	if err != nil {
		return nil, err
	}

	return workflow, nil

}

func (im *instanceMemory) Unwrap() {

	defer func() {
		_ = recover()
	}()

	in := im.in.Unwrap()
	im.in = in
	im.in.Edges.Runtime = im.in.Edges.Runtime.Unwrap()

}

func (im *instanceMemory) Step() int {
	return len(im.in.Edges.Runtime.Flow)
}

func (im *instanceMemory) Status() string {
	return im.in.Status
}

func (im *instanceMemory) Flow() []string {
	return im.in.Edges.Runtime.Flow
}

func (im *instanceMemory) MarshalData() string {

	data, err := json.Marshal(im.data)
	if err != nil {
		panic(err)
	}

	return string(data)

}

func (im *instanceMemory) MarshalOutput() string {

	if im.Status() == "complete" {
		return im.MarshalData()
	}

	return ""

}

func (im *instanceMemory) SetMemory(x interface{}) {
	im.memory = x
}

func (im *instanceMemory) GetMemory() interface{} {
	return im.memory
}

func (im *instanceMemory) MarshalMemory() string {

	data, err := json.Marshal(im.memory)
	if err != nil {
		panic(err)
	}

	return string(data)

}

func (im *instanceMemory) UnmarshalMemory(x interface{}) error {

	if im.memory == nil {
		return nil
	}

	data, err := json.Marshal(im.memory)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, x)
	if err != nil {
		return err
	}

	return nil

}

func (im *instanceMemory) ErrorCode() string {
	return im.in.ErrorCode
}

func (im *instanceMemory) ErrorMessage() string {
	return im.in.ErrorMessage
}

func (im *instanceMemory) StateBeginTime() time.Time {
	return im.in.Edges.Runtime.StateBeginTime
}

func (im *instanceMemory) StoreData(key string, val interface{}) error {

	m, ok := im.data.(map[string]interface{})
	if !ok {
		return NewInternalError(errors.New("unable to store data because state data isn't a valid JSON object"))
	}

	m[key] = val

	return nil

}

func (engine *engine) getInstanceMemory(ctx context.Context, inc *ent.InstanceClient, id string) (*instanceMemory, error) {

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	in, err := inc.Query().Where(entinst.IDEQ(uid)).WithNamespace().WithWorkflow().WithRevision().WithRuntime().Only(ctx)
	if err != nil {
		return nil, err
	}

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

	flow := im.in.Edges.Runtime.Flow
	stateID := flow[len(flow)-1]

	err = engine.loadStateLogic(im, stateID)
	if err != nil {
		return nil, err
	}

	return im, nil

}

func (engine *engine) loadInstanceMemory(id string, step int) (context.Context, *instanceMemory, error) {

	ctx, conn, err := engine.lock(id, defaultLockWait)
	if err != nil {
		return nil, nil, err
	}

	im, err := engine.getInstanceMemory(ctx, engine.db.Instance, id)
	if err != nil {
		engine.unlock(id, conn)
		return nil, nil, err
	}

	im.lock = conn

	if !im.in.EndAt.IsZero() {
		engine.InstanceUnlock(im)
		return nil, nil, NewInternalError(fmt.Errorf("aborting workflow logic: database records instance terminated"))
	}

	if step >= 0 && step != im.Step() {
		engine.InstanceUnlock(im)
		return nil, nil, NewInternalError(fmt.Errorf("aborting workflow logic: steps out of sync (expect/actual - %d/%d)", step, im.Step()))
	}

	return ctx, im, nil

}

func (engine *engine) InstanceCaller(ctx context.Context, im *instanceMemory) *subflowCaller {

	var err error

	str := im.in.Edges.Runtime.CallerData
	if str == "" || str == "cron" {
		return nil
	}

	output := new(subflowCaller)
	err = json.Unmarshal([]byte(str), output)
	if err != nil {
		engine.sugar.Error(err)
		return nil
	}

	return output

}

func (engine *engine) InstanceNamespace(ctx context.Context, im *instanceMemory) (*ent.Namespace, error) {

	var err error
	var ns *ent.Namespace

	if im.in.Edges.Namespace != nil {
		goto out
	}

	ns, err = im.in.Namespace(ctx)
	if err != nil {
		return nil, err
	}

	im.in.Edges.Namespace = ns

out:
	return im.in.Edges.Namespace, nil

}

func (engine *engine) InstanceWorkflow(ctx context.Context, im *instanceMemory) (*ent.Workflow, error) {

	var err error
	var wf *ent.Workflow

	if im.in.Edges.Workflow != nil {
		goto out
	}

	wf, err = im.in.Workflow(ctx)
	if err != nil {
		return nil, err
	}

	im.in.Edges.Workflow = wf

out:
	return im.in.Edges.Workflow, nil

}

func (engine *engine) FreeInstanceMemory(im *instanceMemory) {

	engine.freeResources(im)

	if im.lock != nil {
		engine.InstanceUnlock(im)
	}

	engine.timers.deleteTimersForInstance(im.ID().String())

	err := engine.events.deleteWorkflowEventListenerByInstanceID(im.ID())
	if err != nil {
		engine.sugar.Error(err)
	}

}

func (engine *engine) freeResources(im *instanceMemory) {

	for i := range im.eventQueue {
		engine.events.flushEvent(im.eventQueue[i], im.in.Edges.Namespace, true)
	}

	// TODO: do we actually want to delete variables here? There could be value in keeping them around for a little while.

	// var namespace, workflow, instance string
	// namespace = rec.Edges.Workflow.Edges.Namespace.ID
	// workflow = rec.Edges.Workflow.ID.String()
	// instance = rec.InstanceID
	// we.server.variableStorage.DeleteAllInScope(context.Background(), namespace, workflow, instance)

}
