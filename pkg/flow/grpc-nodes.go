package flow

import (
	"context"
	"errors"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entino "github.com/direktiv/direktiv/pkg/flow/ent/inode"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func directoryOrder(p *pagination) []ent.InodePaginateOption {

	var opts []ent.InodePaginateOption

	for _, o := range p.order {

		if o == nil {
			continue
		}

		order := ent.InodeOrder{
			Direction: ent.OrderDirectionAsc,
			Field:     ent.InodeOrderFieldName,
		}

		switch o.GetField() {
		case "UPDATED":
			order.Field = ent.InodeOrderFieldUpdatedAt
		case "CREATED":
			order.Field = ent.InodeOrderFieldCreatedAt
		case "NAME":
			order.Field = ent.InodeOrderFieldName
		case "TYPE":
			order.Field = ent.InodeOrderFieldType
		default:
			break
		}

		switch o.GetDirection() {
		case "DESC":
			order.Direction = ent.OrderDirectionDesc
		case "ASC":
			order.Direction = ent.OrderDirectionAsc
		default:
			break
		}

		opts = append(opts, ent.WithInodeOrder(&order))

	}

	if len(opts) == 0 {
		opts = append(opts, ent.WithInodeOrder(&ent.InodeOrder{
			Direction: ent.OrderDirectionDesc,
			Field:     ent.InodeOrderFieldType,
		}), ent.WithInodeOrder(&ent.InodeOrder{
			Direction: ent.OrderDirectionAsc,
			Field:     ent.InodeOrderFieldName,
		}))
	}

	return opts

}

func directoryFilter(p *pagination) []ent.InodePaginateOption {

	var filters []func(query *ent.InodeQuery) (*ent.InodeQuery, error)
	var opts []ent.InodePaginateOption

	if p.filter == nil {
		return nil
	}

	for i := range p.filter {

		f := p.filter[i]

		if f == nil {
			continue
		}

		filter := f.Val

		filters = append(filters, func(query *ent.InodeQuery) (*ent.InodeQuery, error) {

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
					return query.Where(entino.NameContains(filter)), nil
				}
			}

			return query, nil

		})

	}

	if len(filters) > 0 {
		opts = append(opts, ent.WithInodeFilter(func(query *ent.InodeQuery) (*ent.InodeQuery, error) {
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

func (flow *flow) Node(ctx context.Context, req *grpc.NodeRequest) (*grpc.NodeResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	var err error
	var resp grpc.NodeResponse

	nsc := flow.db.Namespace
	d, err := flow.traverseToInode(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	resp.Namespace = d.namespace()
	resp.Node.Path = d.path
	resp.Node.Parent = d.dir

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Node.ReadOnly = d.ro

	return &resp, nil

}

func (flow *flow) Directory(ctx context.Context, req *grpc.DirectoryRequest) (*grpc.DirectoryResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.InodePaginateOption{}
	opts = append(opts, directoryOrder(p)...)
	opts = append(opts, directoryFilter(p)...)

	nsc := flow.db.Namespace
	d, err := flow.traverseToInode(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	if d.ino.Type != "directory" {
		return nil, ErrNotDir
	}

	query := d.ino.QueryChildren()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.DirectoryResponse

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Node.ReadOnly = d.ro

	resp.Namespace = d.namespace()
	resp.Node.Path = d.path
	resp.Node.Parent = d.dir

	err = atob(cx, &resp.Children)
	if err != nil {
		return nil, err
	}

	for idx := range resp.Children.Edges {
		child := resp.Children.Edges[idx]
		child.Node.Parent = resp.Node.Path
		child.Node.Path = filepath.Join(resp.Node.Path, child.Node.Name)

		if child.Node.ExpandedType == "" {
			child.Node.ExpandedType = child.Node.Type
		}

		child.Node.ReadOnly = d.ro
		if child.Node.ExpandedType == util.InodeTypeGit {
			mir, err := cx.Edges[idx].Node.Mirror(ctx)
			if err == nil {
				child.Node.ReadOnly = !mir.Locked
			}
		}
	}

	return &resp, nil

}

func (flow *flow) DirectoryStream(req *grpc.DirectoryRequest, srv grpc.Flow_DirectoryStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	opts := []ent.InodePaginateOption{}
	opts = append(opts, directoryOrder(p)...)
	opts = append(opts, directoryFilter(p)...)

	nsc := flow.db.Namespace
	d, err := flow.traverseToInode(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	if d.ino.Type != "directory" {
		return ErrNotDir
	}

	sub := flow.pubsub.SubscribeInode(d.ino)
	defer flow.cleanup(sub.Close)

resend:

	query := d.ino.QueryChildren()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	resp := new(grpc.DirectoryResponse)

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Node.ReadOnly = d.ro

	resp.Node.Path = d.path
	resp.Node.Parent = d.dir

	err = atob(cx, &resp.Children)
	if err != nil {
		return err
	}

	for idx := range resp.Children.Edges {
		child := resp.Children.Edges[idx]
		child.Node.Parent = resp.Node.Path
		child.Node.Path = filepath.Join(resp.Node.Path, child.Node.Name)

		if child.Node.ExpandedType == "" {
			child.Node.ExpandedType = child.Node.Type
		}

		child.Node.ReadOnly = d.ro
		if child.Node.ExpandedType == util.InodeTypeGit {
			mir, err := cx.Edges[idx].Node.Mirror(ctx)
			if err == nil {
				child.Node.ReadOnly = !mir.Locked
			}
		}
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

func (flow *flow) CreateDirectory(ctx context.Context, req *grpc.CreateDirectoryRequest) (*grpc.CreateDirectoryResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	namespace := req.GetNamespace()
	ns, err := flow.getNamespace(ctx, tx.Namespace, namespace)
	if err != nil {
		return nil, err
	}

	path := GetInodePath(req.GetPath())
	dir, base := filepath.Split(path)

	if base == "" || base == "/" {
		return nil, status.Error(codes.AlreadyExists, "root directory already exists")
	}

	inoc := tx.Inode

	pino, err := flow.getInode(ctx, inoc, ns, dir, req.GetParents())
	if err != nil {
		return nil, err
	}

	if pino.ino.Type != "directory" {
		return nil, status.Error(codes.AlreadyExists, "parent node is not a directory")
	}

	if pino.ro {
		return nil, errors.New("cannot write into read-only directory")
	}

	ino, err := inoc.Create().SetName(base).SetNamespace(ns).SetParent(pino.ino).SetType("directory").Save(ctx)
	if err != nil {

		if ent.IsConstraintError(err) && req.GetIdempotent() {
			var d *nodeData
			var e error

			ns, err = flow.getNamespace(ctx, flow.db.Namespace, namespace)
			if err != nil {
				return nil, err
			}

			inoc = flow.db.Inode

			d, e = flow.getInode(ctx, inoc, ns, req.GetPath(), false)
			if e != nil {
				return nil, err
			}

			if d.ino.Type != "directory" {
				return nil, err
			}

			ino = d.ino

			rollback(tx)
			goto respond

		}

		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToNamespace(ctx, time.Now(), ns, "Created directory '%s'.", path)
	flow.pubsub.NotifyInode(pino.ino)

respond:

	var resp grpc.CreateDirectoryResponse

	err = atob(ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Node.ReadOnly = false

	resp.Namespace = namespace
	resp.Node.Parent = dir
	resp.Node.Path = path

	// Broadcast
	err = flow.BroadcastDirectory(ctx, BroadcastEventTypeCreate,
		broadcastDirectoryInput{
			Path:   resp.Node.Path,
			Parent: resp.Node.Parent,
		}, ns)

	if err != nil {
		return nil, err
	}

	return &resp, nil

}

func (flow *flow) DeleteNode(ctx context.Context, req *grpc.DeleteNodeRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	inoc := tx.Inode

	d, err := flow.traverseToInode(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		if IsNotFound(err) && req.GetIdempotent() {
			rollback(tx)
			goto respond
		}
		return nil, err
	}

	if d.path == "/" {
		return nil, status.Error(codes.InvalidArgument, "cannot delete root node")
	}

	if d.ro && d.ino.ExtendedType != util.InodeTypeGit {
		return nil, status.Error(codes.InvalidArgument, "cannot delete contents of read-only directory")
	}

	if !req.GetRecursive() && d.ino.Type == "directory" {
		k, err := d.ino.QueryChildren().Count(ctx)
		if err != nil {
			return nil, err
		}
		if k != 0 {
			return nil, status.Error(codes.InvalidArgument, "refusing to delete non-empty directory without explicit recursive argument")
		}
		// TODO: don't delete if directory has stuff unless 'recursive' explicitly requested
	}

	err = inoc.DeleteOne(d.ino).Exec(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if d.ino.Type == "workflow" {
		metricsWf.WithLabelValues(d.ns().Name, d.ns().Name).Dec()
		metricsWfUpdated.WithLabelValues(d.ns().Name, d.path, d.ns().Name).Inc()

		// Broadcast Event
		err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeDelete,
			broadcastWorkflowInput{
				Name:   d.base,
				Path:   d.path,
				Parent: d.dir,
				Live:   false,
			}, d.ns())

		if err != nil {
			return nil, err
		}
	} else {
		// Broadcast Event
		err = flow.BroadcastDirectory(ctx, BroadcastEventTypeDelete,
			broadcastDirectoryInput{
				Path:   d.path,
				Parent: d.dir,
			}, d.ns())

		if err != nil {
			return nil, err
		}

	}

	flow.logToNamespace(ctx, time.Now(), d.ns(), "Deleted %s '%s'.", d.ino.Type, d.path)
	flow.pubsub.NotifyInode(d.ino.Edges.Parent)
	flow.pubsub.CloseInode(d.ino)

respond:

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) RenameNode(ctx context.Context, req *grpc.RenameNodeRequest) (*grpc.RenameNodeResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToInode(ctx, nsc, req.GetNamespace(), req.GetOld())
	if err != nil {
		return nil, err
	}

	if d.path == "/" {
		return nil, errors.New("cannot rename root node")
	}

	path := GetInodePath(req.GetNew())
	if path == "/" {
		return nil, errors.New("cannot overwrite root node")
	}

	if strings.Contains(path, d.path+"/") {
		return nil, errors.New("cannot move node into itself")
	}

	if d.ro && d.ino.ExtendedType != util.InodeTypeGit {
		return nil, errors.New("cannot move contents of read-only directory")
	}

	oldpd, err := flow.getInode(ctx, nil, d.ns(), d.dir, false)
	if err != nil {
		return nil, err
	}

	dir, base := filepath.Split(path)

	ino := d.ino

	pd, err := flow.getInode(ctx, nil, d.ns(), dir, false)
	if err != nil {
		return nil, err
	}

	if pd.ro {
		return nil, errors.New("cannot write into read-only directory")
	}

	ino, err = ino.Update().SetName(base).SetParent(pd.ino).Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToNamespace(ctx, time.Now(), d.ns(), "Renamed %s from '%s' to '%s'.", d.ino.Type, req.GetOld(), req.GetNew())
	flow.pubsub.NotifyInode(oldpd.ino)
	flow.pubsub.NotifyInode(pd.ino)
	flow.pubsub.CloseInode(d.ino)

	var resp grpc.RenameNodeResponse

	err = atob(ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Node.ReadOnly = false

	resp.Namespace = d.namespace()
	resp.Node.Parent = dir
	resp.Node.Path = path

	return &resp, nil

}

func (flow *flow) CreateNodeAttributes(ctx context.Context, req *grpc.CreateNodeAttributesRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace

	d, err := flow.traverseToInode(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	m := make(map[string]bool)

	for _, attr := range d.ino.Attributes {
		m[attr] = true
	}

	for _, attr := range req.GetAttributes() {
		m[attr] = true
	}

	var attrs []string

	for attr := range m {
		attrs = append(attrs, attr)
	}

	sort.Strings(attrs)

	edges := d.ino.Edges

	d.ino, err = d.ino.Update().SetAttributes(attrs).Save(ctx)
	if err != nil {
		return nil, err
	}

	d.ino.Edges = edges

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) DeleteNodeAttributes(ctx context.Context, req *grpc.DeleteNodeAttributesRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace

	d, err := flow.traverseToInode(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	m := make(map[string]bool)

	for _, attr := range d.ino.Attributes {
		m[attr] = true
	}

	for _, attr := range req.GetAttributes() {
		delete(m, attr)
	}

	var attrs []string

	for attr := range m {
		attrs = append(attrs, attr)
	}

	sort.Strings(attrs)

	edges := d.ino.Edges

	d.ino, err = d.ino.Update().SetAttributes(attrs).Save(ctx)
	if err != nil {
		return nil, err
	}

	d.ino.Edges = edges

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil

}
