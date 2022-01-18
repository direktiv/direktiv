package flow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entref "github.com/direktiv/direktiv/pkg/flow/ent/ref"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
)

func (flow *flow) Revisions(ctx context.Context, req *grpc.RevisionsRequest) (*grpc.RevisionsResponse, error) {

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
	query = query.Where(entref.Immutable(true))
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.RevisionsResponse

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

func (flow *flow) RevisionsStream(req *grpc.RevisionsRequest, srv grpc.Flow_RevisionsStreamServer) error {

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
	query = query.Where(entref.Immutable(true))
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	resp := new(grpc.RevisionsResponse)

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

func (flow *flow) DeleteRevision(ctx context.Context, req *grpc.DeleteRevisionRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	fmt.Println("A")

	nsc := tx.Namespace
	d, err := flow.traverseToRef(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetRevision())
	if err != nil {
		fmt.Println("B")
		return nil, err
	}

	if d.ref.Immutable != true {
		fmt.Println("C")
		return nil, errors.New("not a revision")
	}

	xrefs, err := d.rev().QueryRefs().Where(entref.ImmutableEQ(false)).All(ctx)
	if err != nil {
		fmt.Println("D")
		return nil, err
	}

	if len(xrefs) > 1 || (len(xrefs) == 1 && xrefs[0].Name != "latest") {
		fmt.Println("E")
		return nil, errors.New("cannot delete revision while refs to it exist")
	}

	if len(xrefs) == 1 && xrefs[0].Name == "latest" {
		err = flow.configureRouter(ctx, tx.Events, &d.wf, rcfBreaking,
			func() error {

				fmt.Println("G1")

				refc := tx.Ref
				err := refc.DeleteOne(d.ref).Exec(ctx)
				if err != nil {
					fmt.Println("H1")
					return err
				}

				fmt.Println("I1")

				return nil

			},
			tx.Commit,
		)
		if err != nil {
			fmt.Println("J1")
			return nil, err
		}
	} else {
		err = flow.configureRouter(ctx, tx.Events, &d.wf, rcfBreaking,
			func() error {

				fmt.Println("G2")

				revc := tx.Revision
				err := revc.DeleteOne(d.rev()).Exec(ctx)
				if err != nil {
					fmt.Println("H2")
					return err
				}

				fmt.Println("I2")

				return nil

			},
			tx.Commit,
		)
		if err != nil {
			fmt.Println("J2")
			return nil, err
		}
	}

	// dl, err := flow.traverseToRef(ctx, nsc, req.GetNamespace(), req.GetPath(), "")
	// if err != nil {
	// 	return nil, err
	// }

	// if dl.rev().ID == d.rev().ID {
	// 	return nil, errors.New("cannot delete latest")
	// }

	fmt.Println("F")

	fmt.Println("K")

	flow.logToWorkflow(ctx, time.Now(), d.wfData, "Deleted workflow revision: %s.", d.rev().ID.String())
	flow.pubsub.NotifyWorkflow(d.wf)

	var resp emptypb.Empty

	return &resp, nil

}
