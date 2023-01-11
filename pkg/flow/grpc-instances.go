package flow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) InstanceInput(ctx context.Context, req *grpc.InstanceInputRequest) (*grpc.InstanceInputResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	nsc := flow.db.Namespace

	d, err := flow.getInstance(ctx, nsc, req.GetNamespace(), req.GetInstance(), true)
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceInputResponse

	err = atob(d.in, &resp.Instance)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(d.in.Edges.Runtime.Input, &m)
	if err != nil {
		return nil, err
	}
	delete(m, "private")
	input := marshal(m)

	resp.Data = []byte(input)
	resp.Namespace = d.namespace()

	return &resp, nil

}

func (flow *flow) InstanceOutput(ctx context.Context, req *grpc.InstanceOutputRequest) (*grpc.InstanceOutputResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	nsc := flow.db.Namespace

	d, err := flow.getInstance(ctx, nsc, req.GetNamespace(), req.GetInstance(), true)
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceOutputResponse

	err = atob(d.in, &resp.Instance)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal([]byte(d.in.Edges.Runtime.Output), &m)
	if err != nil {
		return nil, err
	}
	delete(m, "private")
	output := marshal(m)

	resp.Data = []byte(output)
	resp.Namespace = d.namespace()

	return &resp, nil

}

func (flow *flow) InstanceMetadata(ctx context.Context, req *grpc.InstanceMetadataRequest) (*grpc.InstanceMetadataResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	nsc := flow.db.Namespace

	d, err := flow.getInstance(ctx, nsc, req.GetNamespace(), req.GetInstance(), true)
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceMetadataResponse

	err = atob(d.in, &resp.Instance)
	if err != nil {
		return nil, err
	}

	resp.Data = []byte(d.in.Edges.Runtime.Metadata)
	resp.Namespace = d.namespace()

	return &resp, nil

}

func (flow *flow) Instances(ctx context.Context, req *grpc.InstancesRequest) (*grpc.InstancesResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.getNamespace(ctx, flow.db.Namespace, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	query := ns.QueryInstances()

	results, pi, err := paginate[*ent.InstanceQuery, *ent.Instance](ctx, req.Pagination, query, instancesOrderings, instancesFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.InstancesResponse)
	resp.Namespace = ns.Name
	resp.Instances = new(grpc.Instances)
	resp.Instances.PageInfo = pi

	err = atob(results, &resp.Instances.Results)
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

	ns, err := flow.getNamespace(ctx, flow.db.Namespace, req.GetNamespace())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeInstances(ns)
	defer flow.cleanup(sub.Close)

resend:

	query := ns.QueryInstances()

	results, pi, err := paginate[*ent.InstanceQuery, *ent.Instance](ctx, req.Pagination, query, instancesOrderings, instancesFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.InstancesResponse)
	resp.Namespace = ns.Name
	resp.Instances = new(grpc.Instances)
	resp.Instances.PageInfo = pi

	err = atob(results, &resp.Instances.Results)
	if err != nil {
		return err
	}

	nhash = checksum(resp)
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

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace

	d, err := flow.traverseToInstance(ctx, nsc, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceResponse

	err = atob(d.in, &resp.Instance)
	if err != nil {
		return nil, err
	}

	if d.in.Edges.Runtime != nil {
		resp.Flow = d.in.Edges.Runtime.Flow
		if caller := d.in.Edges.Runtime.Edges.Caller; caller != nil {
			resp.InvokedBy = caller.ID.String()
		}
	}

	resp.Namespace = d.namespace()

	rwf := new(grpc.InstanceWorkflow)
	rwf.Name = d.base
	rwf.Parent = d.dir
	rwf.Path = d.path
	if d.in.Edges.Revision != nil {
		rwf.Revision = d.in.Edges.Revision.ID.String()
	}
	resp.Workflow = rwf

	return &resp, nil

}

func (flow *flow) InstanceStream(req *grpc.InstanceRequest, srv grpc.Flow_InstanceStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	var sub *subscription

resend:

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	nsc := tx.Namespace

	d, err := flow.traverseToInstance(ctx, nsc, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return err
	}

	rollback(tx)

	if sub == nil {
		sub = flow.pubsub.SubscribeInstance(d.in)
		defer flow.cleanup(sub.Close)
		goto resend
	}

	resp := new(grpc.InstanceResponse)

	err = atob(d.in, &resp.Instance)
	if err != nil {
		return err
	}

	if d.in.Edges.Runtime != nil {
		resp.Flow = d.in.Edges.Runtime.Flow
		if caller := d.in.Edges.Runtime.Edges.Caller; caller != nil {
			resp.InvokedBy = caller.ID.String()
		}
	}

	resp.Namespace = d.namespace()

	rwf := new(grpc.InstanceWorkflow)
	rwf.Name = d.base
	rwf.Parent = d.dir
	rwf.Path = d.path
	if d.in.Edges.Revision != nil {
		rwf.Revision = d.in.Edges.Revision.ID.String()
	}
	resp.Workflow = rwf

	nhash = checksum(resp)
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
	args.Caller = "api"

	im, err := flow.engine.NewInstance(ctx, args)
	if err != nil {
		flow.sugar.Debugf("Error returned to gRPC request %s: %v", this(), err)
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

	im, err := flow.engine.getInstanceMemory(ctx, flow.db.Instance, req.GetInstance())
	if err != nil {
		return nil, err
	}

	if im.in.Edges.Namespace.Name != req.GetNamespace() {
		return nil, errors.New("instance not found")
	}

	if im.in.Status != util.InstanceStatusPending {
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

	d, err := flow.getInstance(ctx, flow.db.Namespace, req.GetNamespace(), req.GetInstance(), false)
	if err != nil {
		return nil, err
	}

	flow.engine.cancelInstance(d.in.ID.String(), "direktiv.cancels.api", "cancelled by api request", false)

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
	args.Caller = "api"

	im, err := flow.engine.NewInstance(ctx, args)
	if err != nil {
		flow.sugar.Debugf("Error returned to gRPC request %s: %v", this(), err)
		return err
	}

	var sub *subscription

	flow.engine.queue(im)

	var d *instData

resend:

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	nsc := tx.Namespace

	if d == nil {
		d, err = flow.traverseToInstance(ctx, nsc, req.GetNamespace(), im.in.ID.String())
		if err != nil {
			return err
		}
	}

	d, err = flow.fastGetInstance(ctx, d)
	if err != nil {
		return err
	}

	rollback(tx)

	if sub == nil {
		sub = flow.pubsub.SubscribeInstance(d.in)
		defer flow.cleanup(sub.Close)
		goto resend
	}

	resp := new(grpc.AwaitWorkflowResponse)

	err = atob(d.in, &resp.Instance)
	if err != nil {
		return err
	}

	if d.in.Edges.Runtime != nil {
		resp.Flow = d.in.Edges.Runtime.Flow
		if caller := d.in.Edges.Runtime.Edges.Caller; caller != nil {
			resp.InvokedBy = caller.ID.String()
		}

		if d.in.Status == util.InstanceStatusComplete {
			resp.Data = []byte(d.in.Edges.Runtime.Output)
		}
	}

	resp.Namespace = d.namespace()

	rwf := new(grpc.InstanceWorkflow)
	rwf.Name = d.base
	rwf.Parent = d.dir
	rwf.Path = d.path
	if d.in.Edges.Revision != nil {
		rwf.Revision = d.in.Edges.Revision.ID.String()
	}
	resp.Workflow = rwf

	nhash = checksum(resp)
	if nhash != phash {
		err = srv.Send(resp)
		if err != nil {
			return err
		}
	}
	phash = nhash

	if d.in.Status != util.InstanceStatusPending {
		return nil
	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend

}
