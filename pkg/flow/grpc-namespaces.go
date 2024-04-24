package flow

import (
	"context"
	"log/slog"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	pubsub2 "github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) Namespace(ctx context.Context, req *grpc.NamespaceRequest) (*grpc.NamespaceResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	tx, err := flow.beginSQLTx(ctx)
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
	slog.Debug("Handling gRPC request", "this", this())

	tx, err := flow.beginSQLTx(ctx)
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

	return resp, nil
}

func (flow *flow) NamespacesStream(req *grpc.NamespacesRequest, srv grpc.Flow_NamespacesStreamServer) error {
	slog.Debug("Handling gRPC request", "this", this())
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
	slog.Debug("Handling gRPC request", "this", this())

	tx, err := flow.beginSQLTx(ctx)
	if err != nil {
		slog.Warn("Creating a Namespace failed to begin database transaction", "error", err)
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().Create(ctx, &datastore.Namespace{
		Name: req.GetName(),
	})
	if err != nil {
		slog.Warn("Creating a Namespace failed to create namespace", "error", err)
		return nil, err
	}

	_, err = tx.FileStore().CreateRoot(ctx, uuid.New(), ns.Name)
	if err != nil {
		slog.Warn("Creating a Namespace failed to create file-system root", "error", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		slog.Warn("Creating a Namespace failed to commit database transaction", "error", err)
		return nil, err
	}

	slog.Debug("Created namespace", "namespace", ns.Name)
	flow.pubsub.NotifyNamespaces()

	var resp grpc.CreateNamespaceResponse
	resp.Namespace = bytedata.ConvertNamespaceToGrpc(ns)

	err = flow.pBus.Publish(pubsub2.NamespaceCreate, ns.Name)
	if err != nil {
		slog.Error("pubsub publish", "error", err)
	}

	return &resp, nil
}

func (flow *flow) DeleteNamespace(ctx context.Context, req *grpc.DeleteNamespaceRequest) (*emptypb.Empty, error) {
	slog.Debug("Handling gRPC request", "this", this())
	var resp emptypb.Empty

	tx, err := flow.beginSQLTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetName())
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

	slog.Debug("Deleted namespace.", "namespace", ns.Name)
	flow.pubsub.NotifyNamespaces()
	flow.pubsub.CloseNamespace(ns)

	// delete all knative services
	// TODO: yassir, delete knative services here.

	err = flow.pBus.Publish(pubsub2.NamespaceDelete, ns.Name)
	if err != nil {
		slog.Error("pubsub publish", "error", err)
	}

	return &resp, err
}
