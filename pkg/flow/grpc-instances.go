package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.InstancePaginateOption{}
	opts = append(opts, instancesOrder(p)...)
	opts = append(opts, instancesFilter(p)...)

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
	opts = append(opts, instancesOrder(p)...)
	opts = append(opts, instancesFilter(p)...)

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
	args.Caller = "api"

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

func instancesOrder(p *pagination) []ent.InstancePaginateOption {

	var opts []ent.InstancePaginateOption

	for _, o := range p.order {

		if o == nil {
			continue
		}

		field := ent.InstanceOrderFieldCreatedAt
		direction := ent.OrderDirectionDesc

		if x := o.Field; x != "" && x == "ID" {
			field = ent.InstanceOrderFieldID
		}

		if x := o.Field; x != "" && x == "CREATED" {
			field = ent.InstanceOrderFieldCreatedAt
		}

		if x := o.Direction; x != "" && x == "DESC" {
			direction = ent.OrderDirectionDesc
		}

		if x := o.Direction; x != "" && x == "ASC" {
			direction = ent.OrderDirectionAsc
		}

		opts = append(opts, ent.WithInstanceOrder(&ent.InstanceOrder{
			Direction: direction,
			Field:     field,
		}))

	}

	if len(opts) == 0 {
		opts = append(opts, ent.WithInstanceOrder(&ent.InstanceOrder{
			Direction: ent.OrderDirectionDesc,
			Field:     ent.InstanceOrderFieldCreatedAt,
		}))
	}

	return opts

}

func instancesFilter(p *pagination) []ent.InstancePaginateOption {

	var opts []ent.InstancePaginateOption

	if p.filter == nil {
		return nil
	}

	for i := range p.filter {

		f := p.filter[i]

		if f == nil {
			continue
		}

		filter := f.Val

		opts = append(opts, ent.WithInstanceFilter(func(query *ent.InstanceQuery) (*ent.InstanceQuery, error) {

			if filter == "" {
				return query, nil
			}

			field := f.Field
			if field == "" {
				return query, nil
			}

			switch field {
			case "AS":

				ftype := f.Type

				switch ftype {
				case "WORKFLOW":
					return query.Where(entinst.AsHasPrefix(filter)), nil
				case "":
					fallthrough
				case "CONTAINS":
					return query.Where(entinst.AsContains(filter)), nil
				default:
					return nil, fmt.Errorf("unexpected filter type")
				}

			case "CREATED":

				ftype := f.Type
				t, err := time.Parse(time.RFC822, filter)
				if err != nil {
					return nil, err
				}

				switch ftype {

				case "AFTER":
					return query.Where(entinst.CreatedAtLTE(t)), nil
				case "BEFORE":
					return query.Where(entinst.CreatedAtGTE(t)), nil
				case "":
					fallthrough
				default:
					return nil, fmt.Errorf("unexpected filter type")
				}

			case "STATUS":

				ftype := f.Type

				switch ftype {
				case "MATCH":
					return query.Where(entinst.StatusEQ(filter)), nil
				case "":
					fallthrough
				case "CONTAINS":
					return query.Where(entinst.StatusContains(filter)), nil
				default:
					return nil, fmt.Errorf("unexpected filter type")
				}

			case "TRIGGER":

				ftype := f.Type

				switch ftype {
				case "MATCH":
					return query.Where(entinst.InvokerEQ(filter)), nil
				case "":
					fallthrough
				default:
					return nil, fmt.Errorf("unexpected filter type")
				}

			default:
				return nil, fmt.Errorf("bad filter field")

			}

		}))

	}

	return opts

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
