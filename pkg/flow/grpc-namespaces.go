package flow

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vorteil/direktiv/pkg/flow/ent"
	entns "github.com/vorteil/direktiv/pkg/flow/ent/namespace"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func namespaceOrder(p *pagination) ent.NamespacePaginateOption {

	field := ent.NamespaceOrderFieldName
	direction := ent.OrderDirectionAsc

	if p.order != nil {

		if x := p.order.Field; x != "" && x == "NAME" {
			field = ent.NamespaceOrderFieldName
		}

		if x := p.order.Direction; x != "" && x == "DESC" {
			direction = ent.OrderDirectionDesc
		}

	}

	return ent.WithNamespaceOrder(&ent.NamespaceOrder{
		Direction: direction,
		Field:     field,
	})

}

func namespaceFilter(p *pagination) ent.NamespacePaginateOption {

	if p.filter == nil {
		return nil
	}

	filter := p.filter.Val

	return ent.WithNamespaceFilter(func(query *ent.NamespaceQuery) (*ent.NamespaceQuery, error) {

		if filter == "" {
			return query, nil
		}

		field := p.filter.Field
		if field == "" {
			return query, nil
		}

		switch field {
		case "NAME":

			ftype := p.filter.Type
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
	opts = append(opts, namespaceOrder(p))
	filter := namespaceFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

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
	opts = append(opts, namespaceOrder(p))
	filter := namespaceFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

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
