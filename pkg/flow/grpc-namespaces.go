package flow

import (
	"context"
	"errors"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entino "github.com/direktiv/direktiv/pkg/flow/ent/inode"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/functions"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
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

	Namespace := flow.db.Namespace

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, err
	}

	ns, err := Namespace.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespaceResponse

	err = atob(ns, &resp.Namespace)
	if err != nil {
		return nil, err
	}

	resp.Namespace.Oid = ns.ID.String()

	return &resp, nil

}

func (flow *flow) SetNamespaceConfig(ctx context.Context, req *grpc.SetNamespaceConfigRequest) (*grpc.SetNamespaceConfigResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	Namespace := flow.db.Namespace

	cached := new(CacheData)

	err := flow.database.NamespaceByName(ctx, nil, cached, req.GetName())
	if err != nil {
		return nil, err
	}

	patchCfg, err := loadNSConfig([]byte(req.Config))
	if err != nil {
		return nil, err
	}

	var newCfgData string

	data, err := patchCfg.mergeIntoNamespaceConfig([]byte(cached.Namespace.Config))
	if err != nil {
		return nil, err
	}
	newCfgData = string(data)

	_, err = Namespace.UpdateOneID(cached.Namespace.ID).SetConfig(newCfgData).Save(ctx)
	if err != nil {
		return nil, err
	}

	var resp grpc.SetNamespaceConfigResponse
	resp.Config = newCfgData
	resp.Name = cached.Namespace.Name

	return &resp, nil
}

func (flow *flow) GetNamespaceConfig(ctx context.Context, req *grpc.GetNamespaceConfigRequest) (*grpc.GetNamespaceConfigResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(CacheData)

	err := flow.database.NamespaceByName(ctx, nil, cached, req.GetName())
	if err != nil {
		return nil, err
	}

	var resp grpc.GetNamespaceConfigResponse
	resp.Config = cached.Namespace.Config
	resp.Name = cached.Namespace.Name

	return &resp, nil
}

func (flow *flow) Namespace(ctx context.Context, req *grpc.NamespaceRequest) (*grpc.NamespaceResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(CacheData)

	err := flow.database.NamespaceByName(ctx, nil, cached, req.GetName())
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespaceResponse

	err = atob(cached.Namespace, &resp.Namespace)
	if err != nil {
		return nil, err
	}

	resp.Namespace.Oid = cached.Namespace.ID.String()

	return &resp, nil

}

func (flow *flow) Namespaces(ctx context.Context, req *grpc.NamespacesRequest) (*grpc.NamespacesResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	query := flow.db.Namespace.Query()

	results, pi, err := paginate[*ent.NamespaceQuery, *ent.Namespace](ctx, req.Pagination, query, namespacesOrderings, namespacesFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.NamespacesResponse)
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
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

	query := flow.db.Namespace.Query()

	results, pi, err := paginate[*ent.NamespaceQuery, *ent.Namespace](ctx, req.Pagination, query, namespacesOrderings, namespacesFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.NamespacesResponse)
	resp.PageInfo = pi

	err = atob(results, &resp.Results)
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

func (flow *flow) CreateNamespace(ctx context.Context, req *grpc.CreateNamespaceRequest) (*grpc.CreateNamespaceResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	var ns *Namespace
	var x *ent.Namespace

	cached := new(CacheData)

	if req.GetIdempotent() {

		err = flow.database.NamespaceByName(ctx, tx, cached, req.GetName())
		if err == nil {
			rollback(tx)
			goto respond
		}
		if !derrors.IsNotFound(err) {
			return nil, err
		}
	}

	x, err = tx.Namespace.Create().SetName(req.GetName()).Save(ctx)
	if err != nil {
		return nil, err
	}
	cached.Namespace.Name = x.Name

	_, err = tx.Inode.Create().SetNillableName(nil).SetType(util.InodeTypeDirectory).SetNamespaceID(ns.ID).Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToServer(ctx, time.Now(), "Created namespace '%s'.", ns.Name)
	flow.pubsub.NotifyNamespaces()

respond:

	var resp grpc.CreateNamespaceResponse

	err = atob(ns, &resp.Namespace)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}

func (flow *flow) DeleteNamespace(ctx context.Context, req *grpc.DeleteNamespaceRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())
	var resp emptypb.Empty

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached := new(CacheData)

	err = flow.database.NamespaceByName(ctx, tx, cached, req.GetName())
	if err != nil {
		if derrors.IsNotFound(err) && req.GetIdempotent() {
			rollback(tx)
			return &resp, nil
		}
		return nil, err
	}

	if !req.GetRecursive() {
		k, err := tx.Inode.Query().Where(entino.HasNamespaceWith(entns.ID(cached.Namespace.ID))).Count(ctx)
		if err != nil {
			return nil, err
		}
		if k != 1 { // root dir
			return nil, errors.New("refusing to delete non-empty namespace without explicit recursive argument")
		}
	}

	err = tx.Namespace.DeleteOneID(cached.Namespace.ID).Exec(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.deleteNamespaceSecrets(cached.Namespace)

	flow.logToServer(ctx, time.Now(), "Deleted namespace '%s'.", cached.Namespace.Name)
	flow.pubsub.NotifyNamespaces()
	flow.pubsub.CloseNamespace(cached.Namespace)

	// delete all knative services
	annotations := make(map[string]string)
	annotations[functions.ServiceHeaderNamespaceName] = req.Name
	lfr := igrpc.ListFunctionsRequest{
		Annotations: annotations,
	}
	_, err = flow.actions.client.DeleteFunctions(ctx, &lfr)

	return &resp, err

}

func (flow *flow) RenameNamespace(ctx context.Context, req *grpc.RenameNamespaceRequest) (*grpc.RenameNamespaceResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached := new(CacheData)
	err = flow.database.NamespaceByName(ctx, tx, cached, req.GetOld())
	if err != nil {
		return nil, err
	}

	x, err := tx.Namespace.UpdateOneID(cached.Namespace.ID).SetName(req.GetNew()).Save(ctx)
	if err != nil {
		return nil, err
	}

	cached.Namespace.Name = x.Name

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToServer(ctx, time.Now(), "Renamed namespace from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.logToNamespace(ctx, time.Now(), cached, "Renamed namespace from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyNamespaces()
	flow.pubsub.CloseNamespace(cached.Namespace)

	var resp grpc.RenameNamespaceResponse

	err = atob(cached.Namespace, &resp.Namespace)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}
