package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entvar "github.com/direktiv/direktiv/pkg/flow/ent/varref"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/functions"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	secretsgrpc "github.com/direktiv/direktiv/pkg/secrets/grpc"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TEMPORARY EVERYTHING

func (im *instanceMemory) BroadcastCloudevent(ctx context.Context, event *cloudevents.Event, dd int64) error {
	return im.engine.events.BroadcastCloudevent(ctx, im.in.Edges.Namespace, event, dd)
}

func (im *instanceMemory) GetVariables(ctx context.Context, vars []states.VariableSelector) ([]states.Variable, error) {

	var x = make([]states.Variable, 0)

	for _, selector := range vars {

		var err error
		var ref *ent.VarRef

		scope := selector.Scope
		key := selector.Key

		switch scope {

		case "instance":
			ref, err = im.in.QueryVars().Where(entvar.NameEQ(key), entvar.BehaviourIsNil()).WithVardata().Only(ctx)

		case "thread":
			ref, err = im.in.QueryVars().Where(entvar.NameEQ(key), entvar.BehaviourEQ("thread")).WithVardata().Only(ctx)

		case "workflow":

			wf, err := im.engine.InstanceWorkflow(ctx, im)
			if err != nil {
				return nil, derrors.NewInternalError(err)
			}

			// NOTE: this hack seems to be necessary for some reason...
			wf, err = im.engine.db.Workflow.Get(ctx, wf.ID)
			if err != nil {
				return nil, derrors.NewInternalError(err)
			}

			ref, err = wf.QueryVars().Where(entvar.NameEQ(key)).WithVardata().Only(ctx)

		case "namespace":

			ns, err := im.engine.InstanceNamespace(ctx, im)
			if err != nil {
				return nil, derrors.NewInternalError(err)
			}

			// NOTE: this hack seems to be necessary for some reason...
			ns, err = im.engine.db.Namespace.Get(ctx, ns.ID)
			if err != nil {
				return nil, derrors.NewInternalError(err)
			}

			ref, err = ns.QueryVars().Where(entvar.NameEQ(key)).WithVardata().Only(ctx)

		default:
			return nil, derrors.NewInternalError(errors.New("invalid scope"))

		}

		var data []byte

		if err != nil {

			if !derrors.IsNotFound(err) {
				return nil, derrors.NewInternalError(err)
			}

			data = make([]byte, 0)

		} else if ref == nil {

			data = make([]byte, 0)

		} else {

			if ref.Edges.Vardata == nil {

				err = &derrors.NotFoundError{
					Label: fmt.Sprintf("variable data not found"),
				}

				return nil, err

			}

			data = ref.Edges.Vardata.Data

		}

		x = append(x, states.Variable{
			Scope: scope,
			Key:   key,
			Data:  data,
		})

	}

	return x, nil

}

func (im *instanceMemory) ListenForEvents(ctx context.Context, events []*model.ConsumeEventDefinition, all bool) error {

	err := im.engine.events.deleteInstanceEventListeners(ctx, im.in)
	if err != nil {
		return err
	}

	err = im.engine.events.listenForEvents(ctx, im, events, all)
	if err != nil {
		return err
	}

	return nil

}

func (im *instanceMemory) Log(ctx context.Context, a string, x ...interface{}) {

	im.engine.logToInstance(ctx, time.Now(), im.in, a, x...)

}

func (im *instanceMemory) Raise(ctx context.Context, err *derrors.CatchableError) error {

	return im.engine.InstanceRaise(ctx, im, err)

}

func (im *instanceMemory) RetrieveSecret(ctx context.Context, secret string) (string, error) {

	var resp *secretsgrpc.SecretsRetrieveResponse

	ns := im.in.Edges.Namespace.ID.String()

	resp, err := im.engine.secrets.client.RetrieveSecret(ctx, &secretsgrpc.SecretsRetrieveRequest{
		Namespace: &ns,
		Name:      &secret,
	})
	if err != nil {
		s := status.Convert(err)
		if s.Code() == codes.NotFound {
			return "", derrors.NewUncatchableError("direktiv.secrets.notFound", "secret '%s' not found", secret)
		}
		return "", derrors.NewInternalError(err)
	}

	return string(resp.GetData()), nil

}

func (im *instanceMemory) SetVariables(ctx context.Context, vars []states.VariableSetter) error {

	tx, err := im.engine.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	vdatac := tx.VarData
	vrefc := tx.VarRef

	for idx := range vars {

		v := vars[idx]

		var q varQuerier

		var thread bool

		switch v.Scope {

		case "":

			fallthrough

		case "instance":
			q, err = tx.Instance.Get(ctx, im.in.ID)
			if err != nil {
				return err
			}

		case "thread":
			q, err = tx.Instance.Get(ctx, im.in.ID)
			if err != nil {
				return err
			}
			thread = true

		case "workflow":
			wf, err := im.engine.InstanceWorkflow(ctx, im)
			if err != nil {
				return err
			}

			q, err = tx.Workflow.Get(ctx, wf.ID)
			if err != nil {
				return err
			}

		case "namespace":
			ns, err := im.engine.InstanceNamespace(ctx, im)
			if err != nil {
				return err
			}

			q, err = tx.Namespace.Get(ctx, ns.ID)
			if err != nil {
				return err
			}

		default:
			return derrors.NewInternalError(errors.New("invalid scope"))
		}

		// if statements have to be same order

		d := string(v.Data)
		if (v.MIMEType == "text/plain; charset=utf-8" || v.MIMEType == "text/plain" || v.MIMEType == "application/octet-stream") && len(d) == 0 {
			_, _, err = im.engine.flow.DeleteVariable(ctx, vrefc, vdatac, q, v.Key, v.Data, v.MIMEType, thread)
			if err != nil {
				return err
			}
			continue
		}

		if d == "{}" || d == "[]" || d == "0" || d == "" {
			_, _, err = im.engine.flow.DeleteVariable(ctx, vrefc, vdatac, q, v.Key, v.Data, v.MIMEType, thread)
			if err != nil {
				return err
			}
			continue
		}

		if len(d) > 0 {
			_, _, err = im.engine.flow.SetVariable(ctx, vrefc, vdatac, q, v.Key, v.Data, v.MIMEType, thread)
			if err != nil {
				return err
			}
		}

	}

	err = tx.Commit()
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
	return im.in.ID
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

func (im *instanceMemory) LivingChildren(ctx context.Context) []states.ChildInfo {
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
		caller.InstanceID = im.ID().String()
		caller.State = im.logic.GetID()
		caller.Step = im.Step()
		caller.As = im.in.As

		sfim, err := im.engine.subflowInvoke(ctx, caller, im.in.Edges.Namespace, args.Definition.(*model.SubflowFunctionDefinition).Workflow, args.Input)
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
	case model.GlobalKnativeFunctionType:
	case model.NamespacedKnativeFunctionType:
	case model.ReusableContainerFunctionType:
	default:
		return nil, derrors.NewInternalError(fmt.Errorf("unsupported function type: %v", args.Definition.GetType()))
	}

	uid := uuid.New()

	ar, err := im.engine.newIsolateRequest(ctx, im, im.logic.GetID(), args.Timeout, args.Definition, args.Input, uid, args.Async, args.Files)
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
	uid uuid.UUID, async bool, files []model.FunctionFileDefinition) (*functionRequest, error) {

	wf, err := engine.InstanceWorkflow(ctx, im)
	if err != nil {
		return nil, err
	}

	ar := new(functionRequest)
	ar.ActionID = uid.String()
	// ar.Workflow.Name = wli.wf.Name
	ar.Workflow.WorkflowID = wf.ID.String()
	ar.Workflow.Timeout = timeout
	ar.Workflow.Revision = im.in.Edges.Revision.Hash
	ar.Workflow.NamespaceName = im.in.Edges.Namespace.Name
	ar.Workflow.Path = im.in.As

	if !async {
		ar.Workflow.InstanceID = im.ID().String()
		ar.Workflow.NamespaceID = im.in.Edges.Namespace.ID.String()
		ar.Workflow.State = stateId
		ar.Workflow.Step = im.Step()
	}

	// TODO: timeout
	fnt := fn.GetType()
	ar.Container.Type = fnt
	ar.Container.Data = inputData

	wfID := im.in.Edges.Workflow.ID.String()
	revID := im.in.Edges.Revision.Hash
	nsID := im.in.Edges.Namespace.ID.String()

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
		if err != nil {
			panic(err)
		}
	case model.NamespacedKnativeFunctionType:
		con := fn.(*model.NamespacedFunctionDefinition)
		ar.Container.Files = files
		ar.Container.ID = con.ID
		ar.Container.Service, _, _ = functions.GenerateServiceName(&igrpc.BaseInfo{
			Name:          &con.KnativeService,
			Namespace:     &nsID,
			NamespaceName: &ar.Workflow.NamespaceName,
		})
	case model.GlobalKnativeFunctionType:
		con := fn.(*model.GlobalFunctionDefinition)
		ar.Container.Files = files
		ar.Container.ID = con.ID
		ar.Container.Service, _, _ = functions.GenerateServiceName(&igrpc.BaseInfo{
			Name: &con.KnativeService,
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
