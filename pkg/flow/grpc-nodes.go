package flow

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/helpers"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) Node(ctx context.Context, req *grpc.NodeRequest) (*grpc.NodeResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	var file *filestore.File
	var err error
	var ns *database.Namespace
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
		if err != nil {
			return err
		}
		file, err = tx.FileStore().ForNamespace(ns.Name).GetFile(ctx, req.GetPath())
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
	slog.Debug("Handling gRPC request", "this", this())

	var node *filestore.File
	var files []*filestore.File
	var isMirrorNamespace bool
	var err error
	var ns *database.Namespace
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
		if err != nil {
			return err
		}

		_, err = tx.DataStore().Mirror().GetConfig(ctx, ns.Name)
		if errors.Is(err, datastore.ErrNotFound) {
			isMirrorNamespace = false
		} else if err != nil {
			return err
		} else {
			isMirrorNamespace = true
		}

		node, err = tx.FileStore().ForNamespace(ns.Name).GetFile(ctx, req.GetPath())
		if err != nil {
			return err
		}
		files, err = tx.FileStore().ForNamespace(ns.Name).ReadDirectory(ctx, req.GetPath())
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
	slog.Debug("Handling gRPC request", "this", this())
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
	slog.Debug("Handling gRPC request", "this", this())

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

	file, err = tx.FileStore().ForNamespace(ns.Name).CreateFile(ctx, req.GetPath(), filestore.FileTypeDirectory, "", nil)
	if err != nil {
		return nil, err
	}

	slog.Debug("Created directory.", "path", file.Path)

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	var resp grpc.CreateDirectoryResponse

	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)

	return &resp, nil
}

func (flow *flow) DeleteNode(ctx context.Context, req *grpc.DeleteNodeRequest) (*emptypb.Empty, error) {
	slog.Debug("Handling gRPC request", "this", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := tx.FileStore().ForNamespace(ns.Name).GetFile(ctx, req.GetPath())
	if errors.Is(err, filestore.ErrNotFound) && req.GetIdempotent() {
		var resp emptypb.Empty

		return &resp, nil
	}
	if err != nil {
		return nil, err
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
	}

	slog.Debug("Deleted file", "type", file.Typ, "path", file.Path)

	var resp emptypb.Empty

	return &resp, nil
}

//nolint:goconst
func (flow *flow) RenameNode(ctx context.Context, req *grpc.RenameNodeRequest) (*grpc.RenameNodeResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, err := tx.FileStore().ForNamespace(ns.Name).GetFile(ctx, req.GetOld())
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

	if file.Typ.IsDirektivSpecFile() {
		err = helpers.PublishEventDirektivFileChange(flow.pBus, file.Typ, "rename", &pubsub.FileChangeEvent{
			Namespace:   ns.Name,
			NamespaceID: ns.ID,
			FilePath:    file.Path,
			OldPath:     req.GetOld(),
		})
		if err != nil {
			slog.Error("pubsub publish", "error", err)
		}
	}

	slog.Debug("Renamed file.", "path_old", req.GetOld(), "path", req.GetNew())

	var resp grpc.RenameNodeResponse

	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)

	return &resp, nil
}
