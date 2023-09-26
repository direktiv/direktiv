package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	log "github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
)

// TEMPORARY EVERYTHING

func (im *instanceMemory) BroadcastCloudevent(ctx context.Context, event *cloudevents.Event, dd int64) error {
	return im.engine.events.BroadcastCloudevent(ctx, im.Namespace(), event, dd)
}

func (im *instanceMemory) GetVariables(ctx context.Context, vars []states.VariableSelector) ([]states.Variable, error) {
	x := make([]states.Variable, 0)

	tx, err := im.engine.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, selector := range vars {
		if selector.Scope == "" || selector.Scope == util.VarScopeInstance || selector.Scope == util.VarScopeWorkflow || selector.Scope == util.VarScopeNamespace {
			if selector.Scope == "" {
				selector.Scope = util.VarScopeNamespace
			}

			var item *core.RuntimeVariable

			switch selector.Scope {
			case util.VarScopeInstance:
				item, err = tx.DataStore().RuntimeVariables().GetByInstanceAndName(ctx, im.instance.Instance.ID, selector.Key)
			case util.VarScopeWorkflow:
				item, err = tx.DataStore().RuntimeVariables().GetByWorkflowAndName(ctx, im.instance.Instance.NamespaceID, im.instance.Instance.WorkflowPath, selector.Key)
			case util.VarScopeNamespace:
				item, err = tx.DataStore().RuntimeVariables().GetByNamespaceAndName(ctx, im.instance.Instance.NamespaceID, selector.Key)
			default:
				return nil, derrors.NewInternalError(errors.New("invalid scope"))
			}
			if errors.Is(err, datastore.ErrNotFound) {
				x = append(x, states.Variable{
					Scope: selector.Scope,
					Key:   selector.Key,
					Data:  []byte{},
				})
			} else if err != nil {
				return nil, derrors.NewInternalError(err)
			} else {
				data, err := tx.DataStore().RuntimeVariables().LoadData(ctx, item.ID)
				if err != nil {
					return nil, derrors.NewInternalError(err)
				}
				x = append(x, states.Variable{
					Scope: selector.Scope,
					Key:   selector.Key,
					Data:  data,
				})
			}

			continue
		}

		if selector.Scope == util.VarScopeFileSystem {
			file, err := tx.FileStore().ForRootNamespaceAndName(im.instance.Instance.NamespaceID, defaultRootName).GetFile(ctx, selector.Key)
			if errors.Is(err, filestore.ErrNotFound) {
				x = append(x, states.Variable{
					Scope: selector.Scope,
					Key:   selector.Key,
					Data:  make([]byte, 0),
				})
			} else if err != nil {
				return nil, err
			} else {
				// TODO: alan, maybe need to enhance the GetData function to also return us some information like mime type, checksum, and size
				if file.Typ == filestore.FileTypeDirectory {
					return nil, model.ErrVarNotFile
				}
				rc, err := tx.FileStore().ForFile(file).GetData(ctx)
				if err != nil {
					return nil, err
				}
				defer func() { _ = rc.Close() }()
				data, err := io.ReadAll(rc)
				if err != nil {
					return nil, err
				}
				err = rc.Close()
				if err != nil {
					return nil, err
				}
				x = append(x, states.Variable{
					Scope: selector.Scope,
					Key:   selector.Key,
					Data:  data,
				})
			}

			continue
		}
	}

	return x, nil
}

func (im *instanceMemory) ListenForEvents(ctx context.Context, events []*model.ConsumeEventDefinition, all bool) error {
	err := im.engine.events.deleteInstanceEventListeners(ctx, im)
	if err != nil {
		return err
	}

	err = im.engine.events.listenForEvents(ctx, im, events, all)
	if err != nil {
		return err
	}

	return nil
}

func (im *instanceMemory) Log(ctx context.Context, level log.Level, a string, x ...interface{}) {
	switch level {
	case log.Info:
		im.engine.logger.Infof(ctx, im.GetInstanceID(), im.GetAttributes(), a, x...)
	case log.Debug:
		im.engine.logger.Debugf(ctx, im.GetInstanceID(), im.GetAttributes(), a, x...)
	case log.Error:
		im.engine.logger.Errorf(ctx, im.GetInstanceID(), im.GetAttributes(), a, x...)
	case log.Panic:
		im.engine.logger.Errorf(ctx, im.GetInstanceID(), im.GetAttributes(), a, x...)
	}
}

func (im *instanceMemory) AddAttribute(tag, value string) {
	if im.tags == nil {
		im.tags = make(map[string]string)
	}
	im.tags[tag] = value
}

func (im *instanceMemory) Iterator() (int, bool) {
	if im.tags == nil {
		return 0, false
	}
	val, ok := im.tags["loop-index"]
	iterator, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}
	return iterator, ok
}

func (im *instanceMemory) Raise(ctx context.Context, err *derrors.CatchableError) error {
	return im.engine.InstanceRaise(ctx, im, err)
}

func (im *instanceMemory) RetrieveSecret(ctx context.Context, secret string) (string, error) {
	tx, err := im.engine.flow.beginSqlTx(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	secretData, err := tx.DataStore().Secrets().Get(ctx, im.instance.Instance.NamespaceID, secret)
	if err != nil {
		return "", err
	}

	return string(secretData.Data), nil
}

func (im *instanceMemory) SetVariables(ctx context.Context, vars []states.VariableSetter) error {
	tx, err := im.engine.flow.beginSqlTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for idx := range vars {
		v := vars[idx]

		var item *core.RuntimeVariable

		switch v.Scope {
		case util.VarScopeInstance:
			item, err = tx.DataStore().RuntimeVariables().GetByInstanceAndName(ctx, im.instance.Instance.ID, v.Key)
		case util.VarScopeWorkflow:
			item, err = tx.DataStore().RuntimeVariables().GetByWorkflowAndName(ctx, im.instance.Instance.NamespaceID, im.instance.Instance.WorkflowPath, v.Key)
		case util.VarScopeNamespace:
			item, err = tx.DataStore().RuntimeVariables().GetByNamespaceAndName(ctx, im.instance.Instance.NamespaceID, v.Key)
		default:
			return derrors.NewInternalError(errors.New("invalid scope"))
		}

		if err != nil && !errors.Is(err, datastore.ErrNotFound) {
			return err
		}

		d := string(v.Data)

		if len(d) == 0 {
			err = tx.DataStore().RuntimeVariables().Delete(ctx, item.ID)
			if err != nil && !errors.Is(err, datastore.ErrNotFound) {
				return err
			}
			continue
		}

		if !(v.MIMEType == "text/plain; charset=utf-8" || v.MIMEType == "text/plain" || v.MIMEType == "application/octet-stream") && (d == "{}" || d == "[]" || d == "0" || d == `""` || d == "null") {
			if item != nil {
				err = tx.DataStore().RuntimeVariables().Delete(ctx, item.ID)
				if err != nil && !errors.Is(err, datastore.ErrNotFound) {
					return err
				}
			}
		} else {
			newVar := &core.RuntimeVariable{
				Name:        v.Key,
				MimeType:    v.MIMEType,
				Data:        v.Data,
				NamespaceID: im.instance.Instance.NamespaceID,
			}

			switch v.Scope {
			case util.VarScopeInstance:
				newVar.InstanceID = im.instance.Instance.ID
			case util.VarScopeWorkflow:
				newVar.WorkflowPath = im.instance.Instance.WorkflowPath
			}

			_, err = tx.DataStore().RuntimeVariables().Set(ctx, newVar)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (im *instanceMemory) Sleep(ctx context.Context, d time.Duration, x interface{}) error {
	return im.ScheduleRetry(ctx, d, im.logic.GetID(), x)
}

func (im *instanceMemory) GetInstanceData() interface{} {
	return im.data
}

func (im *instanceMemory) GetModel() (*model.Workflow, error) {
	return im.Model()
}

func (im *instanceMemory) GetInstanceID() uuid.UUID {
	return im.instance.Instance.ID
}

func (im *instanceMemory) PrimeDelayedEvent(event cloudevents.Event) {
	im.eventQueue = append(im.eventQueue, event.ID())
}

func (im *instanceMemory) SetMemory(ctx context.Context, x interface{}) error {
	return im.engine.SetMemory(ctx, im, x)
}

func (im *instanceMemory) Deadline(ctx context.Context) time.Time {
	return time.Now().UTC().Add(states.DefaultShortDeadline)
}

func (im *instanceMemory) LivingChildren(ctx context.Context) []*states.ChildInfo {
	return nil
}

func (im *instanceMemory) ScheduleRetry(ctx context.Context, d time.Duration, stateID string, x interface{}) error {
	data, err := json.Marshal(x)
	if err != nil {
		return err
	}

	t := time.Now().UTC().Add(d)

	err = im.engine.scheduleRetry(im.ID().String(), stateID, im.Step(), t, data)
	if err != nil {
		return err
	}

	return nil
}

func (im *instanceMemory) CreateChild(ctx context.Context, args states.CreateChildArgs) (states.Child, error) {
	var ci states.ChildInfo

	if args.Definition.GetType() == model.SubflowFunctionType {
		pi := &enginerefactor.ParentInfo{
			ID:     im.ID(),
			State:  im.logic.GetID(),
			Step:   im.Step(),
			Branch: args.Iterator,
		}
		// TODO: alan
		// caller.CallPath = im.instance.TelemetryInfo.CallPath
		sfim, err := im.engine.subflowInvoke(ctx, pi, im.instance, args.Definition.(*model.SubflowFunctionDefinition).Workflow, args.Input)
		if err != nil {
			return nil, err
		}

		ci.ID = sfim.ID().String()
		ci.Type = "subflow"
		// ci.Attempts: this is ignored here. Must be handled elsewhere.

		return &subflowHandle{
			im:     sfim,
			info:   ci,
			engine: im.engine,
		}, nil
	}

	switch args.Definition.GetType() {
	case model.NamespacedKnativeFunctionType:
	case model.ReusableContainerFunctionType:
	default:
		return nil, derrors.NewInternalError(fmt.Errorf("unsupported function type: %v", args.Definition.GetType()))
	}

	return nil, derrors.NewInternalError(fmt.Errorf("knative function not yet implemented: %v", args.Definition.GetType()))
}

type subflowHandle struct {
	im     *instanceMemory
	info   states.ChildInfo
	engine *engine
}

func (child *subflowHandle) Run(ctx context.Context) {
	child.engine.queue(child.im)
}

func (child *subflowHandle) Info() states.ChildInfo {
	return child.info
}
