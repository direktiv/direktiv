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

	"github.com/vorteil/direktiv/pkg/flow/ent"
	entino "github.com/vorteil/direktiv/pkg/flow/ent/inode"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func directoryOrder(p *pagination) ent.InodePaginateOption {

	field := ent.InodeOrderFieldName
	direction := ent.OrderDirectionAsc

	if p.order != nil {

		if x := p.order.Field; x != "" && x == "NAME" {
			field = ent.InodeOrderFieldName
		}

		if x := p.order.Direction; x != "" && x == "DESC" {
			direction = ent.OrderDirectionDesc
		}

	}

	return ent.WithInodeOrder(&ent.InodeOrder{
		Direction: direction,
		Field:     field,
	})

}

func directoryFilter(p *pagination) ent.InodePaginateOption {

	if p.filter == nil {
		return nil
	}

	filter := p.filter.Val

	return ent.WithInodeFilter(func(query *ent.InodeQuery) (*ent.InodeQuery, error) {

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
				return query.Where(entino.NameContains(filter)), nil
			}
		}

		return query, nil

	})

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

	return &resp, nil

}

func (flow *flow) Directory(ctx context.Context, req *grpc.DirectoryRequest) (*grpc.DirectoryResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.InodePaginateOption{}
	opts = append(opts, directoryOrder(p))
	filter := directoryFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

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
	opts = append(opts, directoryOrder(p))
	filter := directoryFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

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

	path := getInodePath(req.GetPath())
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

	resp.Namespace = namespace
	resp.Node.Parent = dir
	resp.Node.Path = path

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

	if !req.GetRecursive() {
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

	path := getInodePath(req.GetNew())
	if path == "/" {
		return nil, errors.New("cannot overwrite root node")
	}

	if strings.Contains(path, d.path) {
		return nil, errors.New("cannot move node into itself")
	}

	dir, base := filepath.Split(path)

	ino := d.ino

	pd, err := flow.getInode(ctx, nil, d.ns(), dir, false)
	if err != nil {
		return nil, err
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
	flow.pubsub.NotifyInode(pd.ino)
	flow.pubsub.CloseInode(d.ino)

	var resp grpc.RenameNodeResponse

	err = atob(ino, &resp.Node)
	if err != nil {
		return nil, err
	}

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
