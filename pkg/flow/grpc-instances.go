package flow

import (
	"context"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
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

	resp.Data = []byte(d.in.Edges.Runtime.Input)
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

	resp.Data = []byte(d.in.Edges.Runtime.Output)
	resp.Namespace = d.namespace()

	return &resp, nil

}

func (flow *flow) Instances(ctx context.Context, req *grpc.InstancesRequest) (*grpc.InstancesResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.InstancePaginateOption{}
	opts = append(opts, instancesOrder(p))
	filter := instancesFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace
	ns, err := flow.getNamespace(ctx, nsc, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	query := ns.QueryInstances()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.InstancesResponse
	resp.Instances = new(grpc.Instances)
	resp.Instances.PageInfo = new(grpc.PageInfo)
	resp.Namespace = ns.Name

	err = atob(cx, &resp.Instances)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}

func (flow *flow) InstancesStream(req *grpc.InstancesRequest, srv grpc.Flow_InstancesStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	opts := []ent.InstancePaginateOption{}
	opts = append(opts, instancesOrder(p))
	filter := instancesFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace
	ns, err := flow.getNamespace(ctx, nsc, req.GetNamespace())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeInstances(ns)
	defer flow.cleanup(sub.Close)

resend:

	query := ns.QueryInstances()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	resp := new(grpc.InstancesResponse)

	resp.Namespace = ns.Name

	err = atob(cx, &resp.Instances)
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
	args.Caller = "API"

	im, err := flow.engine.NewInstance(ctx, args)
	if err != nil {
		flow.sugar.Debugf("Error returned to gRPC request %s: %v", this(), err)
		return nil, err
	}

	flow.engine.queue(im)

	var resp grpc.StartWorkflowResponse

	resp.Namespace = req.GetNamespace()
	resp.Instance = im.ID().String()

	return &resp, nil

}

func instancesOrder(p *pagination) ent.InstancePaginateOption {

	field := ent.InstanceOrderFieldCreatedAt
	direction := ent.OrderDirectionDesc

	if p.order != nil {

		if x := p.order.Field; x != "" && x == "ID" {
			field = ent.InstanceOrderFieldID
		}

		if x := p.order.Field; x != "" && x == "CREATED" {
			field = ent.InstanceOrderFieldCreatedAt
		}

		if x := p.order.Direction; x != "" && x == "DESC" {
			direction = ent.OrderDirectionDesc
		}

		if x := p.order.Direction; x != "" && x == "ASC" {
			direction = ent.OrderDirectionAsc
		}

	}

	return ent.WithInstanceOrder(&ent.InstanceOrder{
		Direction: direction,
		Field:     field,
	})

}

func instancesFilter(p *pagination) ent.InstancePaginateOption {

	if p.filter == nil {
		return nil
	}

	filter := p.filter.Val

	return ent.WithInstanceFilter(func(query *ent.InstanceQuery) (*ent.InstanceQuery, error) {

		if filter == "" {
			return query, nil
		}

		field := p.filter.Field
		if field == "" {
			return query, nil
		}

		switch field {
		case "WORKFLOW":

			ftype := p.filter.Type
			if ftype == "" {
				return query, nil
			}

			switch ftype {
			case "AS":
				return query.Where(entinst.AsHasPrefix(filter + ":")), nil
			case "CONTAINS":
				return query.Where(entinst.AsContains(filter)), nil
			}
		}

		return query, nil

	})

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
