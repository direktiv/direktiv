package flow

import (
	"context"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	pubsub2 "github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) ResolveNamespaceUID(ctx context.Context, req *grpc.ResolveNamespaceUIDRequest) (*grpc.NamespaceResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespaceResponse
	resp.Namespace = bytedata.ConvertNamespaceToGrpc(ns)

	return &resp, nil
}

func (flow *flow) Namespace(ctx context.Context, req *grpc.NamespaceRequest) (*grpc.NamespaceResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetName())
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespaceResponse
	resp.Namespace = bytedata.ConvertNamespaceToGrpc(ns)

	return &resp, nil
}

func (flow *flow) Namespaces(ctx context.Context, req *grpc.NamespacesRequest) (*grpc.NamespacesResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	list, err := tx.DataStore().Namespaces().GetAll(ctx)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.NamespacesResponse)
	resp.PageInfo = nil

	resp.Results = bytedata.ConvertNamespacesListToGrpc(list)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (flow *flow) NamespacesStream(req *grpc.NamespacesRequest, srv grpc.Flow_NamespacesStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	ctx := srv.Context()

	resp, err := flow.Namespaces(ctx, req)
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

func (flow *flow) CreateNamespace(ctx context.Context, req *grpc.CreateNamespaceRequest) (*grpc.CreateNamespaceResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		flow.sugar.Warnf("CreateNamespace failed to begin database transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().Create(ctx, &database.Namespace{
		Name: req.GetName(),
	})
	if err != nil {
		flow.sugar.Warnf("CreateNamespace failed to create namespace: %v", err)
		return nil, err
	}

	_, err = tx.FileStore().CreateRoot(ctx, uuid.New(), ns.Name)
	if err != nil {
		flow.sugar.Warnf("CreateNamespace failed to create file-system root: %v", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		flow.sugar.Warnf("CreateNamespace failed to commit database transaction: %v", err)
		return nil, err
	}

	flow.sugar.Infof("Created namespace '%s'.", ns.Name)
	flow.logger.Infof(ctx, flow.ID, flow.GetAttributes(), "Created namespace '%s'.", ns.Name)
	flow.pubsub.NotifyNamespaces()

	var resp grpc.CreateNamespaceResponse
	resp.Namespace = bytedata.ConvertNamespaceToGrpc(ns)

	err = flow.pBus.Publish(pubsub2.NamespaceCreate, ns.Name)
	if err != nil {
		flow.sugar.Error("pubsub publish", "error", err)
	}

	return &resp, nil
}

func (flow *flow) DeleteNamespace(ctx context.Context, req *grpc.DeleteNamespaceRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	var resp emptypb.Empty

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetName())
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	err = tx.DataStore().Namespaces().Delete(ctx, ns.Name)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, flow.ID, flow.GetAttributes(), "Deleted namespace '%s'.", ns.Name)
	flow.pubsub.NotifyNamespaces()
	flow.pubsub.CloseNamespace(ns)

	// delete all knative services
	// TODO: yassir, delete knative services here.

	err = flow.pBus.Publish(pubsub2.NamespaceDelete, ns.Name)
	if err != nil {
		flow.sugar.Error("pubsub publish", "error", err)
	}

	return &resp, err
}
