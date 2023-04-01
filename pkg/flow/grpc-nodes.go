package flow

import (
	"context"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) Node(ctx context.Context, req *grpc.NodeRequest) (*grpc.NodeResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	file, err := flow.fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	resp := &grpc.NodeResponse{}
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.Namespace = ns.Name

	return resp, nil
}

func (flow *flow) directory(ctx context.Context, req *grpc.DirectoryRequest) (*grpc.DirectoryResponse, error) {
	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	files, err := flow.fStore.ForRootID(ns.ID).ReadDirectory(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	resp := new(grpc.DirectoryResponse)
	resp.Namespace = ns.Name
	resp.Children = new(grpc.DirectoryChildren)
	resp.Children.PageInfo = new(grpc.PageInfo)

	resp.Children.Results = bytedata.ConvertFilesToGrpcNodeList(files)

	return resp, nil
}

func (flow *flow) Directory(ctx context.Context, req *grpc.DirectoryRequest) (*grpc.DirectoryResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	return flow.directory(ctx, req)
}

func (flow *flow) DirectoryStream(req *grpc.DirectoryRequest, srv grpc.Flow_DirectoryStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	ctx := srv.Context()

	res, err := flow.directory(ctx, req)
	if err != nil {
		return err
	}
	err = srv.Send(res)
	if err != nil {
		return err
	}

	// fake stream.
	time.Sleep(time.Second * 10)

	return nil
}

func (flow *flow) CreateDirectory(ctx context.Context, req *grpc.CreateDirectoryRequest) (*grpc.CreateDirectoryResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	var file *filestore.File
	err = flow.fStore.Tx(ctx, func(ctx context.Context, fStore filestore.FileStore) error {
		file, _, err = flow.fStore.ForRootID(ns.ID).CreateFile(ctx, req.GetPath(), filestore.FileTypeDirectory, nil)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// TODO: alex, need fix here.
	// flow.logger.Infof(ctx, ns.ID, args.pcached.GetAttributes("namespace"), "Created directory '%s'.", args.path)

	// Broadcast
	// TODO: yassir, need question here.
	//err = flow.BroadcastDirectory(ctx, BroadcastEventTypeCreate,
	//	broadcastDirectoryInput{
	//		Path:   req.GetPath(),
	//		Parent: filepath.Dir(req.GetPath()),
	//	}, args.pcached)
	//if err != nil {
	//	return nil, err
	//}

	var resp grpc.CreateDirectoryResponse

	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)

	return &resp, nil
}

func (flow *flow) DeleteNode(ctx context.Context, req *grpc.DeleteNodeRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	fStore, err := flow.fStore.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer fStore.Rollback(ctx)

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err == filestore.ErrNotFound && req.GetIdempotent() {
		var resp emptypb.Empty

		return &resp, nil
	}
	if err != nil {
		return nil, err
	}

	if file.Typ == filestore.FileTypeDirectory {
		isEmptyDir, err := fStore.ForRootID(ns.ID).IsEmptyDirectory(ctx, req.GetPath())
		if err != nil {
			return nil, err
		}
		if !isEmptyDir && !req.GetRecursive() {
			return nil, status.Error(codes.InvalidArgument, "refusing to delete non-empty directory without explicit recursive argument")
		}
	}
	if file.Path == "/" {
		return nil, status.Error(codes.InvalidArgument, "cannot delete root node")
	}

	err = fStore.ForFile(file).Delete(ctx, req.GetRecursive())
	if err != nil {
		return nil, err
	}

	err = fStore.Commit(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: yassir, need fix here.
	//if file.Typ == filestore.FileTypeWorkflow {
	//	metricsWf.WithLabelValues(args.cached.Namespace.Name, args.cached.Namespace.Name).Dec()
	//	metricsWfUpdated.WithLabelValues(args.cached.Namespace.Name, args.cached.Path(), args.cached.Namespace.Name).Inc()
	//
	//	// Broadcast Event
	//	err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeDelete,
	//		broadcastWorkflowInput{
	//			Name:   args.cached.Inode().Name,
	//			Path:   args.cached.Path(),
	//			Parent: args.cached.Dir(),
	//			Live:   false,
	//		}, args.cached)
	//
	//	if err != nil {
	//		return err
	//	}
	//} else {
	//	// Broadcast Event
	//	err = flow.BroadcastDirectory(ctx, BroadcastEventTypeDelete,
	//		broadcastDirectoryInput{
	//			Path:   args.cached.Path(),
	//			Parent: args.cached.Dir(),
	//		}, args.cached)
	//
	//	if err != nil {
	//		return err
	//	}
	//
	//}

	// TODO: alex, need fix here.
	// flow.logger.Infof(ctx, ns.ID, args.cached.GetAttributes(recipient.Namespace), "Deleted %s '%s'.", args.cached.Inode().Type, args.cached.Path())

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) RenameNode(ctx context.Context, req *grpc.RenameNodeRequest) (*grpc.RenameNodeResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	fStore, err := flow.fStore.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer fStore.Rollback(ctx)

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetOld())
	if err != nil {
		return nil, err
	}

	if file.Path == "/" {
		return nil, status.Error(codes.InvalidArgument, "cannot rename root node")
	}

	err = fStore.ForFile(file).SetPath(ctx, req.GetNew())
	if err != nil {
		return nil, err
	}
	// TODO: question if parent dir need to get updated_at change.

	err = fStore.Commit(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: alex, need fix here.
	// flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Renamed %s from '%s' to '%s'.", cached.Inode().Type, req.GetOld(), req.GetNew())

	var resp grpc.RenameNodeResponse

	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)

	return &resp, nil
}

func (flow *flow) CreateNodeAttributes(ctx context.Context, req *grpc.CreateNodeAttributesRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := flow.fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	store, err := flow.dataStore.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer store.Rollback(ctx)

	attributes, err := store.FileAttributes().Get(ctx, file.ID)

	if err == core.ErrFileAttributesNotSet {
		attributes = &core.FileAttributes{
			FileID: file.ID,
			Value:  core.NewFileAttributesValue(req.Attributes),
		}
		err := store.FileAttributes().Set(ctx, attributes)
		if err != nil {
			return nil, err
		}
		var resp emptypb.Empty

		return &resp, nil
	}

	if err != nil {
		return nil, err
	}

	err = store.FileAttributes().Set(ctx, attributes.Add(req.Attributes))
	if err != nil {
		return nil, err
	}

	err = store.Commit(ctx)
	if err != nil {
		return nil, err
	}
	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) DeleteNodeAttributes(ctx context.Context, req *grpc.DeleteNodeAttributesRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := flow.fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	store, err := flow.dataStore.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer store.Rollback(ctx)

	attributes, err := store.FileAttributes().Get(ctx, file.ID)

	if err == core.ErrFileAttributesNotSet {
		status.Error(codes.InvalidArgument, "file attributes are not set")
	}

	if err != nil {
		return nil, err
	}

	err = store.FileAttributes().Set(ctx, attributes.Remove(req.Attributes))
	if err != nil {
		return nil, err
	}

	err = store.Commit(ctx)
	if err != nil {
		return nil, err
	}
	var resp emptypb.Empty

	return &resp, nil
}
