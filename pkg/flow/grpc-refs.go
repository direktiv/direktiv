package flow

import (
	"context"
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entref "github.com/direktiv/direktiv/pkg/flow/ent/ref"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
)

func refOrder(p *pagination) ent.RefPaginateOption {

	field := ent.RefOrderFieldName
	direction := ent.OrderDirectionAsc

	if p.order != nil {

		if x := p.order.Field; x != "" && x == "NAME" {
			field = ent.RefOrderFieldName
		}

		if x := p.order.Direction; x != "" && x == "DESC" {
			direction = ent.OrderDirectionDesc
		}

	}

	return ent.WithRefOrder(&ent.RefOrder{
		Direction: direction,
		Field:     field,
	})

}

func refFilter(p *pagination) ent.RefPaginateOption {

	if p.filter == nil {
		return nil
	}

	filter := p.filter.Val

	return ent.WithRefFilter(func(query *ent.RefQuery) (*ent.RefQuery, error) {

		if filter == "" {
			return query, nil
		}

		field := p.filter.Field
		if field == "" {
			return query, nil
		}

		switch field {
		case "NAME":

			ftype := p.filter.Type
			if ftype == "" {
				return query, nil
			}

			switch ftype {
			case "CONTAINS":
				return query.Where(entref.NameContains(filter)), nil
			}
		}

		return query, nil

	})

}

func (flow *flow) Tags(ctx context.Context, req *grpc.TagsRequest) (*grpc.TagsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.RefPaginateOption{}
	opts = append(opts, refOrder(p))
	filter := refFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace
	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	query := d.wf.QueryRefs()
	query = query.Where(entref.Immutable(false))
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.TagsResponse

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Namespace = d.namespace()
	resp.Node.Path = d.path
	resp.Node.Parent = d.dir

	err = atob(cx, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}

func (flow *flow) TagsStream(req *grpc.TagsRequest, srv grpc.Flow_TagsStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	opts := []ent.RefPaginateOption{}
	opts = append(opts, refOrder(p))
	filter := refFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace
	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflow(d.wf)
	defer flow.cleanup(sub.Close)

resend:

	query := d.wf.QueryRefs()
	query = query.Where(entref.Immutable(false))
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	resp := new(grpc.TagsResponse)

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return err
	}

	resp.Namespace = d.namespace()
	resp.Node.Path = d.path
	resp.Node.Parent = d.dir

	err = atob(cx, resp)
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

func (flow *flow) Refs(ctx context.Context, req *grpc.RefsRequest) (*grpc.RefsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.RefPaginateOption{}
	opts = append(opts, refOrder(p))
	filter := refFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace
	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	query := d.wf.QueryRefs()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.RefsResponse

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Namespace = d.namespace()
	resp.Node.Path = d.path
	resp.Node.Parent = d.dir

	err = atob(cx, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}

func (flow *flow) RefsStream(req *grpc.RefsRequest, srv grpc.Flow_RefsStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	opts := []ent.RefPaginateOption{}
	opts = append(opts, refOrder(p))
	filter := refFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace
	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflow(d.wf)
	defer flow.cleanup(sub.Close)

resend:

	query := d.wf.QueryRefs()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	resp := new(grpc.RefsResponse)

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return err
	}

	resp.Namespace = d.namespace()
	resp.Node.Path = d.path
	resp.Node.Parent = d.dir

	err = atob(cx, resp)
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

func (flow *flow) Tag(ctx context.Context, req *grpc.TagRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToRef(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetRef())
	if err != nil {
		return nil, err
	}

	refc := tx.Ref
	err = refc.Create().SetImmutable(false).SetName(req.GetTag()).SetRevision(d.rev()).SetWorkflow(d.wf).Exec(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), d.wfData, "Tagged workflow: %s -> %s.", req.GetTag(), d.rev().ID.String())
	flow.pubsub.NotifyWorkflow(d.wf)

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) Untag(ctx context.Context, req *grpc.UntagRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToRef(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetTag())
	if err != nil {
		return nil, err
	}

	if d.ref.Immutable == true || d.ref.Name == latest {
		return nil, errors.New("not a tag")
	}

	err = flow.configureRouter(ctx, tx.Events, &d.wf, rcfBreaking,
		func() error {

			refc := tx.Ref
			err = refc.DeleteOne(d.ref).Exec(ctx)
			if err != nil {
				return err
			}

			return nil

		},
		tx.Commit,
	)
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), d.wfData, "Deleted workflow tag: %s.", req.GetTag())
	flow.pubsub.NotifyWorkflow(d.wf)

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) Retag(ctx context.Context, req *grpc.RetagRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToRef(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetRef())
	if err != nil {
		return nil, err
	}

	dt, err := flow.traverseToRef(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetTag())
	if err != nil {
		return nil, err
	}

	if dt.rev().ID == d.rev().ID {
		// no change
		rollback(tx)
		goto respond
	}

	if dt.ref.Immutable == true || dt.ref.Name == latest {
		return nil, errors.New("not a tag")
	}

	err = flow.configureRouter(ctx, tx.Events, &d.wf, rcfBreaking,
		func() error {

			err = dt.ref.Update().SetRevision(d.rev()).Exec(ctx)
			if err != nil {
				return err
			}

			return nil

		},
		tx.Commit,
	)
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), d.wfData, "Changed workflow tag: %s -> %s.", req.GetTag(), d.rev().ID.String())
	flow.pubsub.NotifyWorkflow(d.wf)

respond:

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) ValidateRef(ctx context.Context, req *grpc.ValidateRefRequest) (*grpc.ValidateRefResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	nsc := flow.db.Namespace
	d, err := flow.traverseToRef(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetRef())
	if err != nil {
		return nil, err
	}

	_, err = loadSource(d.rev())

	var resp grpc.ValidateRefResponse

	resp.Namespace = d.namespace()
	resp.Path = d.path
	resp.Ref = d.ref.Name
	resp.Invalid = err != nil
	resp.Reason = err.Error()
	resp.Compiles = err != nil

	return &resp, nil

}
