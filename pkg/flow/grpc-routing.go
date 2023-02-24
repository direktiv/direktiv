package flow

import (
	"context"
	"os"

	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entmux "github.com/direktiv/direktiv/pkg/flow/ent/route"
	entwf "github.com/direktiv/direktiv/pkg/flow/ent/workflow"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
)

func (flow *flow) Router(ctx context.Context, req *grpc.RouterRequest) (*grpc.RouterResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, err := flow.traverseToWorkflow(ctx, nil, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	var resp grpc.RouterResponse

	err = atob(cached.Inode(), &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = cached.Dir()
	resp.Node.Path = cached.Path()
	resp.Live = cached.Workflow.Live

	err = atob(cached.Workflow.Routes, &resp.Routes)
	if err != nil {
		return nil, err
	}

	for i := range cached.Workflow.Routes {
		route := cached.Workflow.Routes[i]
		resp.Routes[i].Ref = route.Ref.Name
	}

	return &resp, nil
}

func (flow *flow) RouterStream(req *grpc.RouterRequest, srv grpc.Flow_RouterStreamServer) error {
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

	resp := new(grpc.RouterResponse)

	err = atob(cached.Inode(), &resp.Node)
	if err != nil {
		return err
	}

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = cached.Dir()
	resp.Node.Path = cached.Path()
	resp.Live = cached.Workflow.Live

	err = atob(cached.Workflow.Routes, &resp.Routes)
	if err != nil {
		return err
	}

	for i := range cached.Workflow.Routes {
		route := cached.Workflow.Routes[i]
		resp.Routes[i].Ref = route.Ref.Name
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

	tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToWorkflow(ctx, tx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	var routes []*ent.Route

	clients := flow.edb.Clients(tx)

	err = flow.configureRouter(ctx, tx, cached, rcfBreaking,
		func() error {
			_, err = clients.Route.Delete().Where(entmux.HasWorkflowWith(entwf.ID(cached.Workflow.ID))).Exec(ctx)
			if err != nil {
				return err
			}

			for i := range req.Route {

				route := req.Route[i]

				// if the api sends a 0 we don't add it at all
				if route.Weight == 0 {
					continue
				}

				var ref *database.Ref

				for idx := range cached.Workflow.Refs {
					if cached.Workflow.Refs[idx].Name == route.Ref {
						ref = cached.Workflow.Refs[idx]
						break
					}
				}

				if ref == nil {
					return os.ErrNotExist
				}

				err = clients.Route.Create().SetWorkflowID(cached.Workflow.ID).SetWeight(int(route.Weight)).SetRefID(ref.ID).Exec(ctx)
				if err != nil {
					return err
				}

			}

			if cached.Workflow.Live != req.GetLive() {
				err = clients.Workflow.UpdateOneID(cached.Workflow.ID).SetLive(req.GetLive()).Exec(ctx)
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

	err = atob(cached.Inode(), &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = cached.Dir()
	resp.Node.Path = cached.Path()
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

	cached, err := flow.traverseToWorkflow(ctx, nil, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	_, verr, err := flow.validateRouter(ctx, nil, cached)
	if err != nil {
		return nil, err
	}

	var resp grpc.ValidateRouterResponse

	resp.Namespace = cached.Namespace.Name
	resp.Path = cached.Path()
	resp.Invalid = verr != nil
	resp.Reason = verr.Error()

	return &resp, nil
}
