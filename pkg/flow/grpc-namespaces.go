package flow

import (
	"context"
	"errors"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/functions"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

var namespacesOrderings = []*orderingInfo{
	{
		db:           entns.FieldName,
		req:          util.PaginationKeyName,
		defaultOrder: ent.Asc,
	},
}

var namespacesFilters = map[*filteringInfo]func(query *ent.NamespaceQuery, v string) (*ent.NamespaceQuery, error){
	{
		field: util.PaginationKeyName,
		ftype: "CONTAINS",
	}: func(query *ent.NamespaceQuery, v string) (*ent.NamespaceQuery, error) {
		return query.Where(entns.NameContains(v)), nil
	},
}

func (flow *flow) ResolveNamespaceUID(ctx context.Context, req *grpc.ResolveNamespaceUIDRequest) (*grpc.NamespaceResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, err
	}

	ns, err := flow.edb.Namespace(ctx, id)
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespaceResponse

	err = bytedata.ConvertDataForOutput(ns, &resp.Namespace)
	if err != nil {
		return nil, err
	}

	resp.Namespace.Oid = ns.ID.String()

	return &resp, nil
}

func (flow *flow) SetNamespaceConfig(ctx context.Context, req *grpc.SetNamespaceConfigRequest) (*grpc.SetNamespaceConfigResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetName())
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

	clients := flow.edb.Clients(ctx)

	_, err = clients.Namespace.UpdateOneID(ns.ID).SetConfig(newCfgData).Save(ctx)
	if err != nil {
		return nil, err
	}

	var resp grpc.SetNamespaceConfigResponse
	resp.Config = newCfgData
	resp.Name = ns.Name

	return &resp, nil
}

func (flow *flow) GetNamespaceConfig(ctx context.Context, req *grpc.GetNamespaceConfigRequest) (*grpc.GetNamespaceConfigResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetName())
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

	ns, err := flow.edb.NamespaceByName(ctx, req.GetName())
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespaceResponse

	err = bytedata.ConvertDataForOutput(ns, &resp.Namespace)
	if err != nil {
		return nil, err
	}

	resp.Namespace.Oid = ns.ID.String()

	return &resp, nil
}

func (flow *flow) Namespaces(ctx context.Context, req *grpc.NamespacesRequest) (*grpc.NamespacesResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	clients := flow.edb.Clients(ctx)

	query := clients.Namespace.Query()

	results, pi, err := paginate[*ent.NamespaceQuery, *ent.Namespace](ctx, req.Pagination, query, namespacesOrderings, namespacesFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.NamespacesResponse)
	resp.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Results)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (flow *flow) NamespacesStream(req *grpc.NamespacesRequest, srv grpc.Flow_NamespacesStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	sub := flow.pubsub.SubscribeNamespaces()
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(ctx)

	query := clients.Namespace.Query()

	results, pi, err := paginate[*ent.NamespaceQuery, *ent.Namespace](ctx, req.Pagination, query, namespacesOrderings, namespacesFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.NamespacesResponse)
	resp.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Results)
	if err != nil {
		return err
	}

	nhash = bytedata.Checksum(resp)
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

func (flow *flow) CreateNamespace(ctx context.Context, req *grpc.CreateNamespaceRequest) (*grpc.CreateNamespaceResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	var ns *database.Namespace
	var x *ent.Namespace

	ctx, tx, err := flow.edb.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	clients := flow.edb.Clients(ctx)

	if req.GetIdempotent() {
		ns, err = flow.edb.NamespaceByName(ctx, req.GetName())
		if err == nil {
			var resp grpc.CreateNamespaceResponse
			err = bytedata.ConvertDataForOutput(ns, &resp.Namespace)
			if err != nil {
				return nil, err
			}

			return &resp, nil
		}
		if !derrors.IsNotFound(err) {
			return nil, err
		}
	}

	x, err = clients.Namespace.Create().SetName(req.GetName()).Save(ctx)
	if err != nil {
		return nil, err
	}

	ns = &database.Namespace{
		ID:        x.ID,
		CreatedAt: x.CreatedAt,
		UpdatedAt: x.UpdatedAt,
		Config:    x.Config,
		Name:      x.Name,
	}

	err = tx.Commit()
	if err != nil {
		flow.logger.Errorf(ctx, flow.ID, flow.GetAttributes(), "Failed to create namespace '%s'.", ns.Name)
		return nil, err
	}

	var txErr error
	err = flow.runSqlTx(ctx, func(fStore filestore.FileStore, store datastore.Store) error {
		var root *filestore.Root
		root, txErr = fStore.CreateRoot(ctx, x.ID)
		if txErr != nil {
			return txErr
		}
		_, _, txErr = fStore.ForRootID(root.ID).CreateFile(ctx, "/", filestore.FileTypeDirectory, nil)
		if txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		clients = flow.edb.Clients(context.Background())
		_ = clients.Namespace.DeleteOneID(ns.ID).Exec(context.Background()) // NOTE: need to find a better way to do this.
		return nil, err
	}

	flow.logger.Infof(ctx, flow.ID, flow.GetAttributes(), "Created namespace '%s'.", ns.Name)
	flow.pubsub.NotifyNamespaces()

	var resp grpc.CreateNamespaceResponse

	err = bytedata.ConvertDataForOutput(ns, &resp.Namespace)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (flow *flow) DeleteNamespace(ctx context.Context, req *grpc.DeleteNamespaceRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	var resp emptypb.Empty

	ns, err := flow.edb.NamespaceByName(ctx, req.GetName())
	if err != nil {
		if derrors.IsNotFound(err) && req.GetIdempotent() {
			return &resp, nil
		}
		return nil, err
	}

	var isEmpty bool
	var txErr error
	err = flow.runSqlTx(ctx, func(fStore filestore.FileStore, store datastore.Store) error {
		isEmpty, txErr = fStore.ForRootID(ns.ID).IsEmptyDirectory(ctx, "/")
		return txErr
	})
	if err != nil {
		if !errors.Is(err, filestore.ErrNotFound) {
			// NOTE: the alternative shouldn't be possible
			return nil, err
		}
	}

	if !req.GetRecursive() && !isEmpty {
		return nil, errors.New("refusing to delete non-empty namespace without explicit recursive argument")
	}

	clients := flow.edb.Clients(ctx)
	err = clients.Namespace.DeleteOneID(ns.ID).Exec(ctx)
	if err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, flow.ID, flow.GetAttributes(), "Deleted namespace '%s'.", ns.Name)
	flow.pubsub.NotifyNamespaces()
	flow.pubsub.CloseNamespace(ns)

	// delete all knative services
	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderNamespaceName] = req.Name
	lfr := igrpc.FunctionsListFunctionsRequest{
		Annotations: annotations,
	}
	_, err = flow.actions.client.DeleteFunctions(ctx, &lfr)

	// delete filter cache
	deleteCacheNamespaceSync(ns.Name)
	flow.server.pubsub.Publish(&pubsub.PubsubUpdate{
		Handler: deleteFilterCacheNamespace,
		Key:     ns.Name,
	})

	return &resp, err
}

func (flow *flow) RenameNamespace(ctx context.Context, req *grpc.RenameNamespaceRequest) (*grpc.RenameNamespaceResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx, tx, err := flow.edb.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	ns, err := flow.edb.NamespaceByName(ctx, req.GetOld())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	x, err := clients.Namespace.UpdateOneID(ns.ID).SetName(req.GetNew()).Save(ctx)
	if err != nil {
		return nil, err
	}

	ns.Name = x.Name

	err = tx.Commit()
	if err != nil {
		flow.logger.Infof(ctx, flow.ID, flow.GetAttributes(), "Could not rename namespace '%s'.", ns.Name)
		return nil, err
	}

	flow.logger.Infof(ctx, flow.ID, flow.GetAttributes(), "Renamed namespace from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.logger.Infof(ctx, ns.ID, database.GetAttributes(recipient.Namespace, ns), "Renamed namespace from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyNamespaces()
	flow.pubsub.CloseNamespace(ns)

	var resp grpc.RenameNamespaceResponse

	err = bytedata.ConvertDataForOutput(ns, &resp.Namespace)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
