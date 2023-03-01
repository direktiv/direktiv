package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	entvar "github.com/direktiv/direktiv/pkg/flow/ent/varref"
	entwf "github.com/direktiv/direktiv/pkg/flow/ent/workflow"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/states"
	"github.com/direktiv/direktiv/pkg/functions"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	secretsgrpc "github.com/direktiv/direktiv/pkg/secrets/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TEMPORARY EVERYTHING

func (im *instanceMemory) BroadcastCloudevent(ctx context.Context, event *cloudevents.Event, dd int64) error {
	return im.engine.events.BroadcastCloudevent(ctx, im.cached, event, dd)
}

func (im *instanceMemory) GetVariables(ctx context.Context, vars []states.VariableSelector) ([]states.Variable, error) {
	x := make([]states.Variable, 0)

	clients := im.engine.edb.Clients(ctx)

	for _, selector := range vars {

		var err error
		var ref *ent.VarRef

		scope := selector.Scope
		key := selector.Key

		switch scope {

		case util.VarScopeInstance:
			ref, err = clients.VarRef.Query().Where(entvar.HasInstanceWith(entinst.ID(im.cached.Instance.ID))).Where(entvar.NameEQ(key), entvar.BehaviourIsNil()).WithVardata().Only(ctx)

		case util.VarScopeThread:
			ref, err = clients.VarRef.Query().Where(entvar.HasInstanceWith(entinst.ID(im.cached.Instance.ID))).Where(entvar.NameEQ(key), entvar.BehaviourEQ("thread")).WithVardata().Only(ctx)

		case util.VarScopeWorkflow:

			// // NOTE: this hack seems to be necessary for some reason...
			// wf, err = im.engine.db.Workflow.Get(ctx, wf.ID)
			// if err != nil {
			// 	return nil, derrors.NewInternalError(err)
			// }

			ref, err = clients.VarRef.Query().Where(entvar.HasWorkflowWith(entwf.ID(im.cached.Workflow.ID))).Where(entvar.NameEQ(key)).WithVardata().Only(ctx)

		case util.VarScopeNamespace:

			// // NOTE: this hack seems to be necessary for some reason...
			// ns, err = im.engine.db.Namespace.Get(ctx, ns.ID)
			// if err != nil {
			// 	return nil, derrors.NewInternalError(err)
			// }

			ref, err = clients.VarRef.Query().Where(entvar.HasNamespaceWith(entns.ID(im.cached.Namespace.ID))).Where(entvar.NameEQ(key)).WithVardata().Only(ctx)

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
					Label: "variable data not found",
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

func (im *instanceMemory) Log(ctx context.Context, a string, x ...interface{}) {
	im.engine.logToInstance(ctx, time.Now(), im.cached, a, x...)
}

func (im *instanceMemory) Raise(ctx context.Context, err *derrors.CatchableError) error {
	return im.engine.InstanceRaise(ctx, im, err)
}

func (im *instanceMemory) RetrieveSecret(ctx context.Context, secret string) (string, error) {
	var resp *secretsgrpc.SecretsRetrieveResponse

	ns := im.cached.Namespace.ID.String()

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
	tctx, tx, err := im.engine.database.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	clients := im.engine.edb.Clients(tctx)

	for idx := range vars {

		v := vars[idx]

		var q varQuerier

		var thread bool

		switch v.Scope {

		case "":

			fallthrough

		case "instance":

			q = &entInstanceVarQuerier{
				clients: clients,
				cached:  im.cached,
			}

		case "thread":

			q = &entInstanceVarQuerier{
				clients: clients,
				cached:  im.cached,
			}

			thread = true

		case "workflow":

			q = &entWorkflowVarQuerier{
				clients: clients,
				cached:  im.cached,
			}

		case "namespace":

			q = &entNamespaceVarQuerier{
				clients: clients,
				cached:  im.cached,
			}

		default:
			return derrors.NewInternalError(errors.New("invalid scope"))
		}

		// if statements have to be same order

		d := string(v.Data)

		if len(d) == 0 {
			_, _, err = im.engine.flow.DeleteVariable(tctx, q, v.Key, v.Data, v.MIMEType, thread)
			if err != nil {
				return err
			}
			continue

		}

		if !(v.MIMEType == "text/plain; charset=utf-8" || v.MIMEType == "text/plain" || v.MIMEType == "application/octet-stream") && (d == "{}" || d == "[]" || d == "0" || d == `""` || d == "null") {
			_, _, err = im.engine.flow.DeleteVariable(tctx, q, v.Key, v.Data, v.MIMEType, thread)
			if err != nil {
				return err
			}
			continue

		} else {
			_, _, err = im.engine.flow.SetVariable(tctx, q, v.Key, v.Data, v.MIMEType, thread)
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
		caller.InstanceID = im.ID()
		caller.State = im.logic.GetID()
		caller.Step = im.Step()
		caller.As = im.cached.Instance.As

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
	uid uuid.UUID, async bool, files []model.FunctionFileDefinition,
) (*functionRequest, error) {
	ar := new(functionRequest)
	ar.ActionID = uid.String()
	ar.Workflow.WorkflowID = im.cached.Workflow.ID.String()
	ar.Workflow.Timeout = timeout
	ar.Workflow.Revision = im.cached.Revision.Hash
	ar.Workflow.NamespaceName = im.cached.Namespace.Name
	ar.Workflow.Path = im.cached.Instance.As

	if !async {
		ar.Workflow.InstanceID = im.ID().String()
		ar.Workflow.NamespaceID = im.cached.Namespace.ID.String()
		ar.Workflow.State = stateId
		ar.Workflow.Step = im.Step()
	}

	fnt := fn.GetType()
	ar.Container.Type = fnt
	ar.Container.Data = inputData

	wfID := im.cached.Workflow.ID.String()
	revID := im.cached.Revision.Hash
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
