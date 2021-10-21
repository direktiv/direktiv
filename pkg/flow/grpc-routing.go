package flow

import (
	"context"

	"github.com/vorteil/direktiv/pkg/flow/ent"
	entmux "github.com/vorteil/direktiv/pkg/flow/ent/route"
	entwf "github.com/vorteil/direktiv/pkg/flow/ent/workflow"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
)

func (flow *flow) Router(ctx context.Context, req *grpc.RouterRequest) (*grpc.RouterResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	nsc := flow.db.Namespace
	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	routes, err := d.wf.QueryRoutes().Order(ent.Desc(entmux.FieldWeight)).All(ctx)
	if err != nil {
		return nil, err
	}

	var resp grpc.RouterResponse

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Namespace = d.namespace()
	resp.Node.Parent = d.dir
	resp.Node.Path = d.path
	resp.Live = d.wf.Live

	err = atob(routes, &resp.Routes)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}

func (flow *flow) RouterStream(req *grpc.RouterRequest, srv grpc.Flow_RouterStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	nsc := flow.db.Namespace
	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflow(d.wf)
	defer flow.cleanup(sub.Close)

resend:

	routes, err := d.wf.QueryRoutes().Order(ent.Desc(entmux.FieldWeight)).All(ctx)
	if err != nil {
		return err
	}

	resp := new(grpc.RouterResponse)

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return err
	}

	resp.Namespace = d.namespace()
	resp.Node.Parent = d.dir
	resp.Node.Path = d.path
	resp.Live = d.wf.Live

	err = atob(routes, &resp.Routes)
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

func (flow *flow) EditRouter(ctx context.Context, req *grpc.EditRouterRequest) (*grpc.EditRouterResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	var routes []*ent.Route

	err = flow.configureRouter(ctx, tx.Events, &d.wf, rcfBreaking,
		func() error {

			muxc := tx.Route
			_, err = muxc.Delete().Where(entmux.HasWorkflowWith(entwf.ID(d.wf.ID))).Exec(ctx)
			if err != nil {
				return err
			}

			routes, err = d.wf.QueryRoutes().Order(ent.Desc(entmux.FieldWeight)).All(ctx)
			if err != nil {
				return err
			}

			for i := range req.Route {

				route := req.Route[i]
				ref, err := flow.getRef(ctx, d.wf, route.Ref)
				if err != nil {
					return err
				}

				err = muxc.Create().SetWorkflow(d.wf).SetWeight(int(route.Weight)).SetRef(ref).Exec(ctx)
				if err != nil {
					return err
				}

			}

			routes, err = d.wf.QueryRoutes().Order(ent.Desc(entmux.FieldWeight)).WithRef().All(ctx)
			if err != nil {
				return err
			}

			if d.wf.Live != req.GetLive() {
				err = d.wf.Update().SetLive(req.GetLive()).Exec(ctx)
				if err != nil {
					return err
				}
			}

			return nil

		},
		tx.Commit,
	)
	if err != nil {
		return nil, err
	}

	var resp grpc.EditRouterResponse

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Namespace = d.namespace()
	resp.Node.Parent = d.dir
	resp.Node.Path = d.path
	resp.Live = req.GetLive()

	err = atob(routes, &resp.Routes)
	if err != nil {
		return nil, err
	}

	for i := range routes {
		route := routes[i]
		resp.Routes[i].Ref = route.Edges.Ref.Name
	}

	return &resp, nil

}

func (flow *flow) ValidateRouter(ctx context.Context, req *grpc.ValidateRouterRequest) (*grpc.ValidateRouterResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	nsc := flow.db.Namespace
	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	_, verr, err := validateRouter(ctx, d.wf)
	if err != nil {
		return nil, err
	}

	var resp grpc.ValidateRouterResponse

	resp.Namespace = d.namespace()
	resp.Path = d.path
	resp.Invalid = verr != nil
	resp.Reason = verr.Error()

	return &resp, nil

}
