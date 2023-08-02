package flow

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) Node(ctx context.Context, req *grpc.NodeRequest) (*grpc.NodeResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	var file *filestore.File
	var err error
	var ns *core.Namespace
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
		if err != nil {
			return err
		}
		file, err = tx.FileStore().ForRootNamespaceAndName(ns.ID, defaultRootName).GetFile(ctx, req.GetPath())
		return err
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

	var node *filestore.File
	var files []*filestore.File
	var isMirrorNamespace bool
	var err error
	var ns *core.Namespace
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
		if err != nil {
			return err
		}

		_, err = tx.DataStore().Mirror().GetConfig(ctx, ns.ID)
		if errors.Is(err, mirror.ErrNotFound) {
			isMirrorNamespace = false
		} else if err != nil {
			return err
		} else {
			isMirrorNamespace = true
		}

		node, err = tx.FileStore().ForRootNamespaceAndName(ns.ID, defaultRootName).GetFile(ctx, req.GetPath())
		if err != nil {
			return err
		}
		files, err = tx.FileStore().ForRootNamespaceAndName(ns.ID, defaultRootName).ReadDirectory(ctx, req.GetPath())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	resp := new(grpc.DirectoryResponse)
	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(node)

	if isMirrorNamespace && node.Path == "/" {
		resp.Node.ExpandedType = "git"
	}

	resp.Children = new(grpc.DirectoryChildren)
	resp.Children.PageInfo = new(grpc.PageInfo)
	resp.Children.PageInfo.Total = int32(len(files))
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

	var file *filestore.File

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, _, err = tx.FileStore().ForRootNamespaceAndName(ns.ID, defaultRootName).CreateFile(ctx, req.GetPath(), filestore.FileTypeDirectory, "", nil)
	if err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, ns.ID, database.GetAttributes(recipient.Namespace, ns), "Created directory '%s'.", file.Path)

	// Broadcast
	err = flow.BroadcastDirectory(ctx, BroadcastEventTypeCreate,
		broadcastDirectoryInput{
			Path:   req.GetPath(),
			Parent: file.Dir(),
		}, ns)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	var resp grpc.CreateDirectoryResponse

	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)

	return &resp, nil
}

func (flow *flow) DeleteNode(ctx context.Context, req *grpc.DeleteNodeRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := tx.FileStore().ForRootNamespaceAndName(ns.ID, defaultRootName).GetFile(ctx, req.GetPath())
	if errors.Is(err, filestore.ErrNotFound) && req.GetIdempotent() {
		var resp emptypb.Empty

		return &resp, nil
	}
	if err != nil {
		return nil, err
	}

	if file.Typ == filestore.FileTypeDirectory {
		isEmptyDir, err := tx.FileStore().ForRootNamespaceAndName(ns.ID, defaultRootName).IsEmptyDirectory(ctx, req.GetPath())
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

	err = tx.FileStore().ForFile(file).Delete(ctx, req.GetRecursive())
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	if file.Typ == filestore.FileTypeWorkflow {
		metricsWf.WithLabelValues(ns.Name, ns.Name).Dec()
		metricsWfUpdated.WithLabelValues(ns.Name, file.Path, ns.Name).Inc()

		// Broadcast Event
		err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeDelete,
			broadcastWorkflowInput{
				Name:   file.Name(),
				Path:   file.Path,
				Parent: file.Dir(),
				Live:   false,
			}, ns)
		if err != nil {
			return nil, err
		}
	} else {
		// Broadcast Event
		err = flow.BroadcastDirectory(ctx, BroadcastEventTypeDelete,
			broadcastDirectoryInput{
				Path:   file.Path,
				Parent: file.Dir(),
			}, ns)

		if err != nil {
			return nil, err
		}
	}

	flow.logger.Infof(ctx, ns.ID, database.GetAttributes(recipient.Namespace, ns), "Deleted %s '%s'.", file.Typ, file.Path)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) RenameNode(ctx context.Context, req *grpc.RenameNodeRequest) (*grpc.RenameNodeResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := tx.FileStore().ForRootNamespaceAndName(ns.ID, defaultRootName).GetFile(ctx, req.GetOld())
	if err != nil {
		return nil, err
	}

	if file.Path == "/" {
		return nil, status.Error(codes.InvalidArgument, "cannot rename root node")
	}
	if file.Typ == filestore.FileTypeWorkflow {
		if filepath.Ext(req.GetNew()) != ".yaml" && filepath.Ext(req.GetNew()) != ".yml" {
			return nil, status.Error(codes.InvalidArgument, "workflow name should have either .yaml or .yaml extension")
		}
	}

	err = tx.FileStore().ForFile(file).SetPath(ctx, req.GetNew())
	if err != nil {
		return nil, err
	}
	// TODO: question if parent dir need to get updated_at change.

	if err := tx.Commit(ctx); err != nil {
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

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := tx.FileStore().ForRootNamespaceAndName(ns.ID, defaultRootName).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	annotations, err := tx.DataStore().FileAnnotations().Get(ctx, file.ID)

	if errors.Is(err, core.ErrFileAnnotationsNotSet) {
		annotations = &core.FileAnnotations{
			FileID: file.ID,
			Data:   map[string]string{},
		}
	} else if err != nil {
		return nil, err
	}

	annotations.Data = annotations.Data.AppendFileUserAttributes(req.GetAttributes())

	err = tx.DataStore().FileAnnotations().Set(ctx, annotations)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) DeleteNodeAttributes(ctx context.Context, req *grpc.DeleteNodeAttributesRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := tx.FileStore().ForRootNamespaceAndName(ns.ID, defaultRootName).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	annotations, err := tx.DataStore().FileAnnotations().Get(ctx, file.ID)

	if errors.Is(err, core.ErrFileAnnotationsNotSet) {
		return nil, status.Error(codes.InvalidArgument, "file annotations are not set")
	} else if err != nil {
		return nil, err
	}

	annotations.Data = annotations.Data.ReduceFileUserAttributes(req.GetAttributes())

	err = tx.DataStore().FileAnnotations().Set(ctx, annotations)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	var resp emptypb.Empty

	return &resp, nil
}
