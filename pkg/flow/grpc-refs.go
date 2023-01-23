package flow

import (
	"context"
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entref "github.com/direktiv/direktiv/pkg/flow/ent/ref"
	entwf "github.com/direktiv/direktiv/pkg/flow/ent/workflow"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/util"
)

func loadSource(rev *database.Revision) (*model.Workflow, error) {

	workflow := new(model.Workflow)

	err := workflow.Load(rev.Source)
	if err != nil {
		return nil, err
	}

	return workflow, nil

}

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

	cached, err := flow.traverseToWorkflow(ctx, nil, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(nil)

	query := clients.Ref.Query().Where(entref.HasWorkflowWith(entwf.ID(cached.Workflow.ID)), entref.Immutable(false))

	results, pi, err := paginate[*ent.RefQuery, *ent.Ref](ctx, req.Pagination, query, refsOrderings, refsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.TagsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return nil, err
	}

	err = atob(cached.Inode(), &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Node.Path = cached.Path()
	resp.Node.Parent = cached.Dir()

	return resp, nil

}

func (flow *flow) TagsStream(req *grpc.TagsRequest, srv grpc.Flow_TagsStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	cached, err := flow.traverseToWorkflow(ctx, nil, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflow(cached)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(nil)

	query := clients.Ref.Query().Where(entref.HasWorkflowWith(entwf.ID(cached.Workflow.ID)), entref.Immutable(false))

	results, pi, err := paginate[*ent.RefQuery, *ent.Ref](ctx, req.Pagination, query, refsOrderings, refsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.TagsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return err
	}

	err = atob(cached.Inode(), &resp.Node)
	if err != nil {
		return err
	}

	resp.Node.Path = cached.Path()
	resp.Node.Parent = cached.Dir()

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

	cached, err := flow.traverseToWorkflow(ctx, nil, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(nil)

	query := clients.Ref.Query().Where(entref.HasWorkflowWith(entwf.ID(cached.Workflow.ID)))

	results, pi, err := paginate[*ent.RefQuery, *ent.Ref](ctx, req.Pagination, query, refsOrderings, refsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.RefsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return nil, err
	}

	err = atob(cached.Inode(), &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Node.Path = cached.Path()
	resp.Node.Parent = cached.Dir()

	return resp, nil

}

func (flow *flow) RefsStream(req *grpc.RefsRequest, srv grpc.Flow_RefsStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	cached, err := flow.traverseToWorkflow(ctx, nil, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflow(cached)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(nil)

	query := clients.Ref.Query().Where(entref.HasWorkflowWith(entwf.ID(cached.Workflow.ID)))

	results, pi, err := paginate[*ent.RefQuery, *ent.Ref](ctx, req.Pagination, query, refsOrderings, refsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.RefsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
	if err != nil {
		return err
	}

	err = atob(cached.Inode(), &resp.Node)
	if err != nil {
		return err
	}

	resp.Node.Path = cached.Path()
	resp.Node.Parent = cached.Dir()

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

	tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToRef(ctx, tx, req.GetNamespace(), req.GetPath(), req.GetRef())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tx)

	err = clients.Ref.Create().SetImmutable(false).SetName(req.GetTag()).SetRevisionID(cached.Revision.ID).SetWorkflowID(cached.Workflow.ID).Exec(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), cached, "Tagged workflow: %s -> %s.", req.GetTag(), cached.Revision.ID.String())
	flow.pubsub.NotifyWorkflow(cached.Workflow)

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) Untag(ctx context.Context, req *grpc.UntagRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToRef(ctx, tx, req.GetNamespace(), req.GetPath(), req.GetTag())
	if err != nil {
		return nil, err
	}

	if cached.Ref.Immutable || cached.Ref.Name == latest {
		return nil, errors.New("not a tag")
	}

	err = flow.configureRouter(ctx, tx, cached, rcfBreaking,
		func() error {

			clients := flow.edb.Clients(tx)
			err = clients.Ref.DeleteOneID(cached.Ref.ID).Exec(ctx)
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

	flow.logToWorkflow(ctx, time.Now(), cached, "Deleted workflow tag: %s.", req.GetTag())
	flow.pubsub.NotifyWorkflow(cached.Workflow)

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) Retag(ctx context.Context, req *grpc.RetagRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToRef(ctx, tx, req.GetNamespace(), req.GetPath(), req.GetRef())
	if err != nil {
		return nil, err
	}

	dt, err := flow.traverseToRef(ctx, tx, req.GetNamespace(), req.GetPath(), req.GetTag())
	if err != nil {
		return nil, err
	}

	if dt.Revision.ID == cached.Revision.ID {
		// no change
		rollback(tx)
		goto respond
	}

	if dt.Ref.Immutable || dt.Ref.Name == latest {
		return nil, errors.New("not a tag")
	}

	err = flow.configureRouter(ctx, tx, cached, rcfBreaking,
		func() error {

			clients := flow.edb.Clients(tx)
			err = clients.Ref.UpdateOneID(dt.Ref.ID).SetRevisionID(cached.Revision.ID).Exec(ctx)
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

	flow.logToWorkflow(ctx, time.Now(), cached, "Changed workflow tag: %s -> %s.", req.GetTag(), cached.Revision.ID.String())
	flow.pubsub.NotifyWorkflow(cached.Workflow)

respond:

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) ValidateRef(ctx context.Context, req *grpc.ValidateRefRequest) (*grpc.ValidateRefResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, err := flow.traverseToRef(ctx, nil, req.GetNamespace(), req.GetPath(), req.GetRef())
	if err != nil {
		return nil, err
	}

	_, err = loadSource(cached.Revision)

	var resp grpc.ValidateRefResponse

	resp.Namespace = cached.Namespace.Name
	resp.Path = cached.Path()
	resp.Ref = cached.Ref.Name
	resp.Invalid = err != nil
	resp.Reason = err.Error()
	resp.Compiles = err != nil

	return &resp, nil

}
