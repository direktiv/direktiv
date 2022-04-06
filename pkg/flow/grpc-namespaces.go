package flow

import (
	"context"
	"errors"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

func namespaceOrder(p *pagination) []ent.NamespacePaginateOption {

	var opts []ent.NamespacePaginateOption

	for _, o := range p.order {

		if o == nil {
			continue
		}

		field := ent.NamespaceOrderFieldName
		direction := ent.OrderDirectionAsc

		if x := o.Field; x != "" && x == "NAME" {
			field = ent.NamespaceOrderFieldName
		}

		if x := o.Direction; x != "" && x == "DESC" {
			direction = ent.OrderDirectionDesc
		}

		opts = append(opts, ent.WithNamespaceOrder(&ent.NamespaceOrder{
			Direction: direction,
			Field:     field,
		}))
	}

	if len(opts) == 0 {
		opts = append(opts, ent.WithNamespaceOrder(&ent.NamespaceOrder{
			Direction: ent.OrderDirectionAsc,
			Field:     ent.NamespaceOrderFieldName,
		}))
	}

	return opts

}

func namespaceFilter(p *pagination) []ent.NamespacePaginateOption {

	var filters []func(query *ent.NamespaceQuery) (*ent.NamespaceQuery, error)
	var opts []ent.NamespacePaginateOption

	if p.filter == nil {
		return nil
	}

	for i := range p.filter {

		f := p.filter[i]

		if f == nil {
			continue
		}

		filter := f.Val

		filters = append(filters, func(query *ent.NamespaceQuery) (*ent.NamespaceQuery, error) {

			if filter == "" {
				return query, nil
			}

			field := f.Field
			if field == "" {
				return query, nil
			}

			switch field {
			case "NAME":

				ftype := f.Type
				if ftype == "" {
					return query, nil
				}

				switch ftype {
				case "CONTAINS":
					return query.Where(entns.NameContains(filter)), nil
				}
			}

			return query, nil

		})

	}

	if len(filters) > 0 {
		opts = append(opts, ent.WithNamespaceFilter(func(query *ent.NamespaceQuery) (*ent.NamespaceQuery, error) {
			var err error
			for _, filter := range filters {
				query, err = filter(query)
				if err != nil {
					return nil, err
				}
			}
			return query, nil
		}))
	}

	return opts

}

func (flow *flow) ResolveNamespaceUID(ctx context.Context, req *grpc.ResolveNamespaceUIDRequest) (*grpc.NamespaceResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	nsc := flow.db.Namespace

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, err
	}

	ns, err := nsc.Get(ctx, id)
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

	nsc := flow.db.Namespace
	ns, err := flow.getNamespace(ctx, nsc, req.GetName())
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

	_, err = nsc.UpdateOneID(ns.ID).SetConfig(newCfgData).Save(ctx)
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

	nsc := flow.db.Namespace
	ns, err := flow.getNamespace(ctx, nsc, req.GetName())
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

	nsc := flow.db.Namespace
	ns, err := flow.getNamespace(ctx, nsc, req.GetName())
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

func (flow *flow) Namespaces(ctx context.Context, req *grpc.NamespacesRequest) (*grpc.NamespacesResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.NamespacePaginateOption{}
	opts = append(opts, namespaceOrder(p)...)
	opts = append(opts, namespaceFilter(p)...)

	nsc := flow.db.Namespace
	query := nsc.Query()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespacesResponse

	err = atob(cx, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}

func (flow *flow) NamespacesStream(req *grpc.NamespacesRequest, srv grpc.Flow_NamespacesStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	opts := []ent.NamespacePaginateOption{}
	opts = append(opts, namespaceOrder(p)...)
	opts = append(opts, namespaceFilter(p)...)

	sub := flow.pubsub.SubscribeNamespaces()
	defer flow.cleanup(sub.Close)

	nsc := flow.db.Namespace

resend:

	query := nsc.Query()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	resp := new(grpc.NamespacesResponse)

	err = atob(cx, &resp)
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

	nsc := tx.Namespace
	inoc := tx.Inode
	var ns *ent.Namespace

	if req.GetIdempotent() {
		ns, err = flow.getNamespace(ctx, nsc, req.GetName())
		if err == nil {
			rollback(tx)
			goto respond
		}
		if !IsNotFound(err) {
			return nil, err
		}
	}

	ns, err = nsc.Create().SetName(req.GetName()).Save(ctx)
	if err != nil {
		return nil, err
	}

	_, err = inoc.Create().SetNillableName(nil).SetType("directory").SetNamespace(ns).Save(ctx)
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

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	ns, err := nsc.Query().Where(entns.NameEQ(req.GetName())).Only(ctx)
	if err != nil {
		if IsNotFound(err) && req.GetIdempotent() {
			rollback(tx)
			goto respond
		}
		return nil, err
	}

	if !req.GetRecursive() {
		k, err := ns.QueryInodes().Count(ctx)
		if err != nil {
			return nil, err
		}
		if k != 1 { // root dir
			return nil, errors.New("refusing to delete non-empty namespace without explicit recursive argument")
		}
		// TODO: don't delete if namespace has stuff unless 'recursive' explicitly requested
	}

	err = nsc.DeleteOne(ns).Exec(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.deleteNamespaceSecrets(ns)

	flow.logToServer(ctx, time.Now(), "Deleted namespace '%s'.", ns.Name)
	flow.pubsub.NotifyNamespaces()
	flow.pubsub.CloseNamespace(ns)

respond:

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) RenameNamespace(ctx context.Context, req *grpc.RenameNamespaceRequest) (*grpc.RenameNamespaceResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	ns, err := nsc.Query().Where(entns.NameEQ(req.GetOld())).Only(ctx)
	if err != nil {
		return nil, err
	}

	ns, err = ns.Update().SetName(req.GetNew()).Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToServer(ctx, time.Now(), "Renamed namespace from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.logToNamespace(ctx, time.Now(), ns, "Renamed namespace from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyNamespaces()
	flow.pubsub.CloseNamespace(ns)

	var resp grpc.RenameNamespaceResponse

	err = atob(ns, &resp.Namespace)
	if err != nil {
		return nil, err
	}

	return &resp, nil

}

// TODO: translate ent errors for grpc
// TODO: validate filters
// TODO: validate orderings
// TODO: validate other request fields
