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

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
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

func (srv *server) traverseToInode(ctx context.Context, namespace, path string) (*database.CacheData, error) {
	cached := new(database.CacheData)

	err := srv.database.NamespaceByName(ctx, cached, namespace)
	if err != nil {
		return nil, err
	}

	err = srv.database.InodeByPath(ctx, cached, path)
	if err != nil {
		return nil, err
	}

	return cached, nil
}

func (flow *flow) Node(ctx context.Context, req *grpc.NodeRequest) (*grpc.NodeResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	var err error
	var resp grpc.NodeResponse

	cached, err := flow.traverseToInode(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		flow.sugar.Debugf("gRPC %s handler failed to traverse to (namespace/inode) %s/%s : %v", this(), req.GetNamespace(), req.GetPath(), err)
		return nil, err
	}

	err = bytedata.ConvertDataForOutput(cached.Inode(), &resp.Node)
	if err != nil {
		flow.sugar.Debugf("gRPC %s handler failed to traverse to marshal response: %v", this(), err)
		return nil, err
	}

	resp.Namespace = cached.Namespace.Name
	resp.Node.Path = cached.Path()
	resp.Node.Parent = cached.Dir()

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	return &resp, nil
}

func (flow *flow) Directory(ctx context.Context, req *grpc.DirectoryRequest) (*grpc.DirectoryResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, err := flow.traverseToInode(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	if cached.Inode().Type != util.InodeTypeDirectory {
		return nil, ErrNotDir
	}

	clients := flow.edb.Clients(ctx)

	query := clients.Inode.Query().Where(entino.HasParentWith(entino.ID(cached.Inode().ID)))

	results, pi, err := paginate[*ent.InodeQuery, *ent.Inode](ctx, req.Pagination, query, inodesOrderings, inodesFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.DirectoryResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Children = new(grpc.DirectoryChildren)
	resp.Children.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Children.Results)
	if err != nil {
		return nil, err
	}

	err = bytedata.ConvertDataForOutput(cached.Inode(), &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Node.Path = cached.Path()
	resp.Node.Parent = cached.Dir()

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

	cached, err := flow.traverseToInode(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	if cached.Inode().Type != util.InodeTypeDirectory {
		return ErrNotDir
	}

	sub := flow.pubsub.SubscribeInode(cached)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(ctx)

	query := clients.Inode.Query().Where(entino.HasParentWith(entino.ID(cached.Inode().ID)))

	results, pi, err := paginate[*ent.InodeQuery, *ent.Inode](ctx, req.Pagination, query, inodesOrderings, inodesFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.DirectoryResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Children = new(grpc.DirectoryChildren)
	resp.Children.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Children.Results)
	if err != nil {
		return err
	}

	err = bytedata.ConvertDataForOutput(cached.Inode(), &resp.Node)
	if err != nil {
		return err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Node.Path = cached.Path()
	resp.Node.Parent = cached.Dir()

	for idx := range resp.Children.Results {
		child := resp.Children.Results[idx]
		child.Parent = resp.Node.Path
		child.Path = filepath.Join(resp.Node.Path, child.Name)

		if child.ExpandedType == "" {
			child.ExpandedType = child.Type
		}

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

type createDirectoryArgs struct {
	pcached *database.CacheData
	path    string
	super   bool
}

func (flow *flow) createDirectory(ctx context.Context, args *createDirectoryArgs) (*database.Inode, error) {
	dir, base := filepath.Split(args.path)

	if args.pcached.Inode().ReadOnly && !args.super {
		return nil, errors.New("cannot write into read-only directory")
	}

	ino, err := flow.database.CreateDirectoryInode(ctx, &database.CreateDirectoryInodeArgs{
		Name:     base,
		ReadOnly: args.pcached.Inode().ReadOnly,
		Parent:   args.pcached.Inode(),
	})
	if err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, args.pcached.Namespace.ID, args.pcached.GetAttributes("namespace"), "Created directory '%s'.", args.path)
	flow.pubsub.NotifyInode(args.pcached.Inode())

	// Broadcast
	err = flow.BroadcastDirectory(ctx, BroadcastEventTypeCreate,
		broadcastDirectoryInput{
			Path:   args.path,
			Parent: dir,
		}, args.pcached)
	if err != nil {
		return nil, err
	}

	return ino, nil
}

func (flow *flow) CreateDirectory(ctx context.Context, req *grpc.CreateDirectoryRequest) (*grpc.CreateDirectoryResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	path := GetInodePath(req.GetPath())
	dir, base := filepath.Split(path)

	if base == "" || base == "/" {
		return nil, status.Error(codes.AlreadyExists, "root directory already exists")
	}

	cached := new(database.CacheData)

	err = flow.database.NamespaceByName(tctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	err = flow.database.InodeByPath(tctx, cached, dir)
	if err != nil {
		return nil, err
	}

	ino, err := flow.createDirectory(tctx, &createDirectoryArgs{
		pcached: cached,
		path:    path,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.database.InvalidateInode(ctx, cached, false)

	var resp grpc.CreateDirectoryResponse

	err = bytedata.ConvertDataForOutput(ino, &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Node.ReadOnly = false

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = dir
	resp.Node.Path = path

	return &resp, nil
}

type deleteNodeArgs struct {
	cached    *database.CacheData
	super     bool
	recursive bool
}

func (flow *flow) deleteNode(ctx context.Context, args *deleteNodeArgs) error {
	if args.cached.Inode().Name == "" {
		return status.Error(codes.InvalidArgument, "cannot delete root node")
	}

	if !args.super && args.cached.ParentInode().ReadOnly {
		return status.Error(codes.InvalidArgument, "cannot delete contents of read-only directory")
	}

	if !args.recursive && args.cached.Inode().Type == util.InodeTypeDirectory {
		if len(args.cached.Inode().Children) != 0 {
			return status.Error(codes.InvalidArgument, "refusing to delete non-empty directory without explicit recursive argument")
		}
	}

	clients := flow.edb.Clients(ctx)

	err := clients.Inode.DeleteOneID(args.cached.Inode().ID).Exec(ctx)
	if err != nil {
		return err
	}

	x, err := clients.Inode.UpdateOneID(args.cached.ParentInode().ID).SetUpdatedAt(time.Now()).Save(ctx)
	if err != nil {
		return err
	}

	args.cached.ParentInode().UpdatedAt = x.UpdatedAt

	if args.cached.Inode().Type == util.InodeTypeWorkflow {
		metricsWf.WithLabelValues(args.cached.Namespace.Name, args.cached.Namespace.Name).Dec()
		metricsWfUpdated.WithLabelValues(args.cached.Namespace.Name, args.cached.Path(), args.cached.Namespace.Name).Inc()

		// Broadcast Event
		err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeDelete,
			broadcastWorkflowInput{
				Name:   args.cached.Inode().Name,
				Path:   args.cached.Path(),
				Parent: args.cached.Dir(),
				Live:   false,
			}, args.cached)

		if err != nil {
			return err
		}
	} else {
		// Broadcast Event
		err = flow.BroadcastDirectory(ctx, BroadcastEventTypeDelete,
			broadcastDirectoryInput{
				Path:   args.cached.Path(),
				Parent: args.cached.Dir(),
			}, args.cached)

		if err != nil {
			return err
		}

	}

	flow.logger.Infof(ctx, args.cached.Namespace.ID, args.cached.GetAttributes(recipient.Namespace), "Deleted %s '%s'.", args.cached.Inode().Type, args.cached.Path())
	flow.pubsub.NotifyInode(args.cached.ParentInode())
	flow.pubsub.CloseInode(args.cached.Inode())

	args.cached.Inodes = args.cached.Inodes[:len(args.cached.Inodes)-1]

	return nil
}

func (flow *flow) DeleteNode(ctx context.Context, req *grpc.DeleteNodeRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToInode(tctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		if derrors.IsNotFound(err) && req.GetIdempotent() {
			rollback(tx)
			goto respond
		}
		return nil, err
	}

	err = flow.deleteNode(tctx, &deleteNodeArgs{
		cached:    cached,
		recursive: req.GetRecursive(),
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.database.InvalidateInode(ctx, cached, false)

respond:

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) RenameNode(ctx context.Context, req *grpc.RenameNodeRequest) (*grpc.RenameNodeResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToInode(tctx, req.GetNamespace(), req.GetOld())
	if err != nil {
		return nil, err
	}

	if cached.Path() == "/" {
		return nil, errors.New("cannot rename root node")
	}

	path := GetInodePath(req.GetNew())
	if path == "/" {
		return nil, errors.New("cannot overwrite root node")
	}

	if strings.Contains(path, cached.Path()+"/") {
		return nil, errors.New("cannot move node into itself")
	}

	if cached.Inode().ReadOnly && cached.Inode().ExtendedType != util.InodeTypeGit {
		return nil, errors.New("cannot move contents of read-only directory")
	}

	dir, base := filepath.Split(path)

	pcached, err := flow.traverseToInode(tctx, req.GetNamespace(), dir)
	if err != nil {
		return nil, err
	}

	if pcached.Inode().ReadOnly {
		return nil, errors.New("cannot write into read-only directory")
	}

	clients := flow.edb.Clients(tctx)

	x, err := clients.Inode.UpdateOneID(cached.ParentInode().ID).SetUpdatedAt(time.Now()).Save(tctx)
	if err != nil {
		return nil, err
	}
	cached.ParentInode().UpdatedAt = x.UpdatedAt

	_, err = clients.Inode.UpdateOneID(cached.Inode().ID).SetName(base).SetParentID(pcached.Inode().ID).Save(tctx)
	if err != nil {
		return nil, err
	}

	_, err = clients.Inode.UpdateOneID(pcached.Inode().ID).SetUpdatedAt(time.Now()).Save(tctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.database.InvalidateInode(ctx, cached.Parent(), false)
	flow.database.InvalidateInode(ctx, pcached, false)

	flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Renamed %s from '%s' to '%s'.", cached.Inode().Type, req.GetOld(), req.GetNew())
	flow.pubsub.NotifyInode(cached.ParentInode())
	flow.pubsub.NotifyInode(pcached.Inode())
	flow.pubsub.CloseInode(cached.Inode())

	var resp grpc.RenameNodeResponse

	err = bytedata.ConvertDataForOutput(cached.Inode(), &resp.Node)
	if err != nil {
		return nil, err
	}

	if resp.Node.ExpandedType == "" {
		resp.Node.ExpandedType = resp.Node.Type
	}

	resp.Node.ReadOnly = false

	resp.Namespace = cached.Namespace.Name
	resp.Node.Parent = dir
	resp.Node.Path = path

	return &resp, nil
}

func (flow *flow) CreateNodeAttributes(ctx context.Context, req *grpc.CreateNodeAttributesRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToInode(tctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	m := make(map[string]bool)

	for _, attr := range cached.Inode().Attributes {
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

	clients := flow.edb.Clients(tctx)

	_, err = clients.Inode.UpdateOneID(cached.Inode().ID).SetAttributes(attrs).Save(tctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) DeleteNodeAttributes(ctx context.Context, req *grpc.DeleteNodeAttributesRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToInode(tctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	m := make(map[string]bool)

	for _, attr := range cached.Inode().Attributes {
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

	clients := flow.edb.Clients(tctx)

	_, err = clients.Inode.UpdateOneID(cached.Inode().ID).SetAttributes(attrs).Save(tctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil
}
