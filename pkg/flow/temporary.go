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
	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	log "github.com/direktiv/direktiv/pkg/flow/internallogger"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/functions"
	"github.com/direktiv/direktiv/pkg/functions/grpc"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/api"
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
	ar.Workflow.Timeout = timeout
	ar.Workflow.Revision = im.instance.Instance.RevisionID.String()
	ar.Workflow.NamespaceName = im.instance.TelemetryInfo.NamespaceName
	ar.Workflow.Path = im.instance.Instance.WorkflowPath
	ar.Iterator = iterator
	if !async {
		ar.Workflow.InstanceID = im.ID().String()
		ar.Workflow.NamespaceID = im.instance.Instance.NamespaceID.String()
		ar.Workflow.State = stateId
		ar.Workflow.Step = im.Step()
	}

	fnt := fn.GetType()
	ar.Container.Type = fnt
	ar.Container.Data = inputData

	wf := bytedata.ShortChecksum(im.instance.Instance.WorkflowPath)
	revID := im.instance.Instance.RevisionID.String()
	nsID := im.instance.Instance.NamespaceID.String()

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
		ar.Container.Service, _, _ = functions.GenerateServiceName(&igrpc.FunctionsBaseInfo{
			Name:          &con.ID,
			Workflow:      &wf,
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
		ar.Container.Service, _, _ = functions.GenerateServiceName(&igrpc.FunctionsBaseInfo{
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

func (server *server) setNamespaceServices(nsID uuid.UUID, services []*api.Service) error {
	ctx := context.Background()

	tx, err := server.flow.beginSqlTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByID(ctx, nsID)
	if err != nil {
		return err
	}

	tx.Rollback()

	// TODO: more error tolerance?

	// NOTE: until the refactor reaches the functions logic I hope we can get away with the simplest solution. Purge all existing and replace, even if unchanged.
	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderScope] = functions.PrefixNamespace
	annotations[functions.ServiceHeaderNamespaceName] = ns.Name

	_, err = server.functionsClient.DeleteFunctions(ctx, &igrpc.FunctionsListFunctionsRequest{
		Annotations: annotations,
	})
	if err != nil {
		return err
	}

	for idx := range services {
		req := new(igrpc.FunctionsCreateFunctionRequest)
		namespace := ns.ID.String()

		var size int32

		switch services[idx].Size {
		case "large":
			size = 2
		case "medium":
			size = 1
		}

		req.Info = &grpc.FunctionsBaseInfo{
			Namespace:     &namespace,
			NamespaceName: &ns.Name,
			Name:          &services[idx].Name,
			Image:         &services[idx].Image,
			Cmd:           &services[idx].Cmd,
			Size:          &size,
			MinScale:      &services[idx].Scale,
			// Path:          &cr.WorkflowPath,
			// Envs:          cr.Envs,
		}

		_, err = server.functionsClient.CreateFunction(ctx, req)
		if err != nil {
			return err
		}
	}

	return nil
}
