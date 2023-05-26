package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
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

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	fStore, _, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	ref := req.GetRef()
	if ref == "" {
		ref = filestore.Latest
	}
	revision, err := fStore.ForFile(file).GetRevision(ctx, ref)
	if err != nil {
		return nil, err
	}

	dataReader, err := fStore.ForRevision(revision).GetData(ctx)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(dataReader)
	if err != nil {
		return nil, err
	}
	if err = commit(ctx); err != nil {
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

func (flow *flow) CreateWorkflow(ctx context.Context, req *grpc.CreateWorkflowRequest) (*grpc.CreateWorkflowResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}
	if filepath.Ext(req.GetPath()) != ".yaml" && filepath.Ext(req.GetPath()) != ".yml" {
		return nil, status.Error(codes.InvalidArgument, "workflow name should have either .yaml or .yaml extension")
	}

	fStore, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, revision, err := fStore.ForRootID(ns.ID).CreateFile(ctx, req.GetPath(), filestore.FileTypeWorkflow, bytes.NewReader(req.GetSource()))
	if err != nil {
		return nil, err
	}

	dataReader, err := fStore.ForRevision(revision).GetData(ctx)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(dataReader)
	if err != nil {
		return nil, err
	}

	workflow := new(model.Workflow)
	err = workflow.Load(data)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, router, err := getRouter(ctx, fStore, store.FileAnnotations(), file)
	if err != nil {
		return nil, err
	}

	err = flow.configureWorkflowStarts(ctx, fStore, store.FileAnnotations(), ns.ID, file, router, true)
	if err != nil {
		return nil, err
	}

	if err = commit(ctx); err != nil {
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

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	fStore, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	if file.Typ != filestore.FileTypeWorkflow {
		return nil, status.Error(codes.InvalidArgument, "file type is not workflow")
	}
	revision, err := fStore.ForFile(file).GetCurrentRevision(ctx)
	if err != nil {
		return nil, err
	}
	newRevision, err := fStore.ForFile(file).CreateRevision(ctx, "", bytes.NewReader(req.GetSource()))
	if err != nil {
		return nil, err
	}
	// delete the previous revision.
	err = fStore.ForRevision(revision).Delete(ctx)
	if err != nil {
		return nil, err
	}

	dataReader, err := fStore.ForRevision(newRevision).GetData(ctx)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(dataReader)
	if err != nil {
		return nil, err
	}

	_, router, err := getRouter(ctx, fStore, store.FileAnnotations(), file)
	if err != nil {
		return nil, err
	}

	err = flow.configureWorkflowStarts(ctx, fStore, store.FileAnnotations(), ns.ID, file, router, true)
	if err != nil {
		return nil, err
	}

	if err = commit(ctx); err != nil {
		return nil, err
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

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	fStore, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	if file.Typ != filestore.FileTypeWorkflow {
		return nil, status.Error(codes.InvalidArgument, "file type is not workflow")
	}
	revision, err := fStore.ForFile(file).GetCurrentRevision(ctx)
	if err != nil {
		return nil, err
	}
	dataReader, err := fStore.ForRevision(revision).GetData(ctx)
	if err != nil {
		return nil, err
	}
	_, err = fStore.ForFile(file).CreateRevision(ctx, "", dataReader)
	if err != nil {
		return nil, err
	}
	dataReader, err = fStore.ForRevision(revision).GetData(ctx)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(dataReader)
	if err != nil {
		return nil, err
	}

	_, router, err := getRouter(ctx, fStore, store.FileAnnotations(), file)
	if err != nil {
		return nil, err
	}

	err = flow.configureWorkflowStarts(ctx, fStore, store.FileAnnotations(), ns.ID, file, router, true)
	if err != nil {
		return nil, err
	}

	if err = commit(ctx); err != nil {
		return nil, err
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

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	fStore, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}
	if file.Typ != filestore.FileTypeWorkflow {
		return nil, status.Error(codes.InvalidArgument, "file type is not workflow")
	}

	// Discarding head is basically reverting to the before latest revision.

	revs, err := fStore.ForFile(file).GetAllRevisions(ctx)
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
	dataReader, err := fStore.ForRevision(beforeLatestRev).GetData(ctx)
	if err != nil {
		return nil, err
	}
	newRev, err := fStore.ForFile(file).CreateRevision(ctx, "", dataReader)
	if err != nil {
		return nil, err
	}
	// delete the old current revision.
	err = fStore.ForRevision(currentRev).Delete(ctx)
	if err != nil {
		return nil, err
	}
	dataReader, err = fStore.ForRevision(newRev).GetData(ctx)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(dataReader)
	if err != nil {
		return nil, err
	}
	_, router, err := getRouter(ctx, fStore, store.FileAnnotations(), file)
	if err != nil {
		return nil, err
	}
	err = flow.configureWorkflowStarts(ctx, fStore, store.FileAnnotations(), ns.ID, file, router, true)
	if err != nil {
		return nil, err
	}

	if err = commit(ctx); err != nil {
		return nil, err
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

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	fStore, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	annotations, router, err := getRouter(ctx, fStore, store.FileAnnotations(), file)
	if err != nil {
		return nil, err
	}

	router.Enabled = !router.Enabled

	annotations.Data = annotations.Data.SetEntry(routerAnnotationKey, router.Marshal())

	err = store.FileAnnotations().Set(ctx, annotations)
	if err != nil {
		return nil, err
	}

	err = commit(ctx)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (flow *flow) SetWorkflowEventLogging(ctx context.Context, req *grpc.SetWorkflowEventLoggingRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	fStore, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	file, err := fStore.ForRootID(ns.ID).GetFile(ctx, req.GetPath())
	if err != nil {
		return nil, err
	}

	annotations, err := store.FileAnnotations().Get(ctx, file.ID)

	if errors.Is(err, core.ErrFileAnnotationsNotSet) {
		annotations = &core.FileAnnotations{
			FileID: file.ID,
			Data:   map[string]string{},
		}
	} else if err != nil {
		return nil, err
	}

	annotations.Data = annotations.Data.SetEntry("workflow_log_event_key", req.GetLogger())

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
