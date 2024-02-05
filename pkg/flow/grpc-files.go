package flow

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/gabriel-vasile/mimetype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (flow *flow) CreateFile(ctx context.Context, req *grpc.CreateFileRequest) (*grpc.CreateFileResponse, error) {
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

	data := req.GetSource()

	mimeType := req.GetMimeType()
	if mimeType == "" {
		mt := mimetype.Detect(data)
		mimeType = mt.String()
	}

	file, err := tx.FileStore().ForNamespace(ns.Name).CreateFile(ctx, req.GetPath(), filestore.FileTypeFile, mimeType, data)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, ns.ID, database.GetAttributes(recipient.Namespace, ns), "Created file '%s'.", file.Path)

	resp := &grpc.CreateFileResponse{}
	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.File = bytedata.ConvertFileToGrpcFile(file)
	resp.File.Source = data

	return resp, nil
}

func (flow *flow) UpdateFile(ctx context.Context, req *grpc.UpdateFileRequest) (*grpc.UpdateFileResponse, error) {
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

	data := req.GetSource()

	mimeType := req.GetMimeType()
	if mimeType == "" {
		mt := mimetype.Detect(data)
		mimeType = mt.String()
	}

	file, err := tx.FileStore().ForNamespace(ns.Name).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	if file.Typ != filestore.FileTypeWorkflow {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("file type '%s'", file.Typ))
	}

	err = tx.FileStore().ForFile(file).Delete(ctx, false)
	if err != nil {
		return nil, err
	}

	file, err = tx.FileStore().ForNamespace(ns.Name).CreateFile(ctx, req.GetPath(), filestore.FileTypeFile, mimeType, data)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, ns.ID, database.GetAttributes(recipient.Namespace, ns), "Updated file '%s'.", file.Path)

	var resp grpc.UpdateFileResponse

	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.File = bytedata.ConvertFileToGrpcFile(file)
	resp.File.Source = data

	return &resp, nil
}

func (flow *flow) File(ctx context.Context, req *grpc.FileRequest) (*grpc.FileResponse, error) {
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

	file, err := tx.FileStore().ForNamespace(ns.Name).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	data, err := tx.FileStore().ForFile(file).GetData(ctx)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	resp := new(grpc.FileResponse)
	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.File = bytedata.ConvertFileToGrpcFile(file)
	resp.File.Source = data

	return resp, nil
}
