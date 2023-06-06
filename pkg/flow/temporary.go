package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	log "github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/functions"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
)

// TEMPORARY EVERYTHING

func (im *instanceMemory) BroadcastCloudevent(ctx context.Context, event *cloudevents.Event, dd int64) error {
	return im.engine.events.BroadcastCloudevent(ctx, im.cached.Namespace, event, dd)
}

func (im *instanceMemory) GetVariables(ctx context.Context, vars []states.VariableSelector) ([]states.Variable, error) {
	x := make([]states.Variable, 0)

	fStore, store, _, rollback, err := im.engine.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	for _, selector := range vars {
		if selector.Scope == util.VarScopeInstance || selector.Scope == util.VarScopeWorkflow || selector.Scope == util.VarScopeNamespace {
			referenceID := im.cached.Namespace.ID
			if selector.Scope == util.VarScopeInstance {
				referenceID = im.cached.Instance.ID
			}
			if selector.Scope == util.VarScopeWorkflow {
				referenceID = im.cached.File.ID
			}

			item, err := store.RuntimeVariables().GetByReferenceAndName(ctx, referenceID, selector.Key)
			if errors.Is(err, datastore.ErrNotFound) {
				x = append(x, states.Variable{
					Scope: selector.Scope,
					Key:   selector.Key,
					Data:  []byte{},
				})
			} else if err != nil {
				return nil, derrors.NewInternalError(err)
			} else {
				data, err := store.RuntimeVariables().LoadData(ctx, item.ID)
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
			file, err := fStore.ForRootID(im.cached.Namespace.ID).GetFile(ctx, selector.Key)
			if errors.Is(err, filestore.ErrNotFound) {
				x = append(x, states.Variable{
					Scope: selector.Scope,
					Key:   selector.Key,
					Data:  []byte{},
				})
			} else if err != nil {
				return nil, err
			} else {
				// TODO: alan, maybe need to enhance the GetData function to also return us some information like mime type, checksum, and size
				if file.Typ != filestore.FileTypeFile {
					return nil, model.ErrVarNotFile
				}
				rc, err := fStore.ForFile(file).GetData(ctx)
				if err != nil {
					return nil, err
				}
				defer func() { _ = rc.Close() }()
				data, err := ioutil.ReadAll(rc)
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
	err := im.engine.events.deleteInstanceEventListeners(ctx, im.cached)
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
	_, store, _, rollback, err := im.engine.flow.beginSqlTx(ctx)
	if err != nil {
		return "", err
	}
	defer rollback()

	secretData, err := store.Secrets().Get(ctx, im.cached.Namespace.ID, secret)
	if err != nil {
		return "", err
	}

	return string(secretData.Data), nil
}

func (im *instanceMemory) SetVariables(ctx context.Context, vars []states.VariableSetter) error {
	_, store, commit, rollback, err := im.engine.flow.beginSqlTx(ctx)
	if err != nil {
		return err
	}
	defer rollback()

	for idx := range vars {
		v := vars[idx]

		var referenceID uuid.UUID
		switch v.Scope {
		case "instance":
			referenceID = im.cached.Instance.ID
		case "workflow":
			referenceID = im.cached.File.ID
		case "namespace":
			referenceID = im.cached.Namespace.ID
		default:
			return derrors.NewInternalError(errors.New("invalid scope"))
		}

		item, err := store.RuntimeVariables().GetByReferenceAndName(ctx, referenceID, v.Key)
		if err != nil && !errors.Is(err, datastore.ErrNotFound) {
			return err
		}

		d := string(v.Data)

		if len(d) == 0 {
			err = store.RuntimeVariables().Delete(ctx, item.ID)
			if err != nil && !errors.Is(err, datastore.ErrNotFound) {
				return err
			}
			continue
		}

		if !(v.MIMEType == "text/plain; charset=utf-8" || v.MIMEType == "text/plain" || v.MIMEType == "application/octet-stream") && (d == "{}" || d == "[]" || d == "0" || d == `""` || d == "null") {
			err = store.RuntimeVariables().Delete(ctx, item.ID)
			if err != nil && !errors.Is(err, datastore.ErrNotFound) {
				return err
			}
		} else {
			newVar := &core.RuntimeVariable{
				Name:     v.Key,
				MimeType: v.MIMEType,
				Data:     v.Data,
			}

			switch v.Scope {
			case "instance":
				newVar.InstanceID = referenceID
			case "workflow":
				newVar.WorkflowID = referenceID
			case "namespace":
				newVar.NamespaceID = referenceID
			}

			_, err = store.RuntimeVariables().Set(ctx, newVar)
			if err != nil {
				return err
			}
		}
	}
	err = commit(ctx)
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
	return im.cached.Instance.ID
}

func (im *instanceMemory) PrimeDelayedEvent(event cloudevents.Event) {
	im.eventQueue = append(im.eventQueue, event.ID())
}

func (im *instanceMemory) SetMemory(ctx context.Context, x interface{}) error {
	return im.engine.SetMemory(ctx, im, x)
}

func (im *instanceMemory) Deadline(ctx context.Context) time.Time {
	return time.Now().Add(states.DefaultShortDeadline)
}

func (im *instanceMemory) LivingChildren(ctx context.Context) []*states.ChildInfo {
	return nil
}

func (im *instanceMemory) ScheduleRetry(ctx context.Context, d time.Duration, stateID string, x interface{}) error {
	data, err := json.Marshal(x)
	if err != nil {
		return err
	}

	t := time.Now().Add(d)

	err = im.engine.scheduleRetry(im.ID().String(), stateID, im.Step(), t, data)
	if err != nil {
		return err
	}

	return nil
}

func (im *instanceMemory) CreateChild(ctx context.Context, args states.CreateChildArgs) (states.Child, error) {
	var ci states.ChildInfo

	if args.Definition.GetType() == model.SubflowFunctionType {
		caller := new(subflowCaller)
		caller.InstanceID = im.ID()
		caller.State = im.logic.GetID()
		caller.Step = im.Step()
		caller.As = im.cached.Instance.As
		caller.CallPath = im.cached.Instance.CallPath
		caller.CallerState = im.GetState()
		caller.Iterator = fmt.Sprintf("%d", args.Iterator)
		sfim, err := im.engine.subflowInvoke(ctx, caller, im.cached, args.Definition.(*model.SubflowFunctionDefinition).Workflow, args.Input)
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

	uid := uuid.New()

	ar, err := im.engine.newIsolateRequest(ctx, im, im.logic.GetID(), args.Timeout, args.Definition, args.Input, uid, args.Async, args.Files, args.Iterator)
	if err != nil {
		return nil, err
	}

	ci.ID = ar.ActionID
	ci.ServiceName = ar.Container.Service
	ci.Type = "isolate"

	return &knativeHandle{
		im:     im,
		info:   ci,
		engine: im.engine,
		ar:     ar,
	}, nil
}

func (engine *engine) newIsolateRequest(ctx context.Context, im *instanceMemory, stateId string, timeout int,
	fn model.FunctionDefinition, inputData []byte,
	uid uuid.UUID, async bool, files []model.FunctionFileDefinition, iterator int,
) (*functionRequest, error) {
	ar := new(functionRequest)
	ar.ActionID = uid.String()
	ar.Workflow.WorkflowID = im.cached.File.ID.String()
	ar.Workflow.Timeout = timeout
	ar.Workflow.Revision = im.cached.Revision.Checksum
	ar.Workflow.NamespaceName = im.cached.Namespace.Name
	ar.Workflow.Path = im.cached.Instance.As
	ar.Iterator = iterator
	if !async {
		ar.Workflow.InstanceID = im.ID().String()
		ar.Workflow.NamespaceID = im.cached.Namespace.ID.String()
		ar.Workflow.State = stateId
		ar.Workflow.Step = im.Step()
	}

	fnt := fn.GetType()
	ar.Container.Type = fnt
	ar.Container.Data = inputData

	wfID := im.cached.File.ID.String()
	revID := im.cached.Revision.Checksum
	nsID := im.cached.Namespace.ID.String()

	switch fnt {
	case model.ReusableContainerFunctionType:

		con := fn.(*model.ReusableFunctionDefinition)

		scale := int32(0)
		size := int32(con.Size)

		ar.Container.Image = con.Image
		ar.Container.Cmd = con.Cmd
		ar.Container.Size = con.Size
		ar.Container.Scale = int(scale)
		ar.Container.Files = files
		ar.Container.ID = con.ID
		ar.Container.Service, _, _ = functions.GenerateServiceName(&igrpc.BaseInfo{
			Name:          &con.ID,
			Workflow:      &wfID,
			Revision:      &revID,
			Namespace:     &nsID,
			NamespaceName: &ar.Workflow.NamespaceName,
			Image:         &con.Image,
			Cmd:           &con.Cmd,
			MinScale:      &scale,
			Size:          &size,
		})
	case model.NamespacedKnativeFunctionType:
		con := fn.(*model.NamespacedFunctionDefinition)
		ar.Container.Files = files
		ar.Container.ID = con.ID
		ar.Container.Service, _, _ = functions.GenerateServiceName(&igrpc.BaseInfo{
			Name:          &con.KnativeService,
			Namespace:     &nsID,
			NamespaceName: &ar.Workflow.NamespaceName,
		})
	default:
		return nil, fmt.Errorf("unexpected function type: %v", fn)
	}

	// check for duplicate file names
	m := make(map[string]*model.FunctionFileDefinition)
	for i := range ar.Container.Files {
		f := &ar.Container.Files[i]
		k := f.As
		if k == "" {
			k = f.Key
		}
		if _, exists := m[k]; exists {
			return nil, fmt.Errorf("multiple files with same name: %s", k)
		}
		m[k] = f
	}

	return ar, nil
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

type knativeHandle struct {
	im     *instanceMemory
	info   states.ChildInfo
	engine *engine
	ar     *functionRequest
}

func (child *knativeHandle) Run(ctx context.Context) {
	go func(ctx context.Context, im *instanceMemory, ar *functionRequest) {
		err := child.engine.doActionRequest(ctx, ar)
		if err != nil {
			return
		}
	}(ctx, child.im, child.ar)
}

func (child *knativeHandle) Info() states.ChildInfo {
	return child.info
}
