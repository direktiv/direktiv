package flow

import (
	"context"
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entref "github.com/direktiv/direktiv/pkg/flow/ent/ref"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
)

var refsOrderings = []*orderingInfo{
	{
		db:           entref.FieldCreatedAt,
		req:          "CREATED",
		defaultOrder: ent.Desc,
	},
	{
		db:           entref.FieldName,
		req:          util.PaginationKeyName,
		defaultOrder: ent.Asc,
	},
}

var refsFilters = map[*filteringInfo]func(query *ent.RefQuery, v string) (*ent.RefQuery, error){
	{
		field: util.PaginationKeyName,
		ftype: "CONTAINS",
	}: func(query *ent.RefQuery, v string) (*ent.RefQuery, error) {
		return query.Where(entref.NameContains(v)), nil
	},
}

func (flow *flow) Tags(ctx context.Context, req *grpc.TagsRequest) (*grpc.TagsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	d, err := flow.traverseToWorkflow(ctx, flow.db.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	query := d.wf.QueryRefs()
	query = query.Where(entref.Immutable(false))

	results, pi, err := paginate[*ent.RefQuery, *ent.Ref](ctx, req.Pagination, query, refsOrderings, refsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.TagsResponse)
	resp.Namespace = d.namespace()
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return nil, err
	}

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Node.Path = d.path
	resp.Node.Parent = d.dir

	return resp, nil

}

func (flow *flow) TagsStream(req *grpc.TagsRequest, srv grpc.Flow_TagsStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	d, err := flow.traverseToWorkflow(ctx, flow.db.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflow(d.wf)
	defer flow.cleanup(sub.Close)

resend:

	query := d.wf.QueryRefs()
	query = query.Where(entref.Immutable(false))

	results, pi, err := paginate[*ent.RefQuery, *ent.Ref](ctx, req.Pagination, query, refsOrderings, refsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.TagsResponse)
	resp.Namespace = d.namespace()
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return err
	}

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return err
	}

	resp.Node.Path = d.path
	resp.Node.Parent = d.dir

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

	d, err := flow.traverseToWorkflow(ctx, flow.db.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	query := d.wf.QueryRefs()

	results, pi, err := paginate[*ent.RefQuery, *ent.Ref](ctx, req.Pagination, query, refsOrderings, refsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.RefsResponse)
	resp.Namespace = d.namespace()
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return nil, err
	}

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Node.Path = d.path
	resp.Node.Parent = d.dir

	return resp, nil

}

func (flow *flow) RefsStream(req *grpc.RefsRequest, srv grpc.Flow_RefsStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	d, err := flow.traverseToWorkflow(ctx, flow.db.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflow(d.wf)
	defer flow.cleanup(sub.Close)

resend:

	query := d.wf.QueryRefs()

	results, pi, err := paginate[*ent.RefQuery, *ent.Ref](ctx, req.Pagination, query, refsOrderings, refsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.RefsResponse)
	resp.Namespace = d.namespace()
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return err
	}

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return err
	}

	resp.Node.Path = d.path
	resp.Node.Parent = d.dir

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

	flow.logWithTagsToWorkflow(ctx, time.Now(), d, "Tagged workflow: %s -> %s.", req.GetTag(), d.rev().ID.String())
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

	if d.ref.Immutable || d.ref.Name == latest {
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

	flow.logWithTagsToWorkflow(ctx, time.Now(), d, "Deleted workflow tag: %s.", req.GetTag())
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

	if dt.ref.Immutable || dt.ref.Name == latest {
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

	flow.logWithTagsToWorkflow(ctx, time.Now(), d, "Changed workflow tag: %s -> %s.", req.GetTag(), d.rev().ID.String())
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
