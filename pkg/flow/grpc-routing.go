package flow

import (
	"context"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
)

func (flow *flow) Router(ctx context.Context, req *grpc.RouterRequest) (*grpc.RouterResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		fmt.Println("A", err)
		return nil, err
	}

	fStore, _, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		fmt.Println("B", err)
		return nil, err
	}
	defer rollback(ctx)

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		fmt.Println("C", err)
		return nil, err
	}

	// NOTE: this is fake output
	resp := &grpc.RouterResponse{}
	resp.Namespace = ns.Name
	resp.Live = true
	resp.Routes = make([]*grpc.Route, 0)
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	return resp, nil

	// TODO: yassir, refactor
	/*
		cached, err := flow.traverseToWorkflow(ctx, req.GetNamespace(), req.GetPath())
		if err != nil {
			return nil, err
		}

		var resp grpc.RouterResponse

		err = bytedata.ConvertDataForOutput(cached.Inode(), &resp.Node)
		if err != nil {
			return nil, err
		}

		resp.Namespace = cached.Namespace.Name
		resp.Node.Parent = cached.Dir()
		resp.Node.Path = cached.Path()
		resp.Live = cached.Workflow.Live

		err = bytedata.ConvertDataForOutput(cached.Workflow.Routes, &resp.Routes)
		if err != nil {
			return nil, err
		}

		for i := range cached.Workflow.Routes {
			route := cached.Workflow.Routes[i]
			resp.Routes[i].Ref = route.Ref.Name
		}

		return &resp, nil
	*/
}

func (flow *flow) RouterStream(req *grpc.RouterRequest, srv grpc.Flow_RouterStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	resp, err := flow.Router(ctx, req)
	if err != nil {
		return err
	}

	// mock streaming response.
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			err = srv.Send(resp)
			if err != nil {
				return err
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func (flow *flow) EditRouter(ctx context.Context, req *grpc.EditRouterRequest) (*grpc.EditRouterResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	//nolint
	return nil, nil
	// TODO: yassir, refactor
	/*
		tctx, tx, err := flow.database.Tx(ctx)
		if err != nil {
			return nil, err
		}
		defer rollback(tx)

		cached, err := flow.traverseToWorkflow(tctx, req.GetNamespace(), req.GetPath())
		if err != nil {
			return nil, err
		}

		var routes []*ent.Route

		clients := flow.edb.Clients(tctx)

		err = flow.configureRouter(tctx, cached, rcfBreaking,
			func() error {
				_, err = clients.Route.Delete().Where(entmux.HasWorkflowWith(entwf.ID(cached.Workflow.ID))).Exec(tctx)
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

					err = clients.Route.Create().SetWorkflowID(cached.Workflow.ID).SetWeight(int(route.Weight)).SetRefID(ref.ID).Exec(tctx)
					if err != nil {
						return err
					}

				}

				if cached.Workflow.Live != req.GetLive() {
					err = clients.Workflow.UpdateOneID(cached.Workflow.ID).SetLive(req.GetLive()).Exec(tctx)
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

		err = bytedata.ConvertDataForOutput(cached.Inode(), &resp.Node)
		if err != nil {
			return nil, err
		}

		resp.Namespace = cached.Namespace.Name
		resp.Node.Parent = cached.Dir()
		resp.Node.Path = cached.Path()
		resp.Live = req.GetLive()

		err = bytedata.ConvertDataForOutput(routes, &resp.Routes)
		if err != nil {
			return nil, err
		}

		for i := range routes {
			route := routes[i]
			resp.Routes[i].Ref = route.Edges.Ref.Name
		}
		return &resp, nil
	*/
}

func (flow *flow) ValidateRouter(ctx context.Context, req *grpc.ValidateRouterRequest) (*grpc.ValidateRouterResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	//nolint
	return nil, nil
	// TODO: yassir, refactor
	/*
		cached, err := flow.traverseToWorkflow(ctx, req.GetNamespace(), req.GetPath())
		if err != nil {
			return nil, err
		}

		_, verr, err := flow.validateRouter(ctx, cached)
		if err != nil {
			return nil, err
		}

		var resp grpc.ValidateRouterResponse

		resp.Namespace = cached.Namespace.Name
		resp.Path = cached.Path()
		resp.Invalid = verr != nil
		resp.Reason = verr.Error()

		return &resp, nil
	*/
}
