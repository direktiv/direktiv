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

func (flow *flow) Revisions(ctx context.Context, req *grpc.RevisionsRequest) (*grpc.RevisionsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	d, err := flow.traverseToWorkflow(ctx, flow.db.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	query := d.wf.QueryRefs()
	query = query.Where(entref.Immutable(true))

	results, pi, err := paginate[*ent.RefQuery, *ent.Ref](ctx, req.Pagination, query, refsOrderings, refsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.RevisionsResponse)
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

func (flow *flow) RevisionsStream(req *grpc.RevisionsRequest, srv grpc.Flow_RevisionsStreamServer) error {

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
	query = query.Where(entref.Immutable(true))

	results, pi, err := paginate[*ent.RefQuery, *ent.Ref](ctx, req.Pagination, query, refsOrderings, refsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.RevisionsResponse)
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

func (flow *flow) DeleteRevision(ctx context.Context, req *grpc.DeleteRevisionRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToRef(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetRevision())
	if err != nil {
		return nil, err
	}

	if d.ref.Immutable != true {
		return nil, errors.New("not a revision")
	}

	xrefs, err := d.rev().QueryRefs().Where(entref.ImmutableEQ(false)).All(ctx)
	if err != nil {
		return nil, err
	}

	if len(xrefs) > 1 || (len(xrefs) == 1 && xrefs[0].Name != "latest") {
		return nil, errors.New("cannot delete revision while refs to it exist")
	}

	if len(xrefs) == 1 && xrefs[0].Name == "latest" {
		err = flow.configureRouter(ctx, tx.Events, &d.wf, rcfBreaking,
			func() error {

				refc := tx.Ref
				err := refc.DeleteOne(d.ref).Exec(ctx)
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
	} else {
		err = flow.configureRouter(ctx, tx.Events, &d.wf, rcfBreaking,
			func() error {

				revc := tx.Revision
				err := revc.DeleteOne(d.rev()).Exec(ctx)
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
	}

	flow.logToWorkflow(ctx, time.Now(), d.wfData, "Deleted workflow revision: %s.", d.rev().ID.String())
	flow.pubsub.NotifyWorkflow(d.wf)

	var resp emptypb.Empty

	return &resp, nil

}
