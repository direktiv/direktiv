package flow

import (
	"context"
	"errors"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
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

func (flow *flow) SetNamespaceConfig(ctx context.Context, req *grpc.SetNamespaceConfigRequest) (*grpc.SetNamespaceConfigResponse, error) {
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

	patchCfg, err := loadNSConfig([]byte(req.Config))
	if err != nil {
		return nil, err
	}

	var newCfgData string

	data, err := patchCfg.mergeIntoNamespaceConfig([]byte(ns.Config))
	if err != nil {
		return nil, err
	}
	newCfgData = string(data)

	ns.Config = newCfgData
	ns, err = tx.DataStore().Namespaces().Update(ctx, ns)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	var resp grpc.SetNamespaceConfigResponse
	resp.Config = newCfgData
	resp.Name = ns.Name

	return &resp, nil
}

func (flow *flow) GetNamespaceConfig(ctx context.Context, req *grpc.GetNamespaceConfigRequest) (*grpc.GetNamespaceConfigResponse, error) {
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

	var resp grpc.GetNamespaceConfigResponse
	resp.Config = ns.Config
	resp.Name = ns.Name

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

	ns, err := tx.DataStore().Namespaces().Create(ctx, &core.Namespace{
		Name:   req.GetName(),
		Config: core.DefaultNamespaceConfig,
	})
	if err != nil {
		flow.sugar.Warnf("CreateNamespace failed to create namespace: %v", err)
		return nil, err
	}

	root, err := tx.FileStore().CreateRoot(ctx, uuid.New(), ns.Name)
	if err != nil {
		flow.sugar.Warnf("CreateNamespace failed to create file-system root: %v", err)
		return nil, err
	}
	_, err = tx.FileStore().ForRootID(root.ID).CreateFile(ctx, "/", filestore.FileTypeDirectory, "", nil)
	if err != nil {
		flow.sugar.Warnf("CreateNamespace failed to create root directory: %v", err)
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

	isEmpty, err := tx.FileStore().ForNamespace(ns.Name).IsEmptyDirectory(ctx, "/")
	if err != nil {
		if !errors.Is(err, filestore.ErrNotFound) {
			// NOTE: the alternative shouldn't be possible
			return nil, err
		}
	}
	if !req.GetRecursive() && !isEmpty {
		return nil, errors.New("refusing to delete non-empty namespace without explicit recursive argument")
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

func (flow *flow) RenameNamespace(ctx context.Context, req *grpc.RenameNamespaceRequest) (*grpc.RenameNamespaceResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetOld())
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	ns.Name = req.GetNew()
	ns, err = tx.DataStore().Namespaces().Update(ctx, ns)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, flow.ID, flow.GetAttributes(), "Renamed namespace from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.logger.Infof(ctx, ns.ID, database.GetAttributes(recipient.Namespace, ns), "Renamed namespace from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyNamespaces()
	flow.pubsub.CloseNamespace(ns)

	var resp grpc.RenameNamespaceResponse
	resp.Namespace = bytedata.ConvertNamespaceToGrpc(ns)

	return &resp, nil
}
