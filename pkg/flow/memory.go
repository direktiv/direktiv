package flow

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/google/uuid"
)

type instanceMemory struct {
	engine *engine

	data   interface{}
	memory interface{}
	logic  stateLogic

	// stores the events to be fired on schedule
	eventQueue []string

	tags       map[string]string
	userLogger *slog.Logger

	instance   *enginerefactor.Instance
	updateArgs *instancestore.UpdateInstanceDataArgs
}

func (im *instanceMemory) Namespace() *database.Namespace {
	return &database.Namespace{
		ID:   im.instance.Instance.NamespaceID,
		Name: im.instance.TelemetryInfo.NamespaceName,
	}
}

func (im *instanceMemory) flushUpdates(ctx context.Context) error {
	data, err := json.Marshal(im.updateArgs)
	if err != nil {
		panic(err)
	}

	if string(data) == `{}` {
		return nil
	}

	im.updateArgs.Server = im.engine.ID

	// NOTE: no need to make this serializable because only a single operation is performed. If we
	// 		expand the number of queries here in the future we should make it serializable. Be
	// 		warned however that making this serializable opens us up to serialization failures, and
	//		therefore we will need to test heavily and potentially implement retries.
	tx, err := im.engine.flow.beginSqlTx(ctx) /*&sql.TxOptions{
		Isolation: sql.LevelSerializable,
	}*/if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.InstanceStore().ForInstanceID(im.instance.Instance.ID).UpdateInstanceData(ctx, im.updateArgs)
	if err != nil {
		if strings.Contains(err.Error(), "got 0") {
			return errors.New("node no longer believes it should modify this instance")
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	im.updateArgs = new(instancestore.UpdateInstanceDataArgs)
	im.updateArgs.Server = im.engine.ID

	im.engine.pubsub.NotifyInstance(im.instance.Instance.ID)
	im.engine.pubsub.NotifyInstances(im.Namespace())

	return nil
}

func (im *instanceMemory) ID() uuid.UUID {
	return im.instance.Instance.ID
}

func (im *instanceMemory) Controller() string {
	return im.instance.RuntimeInfo.Controller
}

func (im *instanceMemory) Model() (*model.Workflow, error) {
	data := im.instance.Instance.Definition

	workflow := new(model.Workflow)

	err := workflow.Load(data)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}

func (im *instanceMemory) Step() int {
	return len(im.instance.RuntimeInfo.Flow)
}

func (im *instanceMemory) Status() string {
	return im.instance.Instance.Status.String()
}

func (im *instanceMemory) Flow() []string {
	return im.instance.RuntimeInfo.Flow
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

	data := im.MarshalMemory()

	im.instance.Instance.StateMemory = []byte(data)
	im.updateArgs.StateMemory = &im.instance.Instance.StateMemory
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
	return im.instance.Instance.ErrorCode
}

func (im *instanceMemory) ErrorMessage() string {
	return string(im.instance.Instance.ErrorMessage)
}

func (im *instanceMemory) StateBeginTime() time.Time {
	return im.instance.RuntimeInfo.StateBeginTime
}

func (im *instanceMemory) replaceData(x map[string]interface{}) {
	im.data = x
	data := im.MarshalData()
	im.instance.Instance.LiveData = []byte(data)
	im.updateArgs.LiveData = &im.instance.Instance.LiveData
}

func (im *instanceMemory) StoreData(key string, val interface{}) error {
	m, ok := im.data.(map[string]interface{})
	if !ok {
		return derrors.NewInternalError(errors.New("unable to store data because state data isn't a valid JSON object"))
	}

	m[key] = val

	im.replaceData(m)

	return nil
}

func (im *instanceMemory) GetAttributes() map[string]string {
	tags := im.instance.GetAttributes(recipient.Instance)
	for k, v := range im.tags {
		tags[k] = v
	}
	if im.logic != nil {
		tags["state-id"] = im.logic.GetID()
		tags["state-type"] = im.logic.GetType().String()
	}

	pi := im.engine.InstanceCaller(im)
	if pi != nil {
		a := strings.Split(pi.State, ":")
		if len(a) >= 1 && a[0] != "" {
			tags["invoker-workflow"] = a[0]
		}
		if len(a) > 1 {
			tags["invoker-state-id"] = a[1]
		}
	}

	return tags
}

func (im *instanceMemory) WithTags(ctx context.Context) context.Context {
	ctx = im.instance.WithTags(ctx)
	tags, ok := ctx.Value(core.LogTagsKey).([]interface{})
	if !ok {
		tags = make([]interface{}, 0)
	}
	if im.logic != nil {
		tags = append(tags, "state", im.logic.GetID())
	}
	return context.WithValue(ctx, core.LogTagsKey, tags)
}

func (im *instanceMemory) GetState() string {
	tags := im.instance.GetAttributes(recipient.Instance)
	if im.logic != nil {
		return fmt.Sprintf("%s:%s", tags["workflow"], im.logic.GetID())
	}
	return tags["workflow"]
}

func (engine *engine) getInstanceMemory(ctx context.Context, id uuid.UUID) (*instanceMemory, error) {
	tx, err := engine.flow.beginSqlTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	idata, err := tx.InstanceStore().ForInstanceID(id).GetMost(ctx)
	if err != nil {
		return nil, err
	}

	if idata.Server != engine.ID {
		if time.Now().Add(-1 * engineOwnershipTimeout).Before(idata.UpdatedAt) {
			return nil, errors.New("instance appears to be under control of another node")
		}

		// TODO: alan DIR-1313
		// we need to ensure there's an auto-reattempter somewhere in the code
	}

	if idata.EndedAt != nil && !idata.EndedAt.IsZero() {
		return nil, derrors.NewInternalError(fmt.Errorf("aborting workflow logic: database records instance terminated"))
	}

	err = tx.InstanceStore().ForInstanceID(id).UpdateInstanceData(ctx, &instancestore.UpdateInstanceDataArgs{
		BypassOwnershipCheck: true,
		Server:               engine.ID,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	instance, err := enginerefactor.ParseInstanceData(idata)
	if err != nil {
		return nil, err
	}

	im := new(instanceMemory)
	im.engine = engine
	im.instance = instance
	im.updateArgs = new(instancestore.UpdateInstanceDataArgs)
	im.updateArgs.Server = engine.ID

	err = json.Unmarshal(im.instance.Instance.LiveData, &im.data)
	if err != nil {
		engine.CrashInstance(ctx, im, derrors.NewUncatchableError("", err.Error()))
		return nil, err
	}

	err = json.Unmarshal(im.instance.Instance.StateMemory, &im.memory)
	if err != nil {
		engine.CrashInstance(ctx, im, derrors.NewUncatchableError("", err.Error()))
		return nil, err
	}

	flow := im.instance.RuntimeInfo.Flow
	stateID := ""

	if len(flow) > 0 {
		stateID = flow[len(flow)-1]
	}

	err = engine.loadStateLogic(im, stateID)
	if err != nil {
		engine.CrashInstance(ctx, im, err)
		return nil, err
	}

	return im, nil
}

func (engine *engine) InstanceCaller(im *instanceMemory) *enginerefactor.ParentInfo {
	di := im.instance.DescentInfo
	if len(di.Descent) == 0 {
		return nil
	}

	return &di.Descent[len(di.Descent)-1]
}

func (engine *engine) StoreMetadata(ctx context.Context, im *instanceMemory, data string) {
	im.instance.Instance.Metadata = []byte(data)
	im.updateArgs.Metadata = &im.instance.Instance.Metadata
}

func (engine *engine) freeArtefacts(im *instanceMemory) {
	engine.timers.deleteTimersForInstance(im.ID().String())

	err := engine.events.deleteInstanceEventListeners(context.Background(), im)
	if err != nil {
		engine.sugar.Error(err)
	}
}

func (engine *engine) freeMemory(ctx context.Context, im *instanceMemory) error {
	for i := range im.eventQueue {
		err := engine.events.flushEvent(true)
		if err != nil {
			return fmt.Errorf("failed to flush event [%d]: %w", i, err)
		}
	}

	im.eventQueue = make([]string, 0)

	err := im.flushUpdates(ctx)
	if err != nil {
		return err
	}

	engine.deregisterScheduled(im.ID())

	return nil
}

func (engine *engine) forceFreeCriticalMemory(ctx context.Context, im *instanceMemory) {
	err := im.flushUpdates(ctx)
	if err != nil {
		engine.sugar.Errorf("failed to force flush updates during instance crash: %v.", err)
	}

	engine.deregisterScheduled(im.ID())
}
