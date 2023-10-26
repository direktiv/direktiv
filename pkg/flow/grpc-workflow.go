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
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) ResolveWorkflowUID(ctx context.Context, req *grpc.ResolveWorkflowUIDRequest) (*grpc.WorkflowResponse, error) {
	// TODO: yassir, low priority. probably un used.
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	var resp *grpc.WorkflowResponse

	return resp, nil
}

func (flow *flow) Workflow(ctx context.Context, req *grpc.WorkflowRequest) (*grpc.WorkflowResponse, error) {
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

	ref := req.GetRef()
	if ref == "" {
		ref = filestore.Latest
	}
	revision, err := tx.FileStore().ForFile(file).GetRevision(ctx, ref)
	if err != nil {
		return nil, err
	}

	data, err := tx.FileStore().ForRevision(revision).GetData(ctx)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	resp := new(grpc.WorkflowResponse)
	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.Revision = bytedata.ConvertRevisionToGrpcRevision(revision)
	resp.Revision.Source = data
	resp.EventLogging = ""
	resp.Oid = file.ID.String()

	return resp, nil
}

func (flow *flow) WorkflowStream(req *grpc.WorkflowRequest, srv grpc.Flow_WorkflowStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	ctx := srv.Context()

	resp, err := flow.Workflow(ctx, req)
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

func (flow *flow) createService(ctx context.Context, req *grpc.CreateWorkflowRequest) (*grpc.CreateWorkflowResponse, error) {
	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	file, revision, err := tx.FileStore().ForNamespace(ns.Name).CreateFile(ctx, req.GetPath(), filestore.FileTypeService, "application/direktiv", req.GetSource())
	if err != nil {
		return nil, err
	}
	data, err := tx.FileStore().ForRevision(revision).GetData(ctx)
	if err != nil {
		return nil, err
	}

	resp := &grpc.CreateWorkflowResponse{}
	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.Revision = bytedata.ConvertRevisionToGrpcRevision(revision)
	resp.Revision.Source = data

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	flow.logger.Infof(ctx, ns.ID, database.GetAttributes(recipient.Namespace, ns), "Created service '%s'.", file.Path)

	err = flow.pBus.Publish(pubsub.ServiceCreate, file.Path)
	if err != nil {
		flow.sugar.Error("pubsub publish", "error", err)
	}

	return resp, nil
}

func (flow *flow) CreateWorkflow(ctx context.Context, req *grpc.CreateWorkflowRequest) (*grpc.CreateWorkflowResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	if filepath.Ext(req.GetPath()) != ".yaml" && filepath.Ext(req.GetPath()) != ".yml" {
		return nil, status.Error(codes.InvalidArgument, "direktiv spec file name should have either .yaml or .yaml extension")
	}

	if _, err := spec.ParseServicesFile(req.GetSource()); err == nil {
		return flow.createService(ctx, req)
	}

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
	file, revision, err := tx.FileStore().ForNamespace(ns.Name).CreateFile(ctx, req.GetPath(), filestore.FileTypeWorkflow, "application/direktiv", req.GetSource())
	if err != nil {
		return nil, err
	}

	data, err := tx.FileStore().ForRevision(revision).GetData(ctx)
	if err != nil {
		return nil, err
	}

	workflow := new(model.Workflow)
	err = workflow.Load(data)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, router, err := getRouter(ctx, tx, file)
	if err != nil {
		return nil, err
	}

	err = flow.configureWorkflowStarts(ctx, tx, ns.ID, file, router, true)
	if err != nil {
		return nil, err
	}

	err = flow.placeholdSecrets(ctx, tx, ns.Name, file)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	metricsWf.WithLabelValues(ns.Name, ns.Name).Inc()
	metricsWfUpdated.WithLabelValues(ns.Name, file.Path, ns.Name).Inc()

	flow.logger.Infof(ctx, ns.ID, database.GetAttributes(recipient.Namespace, ns), "Created workflow '%s'.", file.Path)

	err = flow.BroadcastWorkflow(ctx, BroadcastEventTypeCreate,
		broadcastWorkflowInput{
			Name:   file.Name(),
			Path:   file.Path,
			Parent: file.Dir(),
			Live:   true,
		}, ns)

	if err != nil {
		return nil, err
	}

	err = flow.pBus.Publish(pubsub.WorkflowCreate, file.Path)
	if err != nil {
		flow.sugar.Error("pubsub publish", "error", err)
	}

	resp := &grpc.CreateWorkflowResponse{}
	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.Revision = bytedata.ConvertRevisionToGrpcRevision(revision)
	resp.Revision.Source = data

	return resp, nil
}

func (flow *flow) UpdateWorkflow(ctx context.Context, req *grpc.UpdateWorkflowRequest) (*grpc.UpdateWorkflowResponse, error) {
	// This is being called by the frontend when a user changes a workflow via a UI and press save button.

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
	if file.Typ != filestore.FileTypeWorkflow && file.Typ != filestore.FileTypeService {
		return nil, status.Error(codes.InvalidArgument, "file type is not workflow or service")
	}
	revision, err := tx.FileStore().ForFile(file).GetCurrentRevision(ctx)
	if err != nil {
		return nil, err
	}
	newRevision, err := tx.FileStore().ForFile(file).CreateRevision(ctx, "", req.GetSource())
	if err != nil {
		return nil, err
	}
	// delete the previous revision.
	err = tx.FileStore().ForRevision(revision).Delete(ctx)
	if err != nil {
		return nil, err
	}

	data, err := tx.FileStore().ForRevision(newRevision).GetData(ctx)
	if err != nil {
		return nil, err
	}
	_, router, err := getRouter(ctx, tx, file)
	if err != nil {
		return nil, err
	}

	if file.Typ == filestore.FileTypeWorkflow {
		err = flow.configureWorkflowStarts(ctx, tx, ns.ID, file, router, true)
		if err != nil {
			return nil, err
		}

		err = flow.placeholdSecrets(ctx, tx, ns.Name, file)
		if err != nil {
			return nil, err
		}
		err = flow.pBus.Publish(pubsub.WorkflowUpdate, file.Path)
		if err != nil {
			flow.sugar.Error("pubsub publish", "error", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	if file.Typ == filestore.FileTypeService {
		err = flow.pBus.Publish(pubsub.ServiceUpdate, file.Path)
		if err != nil {
			flow.sugar.Error("pubsub publish", "error", err)
		}
	}

	var resp grpc.UpdateWorkflowResponse

	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.Revision = bytedata.ConvertRevisionToGrpcRevision(newRevision)
	resp.Revision.Source = data

	return &resp, nil
}

func (flow *flow) SaveHead(ctx context.Context, req *grpc.SaveHeadRequest) (*grpc.SaveHeadResponse, error) {
	// This is being called by the UI when a user clicks create revision button.

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
	if file.Typ != filestore.FileTypeWorkflow {
		return nil, status.Error(codes.InvalidArgument, "file type is not workflow")
	}
	revision, err := tx.FileStore().ForFile(file).GetCurrentRevision(ctx)
	if err != nil {
		return nil, err
	}
	dataReader, err := tx.FileStore().ForRevision(revision).GetData(ctx)
	if err != nil {
		return nil, err
	}
	_, err = tx.FileStore().ForFile(file).CreateRevision(ctx, "", dataReader)
	if err != nil {
		return nil, err
	}
	data, err := tx.FileStore().ForRevision(revision).GetData(ctx)
	if err != nil {
		return nil, err
	}

	_, router, err := getRouter(ctx, tx, file)
	if err != nil {
		return nil, err
	}

	err = flow.configureWorkflowStarts(ctx, tx, ns.ID, file, router, true)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	err = flow.pBus.Publish(pubsub.WorkflowUpdate, file.Path)
	if err != nil {
		flow.sugar.Error("pubsub publish", "error", err)
	}

	var resp grpc.SaveHeadResponse

	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.Revision = bytedata.ConvertRevisionToGrpcRevision(revision)
	resp.Revision.Source = data

	return &resp, nil
}

func (flow *flow) DiscardHead(ctx context.Context, req *grpc.DiscardHeadRequest) (*grpc.DiscardHeadResponse, error) {
	// This is being called by the UI when a user clicks 'revert' button.

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
	if file.Typ != filestore.FileTypeWorkflow {
		return nil, status.Error(codes.InvalidArgument, "file type is not workflow")
	}

	// Discarding head is basically reverting to the before latest revision.

	revs, err := tx.FileStore().ForFile(file).GetAllRevisions(ctx)
	if err != nil {
		return nil, err
	}

	var currentRev *filestore.Revision
	var beforeLatestRev *filestore.Revision
	if !revs[0].IsCurrent {
		beforeLatestRev = revs[0]
	} else {
		beforeLatestRev = revs[1]
	}
	for _, rev := range revs {
		if rev.IsCurrent {
			currentRev = rev
			continue
		}
		if rev.CreatedAt.Compare(beforeLatestRev.CreatedAt) >= 0 {
			beforeLatestRev = rev
		}
	}
	dataReader, err := tx.FileStore().ForRevision(beforeLatestRev).GetData(ctx)
	if err != nil {
		return nil, err
	}
	newRev, err := tx.FileStore().ForFile(file).CreateRevision(ctx, "", dataReader)
	if err != nil {
		return nil, err
	}
	// delete the old current revision.
	err = tx.FileStore().ForRevision(currentRev).Delete(ctx)
	if err != nil {
		return nil, err
	}
	data, err := tx.FileStore().ForRevision(newRev).GetData(ctx)
	if err != nil {
		return nil, err
	}
	_, router, err := getRouter(ctx, tx, file)
	if err != nil {
		return nil, err
	}
	err = flow.configureWorkflowStarts(ctx, tx, ns.ID, file, router, true)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	err = flow.pBus.Publish(pubsub.WorkflowUpdate, file.Path)
	if err != nil {
		flow.sugar.Error("pubsub publish", "error", err)
	}

	var resp grpc.DiscardHeadResponse

	resp.Namespace = ns.Name
	resp.Node = bytedata.ConvertFileToGrpcNode(file)
	resp.Revision = bytedata.ConvertRevisionToGrpcRevision(newRev)
	resp.Revision.Source = data

	return &resp, nil
}

func (flow *flow) ToggleWorkflow(ctx context.Context, req *grpc.ToggleWorkflowRequest) (*emptypb.Empty, error) {
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

	annotations, router, err := getRouter(ctx, tx, file)
	if err != nil {
		return nil, err
	}

	router.Enabled = !router.Enabled

	annotations.Data = annotations.Data.SetEntry(routerAnnotationKey, router.Marshal())

	err = tx.DataStore().FileAnnotations().Set(ctx, annotations)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (flow *flow) SetWorkflowEventLogging(ctx context.Context, req *grpc.SetWorkflowEventLoggingRequest) (*emptypb.Empty, error) {
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

	annotations, err := tx.DataStore().FileAnnotations().Get(ctx, file.ID)

	if errors.Is(err, core.ErrFileAnnotationsNotSet) {
		annotations = &core.FileAnnotations{
			FileID: file.ID,
			Data:   map[string]string{},
		}
	} else if err != nil {
		return nil, err
	}

	annotations.Data = annotations.Data.SetEntry("workflow_log_event_key", req.GetLogger())

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
