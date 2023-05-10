package flow

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (srv *server) getInstance(ctx context.Context, namespace, instanceID string) (*database.CacheData, error) {
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, err
	}

	cached := new(database.CacheData)

	err = srv.database.Instance(ctx, cached, id)
	if err != nil {
		return nil, err
	}

	fStore, _, _, rollback, err := srv.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, revision, err := fStore.GetRevision(ctx, cached.Instance.Revision)
	if err != nil {
		if !errors.Is(err, filestore.ErrNotFound) {
			return nil, err
		}
		file, err = fStore.GetFile(ctx, cached.Instance.Workflow)
		if err != nil {
			return nil, err
		}
		err = nil
	}

	cached.File = file
	cached.Revision = revision

	if namespace != cached.Namespace.Name {
		return nil, os.ErrNotExist
	}

	return cached, nil
}

func (internal *internal) getInstance(ctx context.Context, instanceID string) (*database.CacheData, error) {
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, err
	}

	cached := new(database.CacheData)

	err = internal.database.Instance(ctx, cached, id)
	if err != nil {
		return nil, err
	}

	fStore, _, _, rollback, err := internal.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, revision, err := fStore.GetRevision(ctx, cached.Instance.Revision)
	if err != nil {
		return nil, err
	}

	cached.File = file
	cached.Revision = revision

	return cached, nil
}

func (srv *server) getInstanceRuntime(ctx context.Context, namespace, instanceID string) (*database.CacheData, *database.InstanceRuntime, error) {
	cached, err := srv.getInstance(ctx, namespace, instanceID)
	if err != nil {
		return nil, nil, err
	}

	rt, err := srv.database.InstanceRuntime(ctx, cached.Instance.Runtime)
	if err != nil {
		return nil, nil, err
	}

	return cached, rt, nil
}

func (flow *flow) InstanceInput(ctx context.Context, req *grpc.InstanceInputRequest) (*grpc.InstanceInputResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, rt, err := flow.getInstanceRuntime(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceInputResponse

	err = bytedata.ConvertDataForOutput(cached.Instance, &resp.Instance)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(rt.Input, &m)
	if err != nil {
		return nil, err
	}
	delete(m, "private")
	input := bytedata.Marshal(m)

	resp.Data = []byte(input)
	resp.Namespace = cached.Namespace.Name

	return &resp, nil
}

func (flow *flow) InstanceOutput(ctx context.Context, req *grpc.InstanceOutputRequest) (*grpc.InstanceOutputResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, rt, err := flow.getInstanceRuntime(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceOutputResponse

	err = bytedata.ConvertDataForOutput(cached.Instance, &resp.Instance)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal([]byte(rt.Output), &m)
	if err != nil {
		return nil, err
	}
	delete(m, "private")
	output := bytedata.Marshal(m)

	resp.Data = []byte(output)
	resp.Namespace = cached.Namespace.Name

	return &resp, nil
}

func (flow *flow) InstanceMetadata(ctx context.Context, req *grpc.InstanceMetadataRequest) (*grpc.InstanceMetadataResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, rt, err := flow.getInstanceRuntime(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceMetadataResponse

	err = bytedata.ConvertDataForOutput(cached.Instance, &resp.Instance)
	if err != nil {
		return nil, err
	}

	resp.Data = []byte(rt.Metadata)
	resp.Namespace = cached.Namespace.Name

	return &resp, nil
}

func (flow *flow) Instances(ctx context.Context, req *grpc.InstancesRequest) (*grpc.InstancesResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	query := clients.Instance.Query().Where(entinst.HasNamespaceWith(entns.ID(cached.Namespace.ID)))

	results, pi, err := paginate[*ent.InstanceQuery, *ent.Instance](ctx, req.Pagination, query, instancesOrderings, instancesFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.InstancesResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Instances = new(grpc.Instances)
	resp.Instances.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Instances.Results)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (flow *flow) InstancesStream(req *grpc.InstancesRequest, srv grpc.Flow_InstancesStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeInstances(cached.Namespace)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(ctx)

	query := clients.Instance.Query().Where(entinst.HasNamespaceWith(entns.ID(cached.Namespace.ID)))

	results, pi, err := paginate[*ent.InstanceQuery, *ent.Instance](ctx, req.Pagination, query, instancesOrderings, instancesFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.InstancesResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Instances = new(grpc.Instances)
	resp.Instances.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Instances.Results)
	if err != nil {
		return err
	}

	nhash = bytedata.Checksum(resp)
	if nhash != phash {
		err = srv.Send(resp)
		if err != nil {
			return err
		}
	}
	phash = nhash

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend
}

func (flow *flow) Instance(ctx context.Context, req *grpc.InstanceRequest) (*grpc.InstanceResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, rt, err := flow.getInstanceRuntime(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceResponse

	err = bytedata.ConvertDataForOutput(cached.Instance, &resp.Instance)
	if err != nil {
		return nil, err
	}

	resp.Flow = rt.Flow
	if rt.Caller != uuid.Nil {
		resp.InvokedBy = rt.Caller.String()
	}

	resp.Namespace = cached.Namespace.Name

	rwf := new(grpc.InstanceWorkflow)
	if cached.File != nil {
		rwf.Name = cached.File.Name()
		rwf.Parent = strings.TrimPrefix(cached.File.Dir(), "/") // TODO: get rid of the trim?
		rwf.Path = strings.TrimPrefix(cached.File.Path, "/")    // TODO: get rid of the trim?
	}
	if cached.Revision != nil {
		rwf.Revision = cached.Revision.ID.String()
	}
	resp.Workflow = rwf

	return &resp, nil
}

func (flow *flow) InstanceStream(req *grpc.InstanceRequest, srv grpc.Flow_InstanceStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	var sub *pubsub.Subscription

resend:

	cached, rt, err := flow.getInstanceRuntime(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return err
	}

	if sub == nil {
		sub = flow.pubsub.SubscribeInstance(cached)
		defer flow.cleanup(sub.Close)
		goto resend
	}

	resp := new(grpc.InstanceResponse)

	err = bytedata.ConvertDataForOutput(cached.Instance, &resp.Instance)
	if err != nil {
		return err
	}

	resp.Flow = rt.Flow
	resp.InvokedBy = rt.Caller.String()

	resp.Namespace = cached.Namespace.Name

	rwf := new(grpc.InstanceWorkflow)
	if cached.File != nil {
		rwf.Name = cached.File.Name()
		rwf.Parent = strings.TrimPrefix(cached.File.Dir(), "/") // TODO: get rid of the trim?
		rwf.Path = strings.TrimPrefix(cached.File.Path, "/")    // TODO: get rid of the trim?
	}
	if cached.Revision != nil {
		rwf.Revision = cached.Revision.ID.String()
	}
	resp.Workflow = rwf

	nhash = bytedata.Checksum(resp)
	if nhash != phash {
		err = srv.Send(resp)
		if err != nil {
			return err
		}
	}
	phash = nhash

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend
}

func (flow *flow) StartWorkflow(ctx context.Context, req *grpc.StartWorkflowRequest) (*grpc.StartWorkflowResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	args := new(newInstanceArgs)
	args.Namespace = req.GetNamespace()
	args.Path = req.GetPath()
	args.Ref = req.GetRef()
	args.Input = req.GetInput()
	args.Caller = apiCaller

	im, err := flow.engine.NewInstance(ctx, args)
	if err != nil {
		flow.sugar.Debugf("Error returned to gRPC request %s: %v", this(), err)
		flow.logger.Errorf(ctx, flow.ID, flow.GetAttributes(), "Failed starting a Workflow")
		return nil, err
	}

	if !req.GetHold() {
		flow.engine.queue(im)
	}

	var resp grpc.StartWorkflowResponse

	resp.Namespace = req.GetNamespace()
	resp.Instance = im.ID().String()

	return &resp, nil
}

func (flow *flow) ReleaseInstance(ctx context.Context, req *grpc.ReleaseInstanceRequest) (*grpc.ReleaseInstanceResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	im, err := flow.engine.getInstanceMemory(ctx, req.GetInstance())
	if err != nil {
		return nil, err
	}

	if im.cached.Namespace.Name != req.GetNamespace() {
		return nil, errors.New("instance not found")
	}

	if im.cached.Instance.Status != util.InstanceStatusPending {
		return nil, errors.New("instance already released")
	}

	flow.engine.queue(im)

	var resp grpc.ReleaseInstanceResponse

	resp.Namespace = req.GetNamespace()
	resp.Instance = im.ID().String()

	return &resp, nil
}

var instancesOrderings = []*orderingInfo{
	{
		db:           entinst.FieldCreatedAt,
		req:          "CREATED",
		defaultOrder: ent.Desc,
	},
	{
		db:           entinst.FieldID,
		req:          "ID",
		defaultOrder: ent.Desc,
	},
}

var instancesFilters = map[*filteringInfo]func(query *ent.InstanceQuery, v string) (*ent.InstanceQuery, error){
	{
		field: "AS",
		ftype: "WORKFLOW",
	}: func(query *ent.InstanceQuery, v string) (*ent.InstanceQuery, error) {
		return query.Where(entinst.AsHasPrefix(v)), nil
	},
	{
		field: "AS",
		ftype: "CONTAINS",
	}: func(query *ent.InstanceQuery, v string) (*ent.InstanceQuery, error) {
		return query.Where(entinst.AsContains(v)), nil
	},
	{
		field: "CREATED",
		ftype: "BEFORE",
	}: func(query *ent.InstanceQuery, v string) (*ent.InstanceQuery, error) {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, err
		}
		return query.Where(entinst.CreatedAtGTE(t)), nil
	},
	{
		field: "CREATED",
		ftype: "AFTER",
	}: func(query *ent.InstanceQuery, v string) (*ent.InstanceQuery, error) {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, err
		}
		return query.Where(entinst.CreatedAtLTE(t)), nil
	},
	{
		field: "STATUS",
		ftype: "MATCH",
	}: func(query *ent.InstanceQuery, v string) (*ent.InstanceQuery, error) {
		return query.Where(entinst.StatusEQ(v)), nil
	},
	{
		field: "STATUS",
		ftype: "CONTAINS",
	}: func(query *ent.InstanceQuery, v string) (*ent.InstanceQuery, error) {
		return query.Where(entinst.StatusContains(v)), nil
	},
	{
		field: "TRIGGER",
		ftype: "MATCH",
	}: func(query *ent.InstanceQuery, v string) (*ent.InstanceQuery, error) {
		return query.Where(entinst.InvokerEQ(v)), nil
	},
	{
		field: "TRIGGER",
		ftype: "CONTAINS",
	}: func(query *ent.InstanceQuery, v string) (*ent.InstanceQuery, error) {
		return query.Where(entinst.InvokerContains(v)), nil
	},
}

func (flow *flow) CancelInstance(ctx context.Context, req *grpc.CancelInstanceRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		flow.logger.Errorf(ctx, flow.ID, flow.GetAttributes(), "Failed to resolve instance %s", req.GetInstance())
		return nil, err
	}

	flow.engine.cancelInstance(cached.Instance.ID.String(), "direktiv.cancels.api", "cancelled by api request", false)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) AwaitWorkflow(req *grpc.AwaitWorkflowRequest, srv grpc.Flow_AwaitWorkflowServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	args := new(newInstanceArgs)
	args.Namespace = req.GetNamespace()
	args.Path = req.GetPath()
	args.Ref = req.GetRef()
	args.Input = req.GetInput()
	args.Caller = apiCaller

	im, err := flow.engine.NewInstance(ctx, args)
	if err != nil {
		flow.logger.Errorf(ctx, flow.ID, flow.GetAttributes(), "Failed to create a instance")
		flow.sugar.Debugf("Error returned to gRPC request %s: %v", this(), err)
		return err
	}

	sub := flow.pubsub.SubscribeInstance(im.cached)
	defer flow.cleanup(sub.Close)

	flow.engine.queue(im)

	var cached *database.CacheData

resend:

	if cached == nil {
		cached, err = flow.getInstance(ctx, req.GetNamespace(), im.cached.Instance.ID.String())
		if err != nil {
			return err
		}
	}

	err = flow.database.Instance(ctx, cached, cached.Instance.ID)
	if err != nil {
		return err
	}

	fStore, _, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return err
	}
	defer rollback()

	file, revision, err := fStore.GetRevision(ctx, cached.Instance.Revision)
	if err != nil {
		return err
	}

	rollback()

	cached.File = file
	cached.Revision = revision

	resp := new(grpc.AwaitWorkflowResponse)

	err = bytedata.ConvertDataForOutput(cached.Instance, &resp.Instance)
	if err != nil {
		return err
	}

	rwf := new(grpc.InstanceWorkflow)
	rwf.Name = cached.File.Name()
	rwf.Parent = cached.Dir()
	rwf.Path = cached.File.Path
	resp.Namespace = cached.Namespace.Name
	rwf.Revision = cached.Revision.ID.String()
	resp.Workflow = rwf

	if cached.Instance.Status == util.InstanceStatusComplete {
		runtime, err := flow.database.InstanceRuntime(ctx, cached.Instance.Runtime)
		if err != nil {
			return err
		}
		resp.Data = []byte(runtime.Output)
	}

	nhash = bytedata.Checksum(resp)
	if nhash != phash {
		err = srv.Send(resp)
		if err != nil {
			return err
		}
	}
	phash = nhash

	if cached.Instance.Status != util.InstanceStatusPending {
		return nil
	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend
}
