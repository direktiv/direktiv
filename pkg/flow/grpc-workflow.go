package flow

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/helpers"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

func (flow *flow) Workflow(ctx context.Context, req *grpc.WorkflowRequest) (*grpc.WorkflowResponse, error) {
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

	resp := new(grpc.WorkflowResponse)
	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.Source = data

	return resp, nil
}

func (flow *flow) createFileSystemObject(ctx context.Context, fileType filestore.FileType, req *grpc.CreateWorkflowRequest,
) (*grpc.CreateWorkflowResponse, error) {
	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	file, err := tx.FileStore().ForNamespace(ns.Name).CreateFile(ctx, req.GetPath(),
		fileType, "application/yaml", req.GetSource())
	if err != nil {
		return nil, err
	}

	data, err := tx.FileStore().ForFile(file).GetData(ctx)
	if err != nil {
		return nil, err
	}

	resp := &grpc.CreateWorkflowResponse{}
	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.Source = data

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	slog.Debug("Created file.", "type", fileType, "path", file.Path)

	err = helpers.PublishEventDirektivFileChange(flow.pBus, file.Typ, "create", &pubsub.FileChangeEvent{
		Namespace:   ns.Name,
		NamespaceID: ns.ID,
		FilePath:    file.Path,
	})
	if err != nil {
		slog.Error("pubsub publish", "error", err)
	}

	return resp, nil
}

func (flow *flow) CreateWorkflow(ctx context.Context, req *grpc.CreateWorkflowRequest) (*grpc.CreateWorkflowResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	if filepath.Ext(req.GetPath()) != ".yaml" && filepath.Ext(req.GetPath()) != ".yml" {
		return nil, status.Error(codes.InvalidArgument, "direktiv spec file name should have either .yaml or .yaml extension")
	}

	type APIFile struct {
		DirektivAPI string `yaml:"direktiv_api"`
	}

	apiFile := &APIFile{}
	err := yaml.Unmarshal(req.GetSource(), apiFile)
	if err != nil {
		return nil, err
	}

	// check for other file types first
	switch apiFile.DirektivAPI {
	case model.ServiceAPIV1:
		return flow.createFileSystemObject(ctx, filestore.FileTypeService, req)
	case model.EndpointAPIV1:
		return flow.createFileSystemObject(ctx, filestore.FileTypeEndpoint, req)
	case model.ConsumerAPIV1:
		return flow.createFileSystemObject(ctx, filestore.FileTypeConsumer, req)
	}

	// do workflow if no other type detected
	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	if len(req.GetSource()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty workflow is not allowed")
	}
	file, err := tx.FileStore().ForNamespace(ns.Name).CreateFile(ctx, req.GetPath(), filestore.FileTypeWorkflow, "application/yaml", req.GetSource())
	if err != nil {
		return nil, err
	}

	data, err := tx.FileStore().ForFile(file).GetData(ctx)
	if err != nil {
		return nil, err
	}

	workflow := new(model.Workflow)
	err = workflow.Load(data)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	metricsWf.WithLabelValues(ns.Name, ns.Name).Inc()
	metricsWfUpdated.WithLabelValues(ns.Name, file.Path, ns.Name).Inc()

	slog.Debug("Created workflow.", "path", file.Path, "namespace", req.Namespace)

	if file.Typ.IsDirektivSpecFile() {
		err = helpers.PublishEventDirektivFileChange(flow.pBus, file.Typ, "create", &pubsub.FileChangeEvent{
			Namespace:   ns.Name,
			NamespaceID: ns.ID,
			FilePath:    file.Path,
		})
		if err != nil {
			slog.Error("pubsub publish", "error", err)
		}
	}

	resp := &grpc.CreateWorkflowResponse{}
	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.Source = data

	return resp, nil
}

func (flow *flow) UpdateWorkflow(ctx context.Context, req *grpc.UpdateWorkflowRequest) (*grpc.UpdateWorkflowResponse, error) {
	// This is being called by the frontend when a user changes a workflow via a UI and press save button.

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
	if err != nil {
		return nil, err
	}
	switch file.Typ {
	case filestore.FileTypeWorkflow, filestore.FileTypeService, filestore.FileTypeEndpoint, filestore.FileTypeConsumer:
		// Valid file type, continue processing
	default:
		return nil, status.Error(codes.InvalidArgument, "file type is not workflow or service or endpoint or consumer")
	}
	_, err = tx.FileStore().ForFile(file).SetData(ctx, req.GetSource())
	if err != nil {
		return nil, err
	}
	file, err = tx.FileStore().ForNamespace(ns.Name).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	if file.Typ.IsDirektivSpecFile() {
		err = helpers.PublishEventDirektivFileChange(flow.pBus, file.Typ, "update", &pubsub.FileChangeEvent{
			Namespace:   ns.Name,
			NamespaceID: ns.ID,
			FilePath:    file.Path,
		})
		if err != nil {
			slog.Error("pubsub publish", "error", err)
		}
	}

	var resp grpc.UpdateWorkflowResponse

	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.Source = req.GetSource()

	return &resp, nil
}
