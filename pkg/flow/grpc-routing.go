package flow

import (
	"context"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
)

func convertRoutesForOutput(router *routerData) []*grpc.Route {
	routes := make([]*grpc.Route, 0)
	for k, v := range router.Routes {
		routes = append(routes, &grpc.Route{
			Ref:    k,
			Weight: int32(v),
		})
	}
	return routes
}

func (flow *flow) Router(ctx context.Context, req *grpc.RouterRequest) (*grpc.RouterResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := tx.FileStore().ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	_, router, err := getRouter(ctx, tx, file)
	if err != nil {
		return nil, err
	}

	resp := &grpc.RouterResponse{}
	resp.Namespace = ns.Name
	resp.Live = true
	resp.Routes = make([]*grpc.Route, 0)
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.Live = router.Enabled
	resp.Routes = convertRoutesForOutput(router)
	return resp, nil
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

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := tx.FileStore().ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	annotations, router, err := getRouter(ctx, tx, file)
	if err != nil {
		return nil, err
	}

	router.Enabled = req.Live
	router.Routes = make(map[string]int)
	for _, r := range req.Route {
		router.Routes[r.Ref] = int(r.Weight)
	}

	annotations.Data = annotations.Data.SetEntry(routerAnnotationKey, router.Marshal())

	err = tx.DataStore().FileAnnotations().Set(ctx, annotations)
	if err != nil {
		return nil, err
	}

	err = flow.configureWorkflowStarts(ctx, tx, ns.ID, file, router, true)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	var resp grpc.EditRouterResponse
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.Namespace = ns.Name
	resp.Node.Parent = file.Dir()
	resp.Node.Path = file.Path
	resp.Live = router.Enabled
	resp.Routes = convertRoutesForOutput(router)
	return &resp, nil
}

func (flow *flow) ValidateRouter(ctx context.Context, req *grpc.ValidateRouterRequest) (*grpc.ValidateRouterResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := tx.FileStore().ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	_, verr, err := flow.validateRouter(ctx, tx, file)
	if err != nil {
		return nil, err
	}

	var resp grpc.ValidateRouterResponse

	resp.Namespace = ns.Name
	resp.Path = file.Path
	resp.Invalid = verr != nil
	resp.Reason = verr.Error()

	return &resp, nil
}
