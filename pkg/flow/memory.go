package flow

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
)

type instanceMemory struct {
	engine *engine

	lock   *sql.Conn
	data   interface{}
	memory interface{}
	logic  stateLogic

	// stores the events to be fired on schedule
	eventQueue []string

	// stores row updates to fire all at once
	instanceUpdater *ent.InstanceUpdateOne
	runtimeUpdater  *ent.InstanceRuntimeUpdateOne

	// tx      database.Transaction
	cached  *database.CacheData
	runtime *database.InstanceRuntime

	tags map[string]string
}

func (im *instanceMemory) getInstanceUpdater() *ent.InstanceUpdateOne {
	if im.instanceUpdater == nil {
		clients := im.engine.edb.Clients(context.Background())
		im.instanceUpdater = clients.Instance.UpdateOneID(im.cached.Instance.ID)
	}

	return im.instanceUpdater
}

func (im *instanceMemory) getRuntimeUpdater() *ent.InstanceRuntimeUpdateOne {
	if im.runtimeUpdater == nil {
		clients := im.engine.edb.Clients(context.Background())
		im.runtimeUpdater = clients.InstanceRuntime.UpdateOneID(im.runtime.ID)
	}

	return im.runtimeUpdater
}

func (im *instanceMemory) flushUpdates(ctx context.Context) error {
	var changes bool

	if im.runtimeUpdater != nil {
		changes = true

		updater := im.runtimeUpdater
		im.runtimeUpdater = nil

		rt, err := updater.Save(ctx)
		if err != nil {
			return err
		}

		im.runtime = entwrapper.EntInstanceRuntime(rt)
	}

	if im.instanceUpdater != nil {
		changes = true

		updater := im.instanceUpdater
		im.instanceUpdater = nil

		in, err := updater.Save(ctx)
		if err != nil {
			return err
		}

		im.cached.Instance = entwrapper.EntInstance(in)
		im.cached.Instance.Namespace = im.cached.Namespace.ID
		if im.cached.File != nil {
			im.cached.Instance.Workflow = im.cached.File.ID
		}
		if im.cached.Revision != nil {
			im.cached.Instance.Revision = im.cached.Revision.ID
		}
		im.cached.Instance.Runtime = im.runtime.ID

		err = im.engine.database.FlushInstance(ctx, im.cached.Instance)
		if err != nil {
			return err
		}
	}

	if changes {
		im.engine.pubsub.NotifyInstance(im.cached.Instance)
		im.engine.pubsub.NotifyInstances(im.cached.Namespace)
	}

	return nil
}

func (im *instanceMemory) ID() uuid.UUID {
	return im.cached.Instance.ID
}

func (im *instanceMemory) Controller() string {
	return im.runtime.Controller
}

func (im *instanceMemory) Model() (*model.Workflow, error) {
	data := im.cached.Revision.Data

	workflow := new(model.Workflow)

	err := workflow.Load(data)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}

func (im *instanceMemory) Step() int {
	return len(im.runtime.Flow)
}

func (im *instanceMemory) Status() string {
	return im.cached.Instance.Status
}

func (im *instanceMemory) Flow() []string {
	return im.runtime.Flow
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

func (im *instanceMemory) setMemory(x interface{}) {
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
	return im.cached.Instance.ErrorCode
}

func (im *instanceMemory) ErrorMessage() string {
	return im.cached.Instance.ErrorMessage
}

func (im *instanceMemory) StateBeginTime() time.Time {
	return im.runtime.StateBeginTime
}

func (im *instanceMemory) StoreData(key string, val interface{}) error {
	m, ok := im.data.(map[string]interface{})
	if !ok {
		return derrors.NewInternalError(errors.New("unable to store data because state data isn't a valid JSON object"))
	}

	m[key] = val

	return nil
}

func (im *instanceMemory) GetAttributes() map[string]string {
	tags := im.cached.GetAttributes(recipient.Instance)
	for k, v := range im.tags {
		tags[k] = v
	}
	if im.logic != nil {
		tags["state-id"] = im.logic.GetID()
		tags["state-type"] = im.logic.GetType().String()
	}
	a := strings.Split(im.cached.Instance.InvokerState, ":")
	if len(a) >= 1 && a[0] != "" {
		tags["invoker-workflow"] = a[0]
	}
	if len(a) > 1 {
		tags["invoker-state-id"] = a[1]
	}
	return tags
}

func (im *instanceMemory) GetState() string {
	tags := im.cached.GetAttributes(recipient.Instance)
	if im.logic != nil {
		return fmt.Sprintf("%s:%s", tags["workflow"], im.logic.GetID())
	}
	return tags["workflow"]
}

func (engine *engine) getInstanceMemory(ctx context.Context, id string) (*instanceMemory, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	cached := new(database.CacheData)

	// TODO: alan, need to load all of this information in a more performant manner
	err = engine.database.Instance(ctx, cached, uid)
	if err != nil {
		return nil, err
	}

	fStore, _, _, rollback, err := engine.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, revision, err := fStore.GetRevision(ctx, cached.Instance.Revision)
	if err != nil {
		if !errors.Is(err, filestore.ErrNotFound) {
			return nil, err
		}
	}

	cached.File = file
	cached.Revision = revision

	rt, err := engine.database.InstanceRuntime(ctx, cached.Instance.Runtime)
	if err != nil {
		return nil, err
	}

	im := new(instanceMemory)
	im.engine = engine
	im.cached = cached
	im.runtime = rt

	defer func() {
		e := im.flushUpdates(ctx)
		if e != nil {
			err = e
		}
	}()

	if cached.File == nil || cached.Revision == nil {
		err = errors.New("the workflow or revision was deleted")
		engine.CrashInstance(ctx, im, derrors.NewUncatchableError("", err.Error()))
		return nil, err
	}

	err = json.Unmarshal([]byte(im.runtime.Data), &im.data)
	if err != nil {
		engine.CrashInstance(ctx, im, derrors.NewUncatchableError("", err.Error()))
		return nil, err
	}

	err = json.Unmarshal([]byte(im.runtime.Memory), &im.memory)
	if err != nil {
		engine.CrashInstance(ctx, im, derrors.NewUncatchableError("", err.Error()))
		return nil, err
	}

	flow := im.runtime.Flow
	stateID := flow[len(flow)-1]

	err = engine.loadStateLogic(im, stateID)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return nil, err
	}

	return im, nil
}

func (engine *engine) loadInstanceMemory(id string, step int) (context.Context, *instanceMemory, error) {
	ctx, conn, err := engine.lock(id, defaultLockWait)
	if err != nil {
		return nil, nil, err
	}

	im, err := engine.getInstanceMemory(ctx, id)
	if err != nil {
		engine.unlock(id, conn)
		return nil, nil, err
	}

	im.lock = conn

	if !im.cached.Instance.EndAt.IsZero() {
		engine.InstanceUnlock(im)
		return nil, nil, derrors.NewInternalError(fmt.Errorf("aborting workflow logic: database records instance terminated"))
	}

	if step >= 0 && step != im.Step() {
		engine.InstanceUnlock(im)
		return nil, nil, derrors.NewInternalError(fmt.Errorf("aborting workflow logic: steps out of sync (expect/actual - %d/%d)", step, im.Step()))
	}

	return ctx, im, nil
}

func (engine *engine) InstanceCaller(ctx context.Context, im *instanceMemory) *subflowCaller {
	var err error

	str := im.runtime.CallerData
	if str == "" || str == util.CallerCron {
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

func (engine *engine) StoreMetadata(ctx context.Context, im *instanceMemory, data string) {
	updater := im.getRuntimeUpdater()
	updater = updater.SetMetadata(data)
	im.runtime.Metadata = data
	im.runtimeUpdater = updater
}

func (engine *engine) FreeInstanceMemory(im *instanceMemory) {
	engine.freeResources(im)

	if im.lock != nil {
		engine.InstanceUnlock(im)
	}

	engine.timers.deleteTimersForInstance(im.ID().String())

	ctx := context.Background()

	err := engine.events.deleteInstanceEventListeners(ctx, im.cached)
	if err != nil {
		engine.sugar.Error(err)
	}
}

func (engine *engine) freeResources(im *instanceMemory) {
	ctx := context.Background()

	for i := range im.eventQueue {
		err := engine.events.flushEvent(ctx, im.eventQueue[i], im.cached.Namespace, true)
		if err != nil {
			engine.sugar.Errorf("Failed to flush event: %v.", err)
		}
	}

	// do we actually want to delete variables here? There could be value in keeping them around for a little while.

	// var namespace, workflow, instance string
	// namespace = rec.Edges.Workflow.Edges.Namespace.ID
	// workflow = rec.Edges.Workflow.ID.String()
	// instance = rec.InstanceID
	// we.server.variableStorage.DeleteAllInScope(context.Background(), namespace, workflow, instance)
}
