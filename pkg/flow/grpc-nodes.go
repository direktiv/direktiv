package flow

import (
	"context"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) Node(ctx context.Context, req *grpc.NodeRequest) (*grpc.NodeResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	var file *filestore.File
	var txErr error
	err = flow.runSqlTx(ctx, func(fStore filestore.FileStore, store datastore.Store) error {
		file, txErr = fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
		return txErr
	})
	if err != nil {
		return nil, err
	}
	resp := &grpc.NodeResponse{}
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.Namespace = ns.Name

	return resp, nil
}

func (flow *flow) Directory(ctx context.Context, req *grpc.DirectoryRequest) (*grpc.DirectoryResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	var files []*filestore.File
	var txErr error
	err = flow.runSqlTx(ctx, func(fStore filestore.FileStore, store datastore.Store) error {
		files, txErr = fStore.ForRootID(ns.ID).ReadDirectory(ctx, req.GetPath())
		return txErr
	})
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

func (flow *flow) DirectoryStream(req *grpc.DirectoryRequest, srv grpc.Flow_DirectoryStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	ctx := srv.Context()

	resp, err := flow.Directory(ctx, req)
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

func (flow *flow) CreateDirectory(ctx context.Context, req *grpc.CreateDirectoryRequest) (*grpc.CreateDirectoryResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	var file *filestore.File

	fStore, _, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(ctx)

	file, _, err = fStore.ForRootID(ns.ID).CreateFile(ctx, req.GetPath(), filestore.FileTypeDirectory, nil)
	if err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, ns.ID, database.GetAttributes(recipient.Namespace, ns), "Created directory '%s'.", file.Path)

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

	if err := commit(ctx); err != nil {
		return nil, err
	}

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

	fStore, _, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(ctx)

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

	if err := commit(ctx); err != nil {
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

	flow.logger.Infof(ctx, ns.ID, database.GetAttributes(recipient.Namespace, ns), "Deleted %s '%s'.", file.Typ, file.Path)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) RenameNode(ctx context.Context, req *grpc.RenameNodeRequest) (*grpc.RenameNodeResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	fStore, _, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(ctx)

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

	if err := commit(ctx); err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, ns.ID, database.GetAttributes(recipient.Namespace, ns), "Renamed %s from '%s' to '%s'.", file.Typ, req.GetOld(), req.GetNew())

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

	fStore, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(ctx)

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	annotations, err := store.FileAnnotations().Get(ctx, file.ID)

	if err == core.ErrFileAnnotationsNotSet {
		annotations = &core.FileAnnotations{
			FileID: file.ID,
			Data:   nil,
		}
	} else if err != nil {
		return nil, err
	}

	annotations.Data = annotations.Data.AppendFileUserAttributes(req.GetAttributes())

	err = store.FileAnnotations().Set(ctx, annotations)
	if err != nil {
		return nil, err
	}

	if err := commit(ctx); err != nil {
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
	fStore, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(ctx)

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	annotations, err := store.FileAnnotations().Get(ctx, file.ID)

	if err == core.ErrFileAnnotationsNotSet {
		return nil, status.Error(codes.InvalidArgument, "file annotations are not set")
	} else if err != nil {
		return nil, err
	}

	annotations.Data = annotations.Data.ReduceFileUserAttributes(req.GetAttributes())

	err = store.FileAnnotations().Set(ctx, annotations)
	if err != nil {
		return nil, err
	}

	if err := commit(ctx); err != nil {
		return nil, err
	}
	var resp emptypb.Empty

	return &resp, nil
}
