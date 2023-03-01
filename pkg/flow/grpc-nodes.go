package flow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entino "github.com/direktiv/direktiv/pkg/flow/ent/inode"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

var inodesOrderings = []*orderingInfo{
	{
		db:           entino.FieldType,
		req:          "TYPE",
		defaultOrder: ent.Desc,
		isDefault:    true,
	},
	{
		db:           entino.FieldName,
		req:          util.PaginationKeyName,
		defaultOrder: ent.Asc,
		isDefault:    true,
	},
	{
		db:           entino.FieldCreatedAt,
		req:          "CREATED",
		defaultOrder: ent.Asc,
	},
	{
		db:           entino.FieldUpdatedAt,
		req:          "UPDATED",
		defaultOrder: ent.Asc,
	},
}

var inodesFilters = map[*filteringInfo]func(query *ent.InodeQuery, v string) (*ent.InodeQuery, error){
	{
		field: util.PaginationKeyName,
		ftype: "CONTAINS",
	}: func(query *ent.InodeQuery, v string) (*ent.InodeQuery, error) {
		return query.Where(entino.NameContains(v)), nil
	},
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

	return &resp, nil

}

func (flow *flow) Directory(ctx context.Context, req *grpc.DirectoryRequest) (*grpc.DirectoryResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	d, err := flow.traverseToInode(ctx, flow.db.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	if d.ino.Type != util.InodeTypeDirectory {
		return nil, ErrNotDir
	}

	query := d.ino.QueryChildren()

	results, pi, err := paginate[*ent.InodeQuery, *ent.Inode](ctx, req.Pagination, query, inodesOrderings, inodesFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.DirectoryResponse)
	resp.Namespace = d.namespace()
	resp.Children = new(grpc.DirectoryChildren)
	resp.Children.PageInfo = pi

	err = atob(results, &resp.Children.Results)
	if err != nil {
		return nil, err
	}

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Node.Path = d.path
	resp.Node.Parent = d.dir

	for idx := range resp.Children.Results {
		child := resp.Children.Results[idx]
		child.Parent = resp.Node.Path
		child.Path = filepath.Join(resp.Node.Path, child.Name)

		if child.ExpandedType == "" {
			child.ExpandedType = child.Type
		}

	}

	return resp, nil

}

func (flow *flow) DirectoryStream(req *grpc.DirectoryRequest, srv grpc.Flow_DirectoryStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	d, err := flow.traverseToInode(ctx, flow.db.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	if d.ino.Type != util.InodeTypeDirectory {
		return ErrNotDir
	}

	sub := flow.pubsub.SubscribeInode(d.ino)
	defer flow.cleanup(sub.Close)

resend:

	query := d.ino.QueryChildren()

	results, pi, err := paginate[*ent.InodeQuery, *ent.Inode](ctx, req.Pagination, query, inodesOrderings, inodesFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.DirectoryResponse)
	resp.Namespace = d.namespace()
	resp.Children = new(grpc.DirectoryChildren)
	resp.Children.PageInfo = pi

	err = atob(results, &resp.Children.Results)
	if err != nil {
		return err
	}

	err = atob(d.ino, &resp.Node)
	if err != nil {
		return err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Node.Path = d.path
	resp.Node.Parent = d.dir

	for idx := range resp.Children.Results {
		child := resp.Children.Results[idx]
		child.Parent = resp.Node.Path
		child.Path = filepath.Join(resp.Node.Path, child.Name)

		if child.ExpandedType == "" {
			child.ExpandedType = child.Type
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

type lookupInodeFromParentArgs struct {
	pino *ent.Inode
	name string
}

func (flow *flow) lookupInodeFromParent(ctx context.Context, args *lookupInodeFromParentArgs) (*ent.Inode, error) {

	ino, err := args.pino.QueryChildren().Where(entino.NameEQ(args.name)).Only(ctx)
	if err != nil {
		return nil, err
	}

	return ino, nil

}

type createDirectoryArgs struct {
	inoc *ent.InodeClient

	ns    *ent.Namespace
	pino  *ent.Inode
	path  string
	super bool
}

func (flow *flow) createDirectory(ctx context.Context, args *createDirectoryArgs) (*ent.Inode, error) {

	inoc := args.inoc
	ns := args.ns
	pino := args.pino
	path := args.path
	dir, base := filepath.Split(args.path)

	if pino.Type != util.InodeTypeDirectory {
		return nil, status.Error(codes.AlreadyExists, "parent node is not a directory")
	}

	if pino.ReadOnly && !args.super {
		return nil, errors.New("cannot write into read-only directory")
	}

	ino, err := pino.QueryChildren().Where(entino.NameEQ(base)).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	} else if err == nil {
		if ino.Type != util.InodeTypeDirectory {
			return nil, os.ErrExist
		}
		return ino, os.ErrExist
	}

	ino, err = inoc.Create().SetName(base).SetNamespace(ns).SetParent(pino).SetReadOnly(pino.ReadOnly).SetType(util.InodeTypeDirectory).Save(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, os.ErrExist
		}
		return nil, err
	}

	pino, err = pino.Update().SetUpdatedAt(time.Now()).Save(ctx)
	if err != nil {
		return nil, err
	}

	flow.logToNamespace(ctx, time.Now(), ns, "Created directory '%s'.", path)
	flow.pubsub.NotifyInode(pino)

	// Broadcast
	err = flow.BroadcastDirectory(ctx, BroadcastEventTypeCreate,
		broadcastDirectoryInput{
			Path:   path,
			Parent: dir,
		}, ns)
	if err != nil {
		return nil, err
	}

	return ino, nil

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

	ino, err := flow.createDirectory(ctx, &createDirectoryArgs{
		inoc: tx.Inode,
		ns:   ns,
		pino: pino.ino,
		path: path,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

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

	return &resp, nil

}

type deleteNodeArgs struct {
	inoc *ent.InodeClient

	ns   *ent.Namespace
	pino *ent.Inode
	ino  *ent.Inode

	path string

	super     bool
	recursive bool
}

func (flow *flow) deleteNode(ctx context.Context, args *deleteNodeArgs) error {

	inoc := args.inoc

	ns := args.ns
	pino := args.pino
	ino := args.ino

	path := args.path
	dir, base := filepath.Split(path)

	if ino.Name == "" {
		return status.Error(codes.InvalidArgument, "cannot delete root node")
	}

	if !args.super && pino.ReadOnly {
		return status.Error(codes.InvalidArgument, "cannot delete contents of read-only directory")
	}

	if !args.recursive && ino.Type == util.InodeTypeDirectory {
		k, err := ino.QueryChildren().Count(ctx)
		if err != nil {
			return err
		}
		if k != 0 {
			return status.Error(codes.InvalidArgument, "refusing to delete non-empty directory without explicit recursive argument")
		}
	}

	err := inoc.DeleteOne(ino).Exec(ctx)
	if err != nil {
		return err
	}

	_, err = pino.Update().SetUpdatedAt(time.Now()).Save(ctx)
	if err != nil {
		return err
	}

	if ino.Type == util.InodeTypeWorkflow {
		metricsWf.WithLabelValues(ns.Name, ns.Name).Dec()
		metricsWfUpdated.WithLabelValues(ns.Name, path, ns.Name).Inc()

		// Broadcast Event
		err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeDelete,
			broadcastWorkflowInput{
				Name:   base,
				Path:   path,
				Parent: dir,
				Live:   false,
			}, ns)

		if err != nil {
			return err
		}
	} else {
		// Broadcast Event
		err = flow.BroadcastDirectory(ctx, BroadcastEventTypeDelete,
			broadcastDirectoryInput{
				Path:   path,
				Parent: dir,
			}, ns)

		if err != nil {
			return err
		}

	}

	flow.logToNamespace(ctx, time.Now(), ns, "Deleted %s '%s'.", ino.Type, path)
	flow.pubsub.NotifyInode(pino)
	flow.pubsub.CloseInode(ino)

	return nil

}

func (flow *flow) DeleteNode(ctx context.Context, req *grpc.DeleteNodeRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace

	d, err := flow.traverseToInode(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		if derrors.IsNotFound(err) && req.GetIdempotent() {
			rollback(tx)
			goto respond
		}
		return nil, err
	}

	err = flow.deleteNode(ctx, &deleteNodeArgs{
		inoc: tx.Inode,
		ns:   d.ns(),
		pino: d.ino.Edges.Parent,
		ino:  d.ino,

		path: d.path,

		recursive: req.GetRecursive(),
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

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

	if d.ino.ReadOnly && d.ino.ExtendedType != util.InodeTypeGit {
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

	if pd.ino.ReadOnly {
		return nil, errors.New("cannot write into read-only directory")
	}

	_, err = d.ino.Edges.Parent.Update().SetUpdatedAt(time.Now()).Save(ctx)
	if err != nil {
		return nil, err
	}

	ino, err = ino.Update().SetName(base).SetParent(pd.ino).Save(ctx)
	if err != nil {
		return nil, err
	}

	_, err = pd.ino.Update().SetUpdatedAt(time.Now()).Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logWithTagsToNamespace(ctx, time.Now(), d, "Renamed %s from '%s' to '%s'.", d.ino.Type, req.GetOld(), req.GetNew())
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
