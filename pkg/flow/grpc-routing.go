package flow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/core"
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

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	fStore, store, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	var router = &routerData{
		Enabled: true,
		Routes:  make(map[string]int),
	}

	annotations, err := store.FileAnnotations().Get(ctx, file.ID)
	if err != nil {
		if !errors.Is(err, core.ErrFileAnnotationsNotSet) {
			return nil, err
		}
	} else {
		s := annotations.Data.GetEntry(routerAnnotationKey)
		if s != "" && s != `""` && s != `\"\"` {
			err = json.Unmarshal([]byte(s), &router)
			if err != nil {
				return nil, err
			}
		}
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

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	fStore, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	annotations, err := store.FileAnnotations().Get(ctx, file.ID)
	if err != nil {
		if !errors.Is(err, core.ErrFileAnnotationsNotSet) {
			return nil, err
		}
		annotations = &core.FileAnnotations{
			FileID: file.ID,
			Data:   core.NewFileAnnotationsData(make(map[string]string)),
		}
	}

	s := annotations.Data.GetEntry(routerAnnotationKey)
	router := &routerData{
		Enabled: true,
		Routes:  make(map[string]int),
	}
	if s != "" && s != `""` {
		err = json.Unmarshal([]byte(s), &router)
		if err != nil {
			return nil, err
		}
	}

	router.Enabled = req.Live
	router.Routes = make(map[string]int)
	for _, r := range req.Route {
		router.Routes[r.Ref] = int(r.Weight)
	}

	annotations.Data = annotations.Data.SetEntry(routerAnnotationKey, router.Marshal())

	err = store.FileAnnotations().Set(ctx, annotations)
	if err != nil {
		return nil, err
	}

	err = commit(ctx)
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

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	fStore, store, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	_, verr, err := flow.validateRouter(ctx, fStore, store, file)
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
