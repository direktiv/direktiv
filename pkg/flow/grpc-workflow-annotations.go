package flow

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entnote "github.com/direktiv/direktiv/pkg/flow/ent/annotation"
	entwf "github.com/direktiv/direktiv/pkg/flow/ent/workflow"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (flow *flow) traverseToWorkflowAnnotation(ctx context.Context, namespace, path, key string) (*database.CacheData, *database.Annotation, error) {
	cached, err := flow.traverseToWorkflow(ctx, namespace, path)
	if err != nil {
		return nil, nil, err
	}

	annotation, err := flow.database.WorkflowAnnotation(ctx, cached.Workflow.ID, key)
	if err != nil {
		return nil, nil, err
	}

	return cached, annotation, nil
}

func (flow *flow) WorkflowAnnotation(ctx context.Context, req *grpc.WorkflowAnnotationRequest) (*grpc.WorkflowAnnotationResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, annotation, err := flow.traverseToWorkflowAnnotation(ctx, req.GetNamespace(), req.GetPath(), req.GetKey())
	if err != nil {
		return nil, err
	}

	var resp grpc.WorkflowAnnotationResponse

	resp.Namespace = cached.Namespace.Name
	resp.Path = cached.Path()
	resp.Key = annotation.Name
	resp.CreatedAt = timestamppb.New(annotation.CreatedAt)
	resp.UpdatedAt = timestamppb.New(annotation.UpdatedAt)
	resp.Checksum = annotation.Hash
	resp.Size = int64(annotation.Size)
	resp.MimeType = annotation.MimeType

	if resp.Size > parcelSize {
		return nil, status.Error(codes.ResourceExhausted, "annotation too large to return without using the parcelling API")
	}

	resp.Data = annotation.Data

	return &resp, nil
}

func (flow *flow) WorkflowAnnotationParcels(req *grpc.WorkflowAnnotationRequest, srv grpc.Flow_WorkflowAnnotationParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	cached, annotation, err := flow.traverseToWorkflowAnnotation(ctx, req.GetNamespace(), req.GetPath(), req.GetKey())
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(annotation.Data)

	for {

		resp := new(grpc.WorkflowAnnotationResponse)

		resp.Namespace = cached.Namespace.Name
		resp.Path = cached.Path()
		resp.Key = annotation.Name
		resp.CreatedAt = timestamppb.New(annotation.CreatedAt)
		resp.UpdatedAt = timestamppb.New(annotation.UpdatedAt)
		resp.Checksum = annotation.Hash
		resp.Size = int64(annotation.Size)
		resp.MimeType = annotation.MimeType

		buf := new(bytes.Buffer)
		k, err := io.CopyN(buf, rdr, parcelSize)
		if err != nil {

			if errors.Is(err, io.EOF) {
				err = nil
			}

			if err == nil && k == 0 {
				if resp.Size == 0 {
					resp.Data = buf.Bytes()
					err = srv.Send(resp)
					if err != nil {
						return err
					}
				}
				return nil
			}

			if err != nil {
				return err
			}

		}

		resp.Data = buf.Bytes()

		err = srv.Send(resp)
		if err != nil {
			return err
		}

	}
}

func (flow *flow) WorkflowAnnotations(ctx context.Context, req *grpc.WorkflowAnnotationsRequest) (*grpc.WorkflowAnnotationsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, err := flow.traverseToWorkflow(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	query := clients.Annotation.Query().Where(entnote.HasWorkflowWith(entwf.ID(cached.Workflow.ID)))

	results, pi, err := paginate[*ent.AnnotationQuery, *ent.Annotation](ctx, req.Pagination, query, annotationsOrderings, annotationsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.WorkflowAnnotationsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Annotations = new(grpc.Annotations)
	resp.Annotations.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Annotations.Results)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (flow *flow) WorkflowAnnotationsStream(req *grpc.WorkflowAnnotationsRequest, srv grpc.Flow_WorkflowAnnotationsStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	cached, err := flow.traverseToWorkflow(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflowAnnotations(cached)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(ctx)

	query := clients.Annotation.Query().Where(entnote.HasWorkflowWith(entwf.ID(cached.Workflow.ID)))

	results, pi, err := paginate[*ent.AnnotationQuery, *ent.Annotation](ctx, req.Pagination, query, annotationsOrderings, annotationsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.WorkflowAnnotationsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Annotations = new(grpc.Annotations)
	resp.Annotations.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Annotations.Results)
	if err != nil {
		return err
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

func (flow *flow) SetWorkflowAnnotation(ctx context.Context, req *grpc.SetWorkflowAnnotationRequest) (*grpc.SetWorkflowAnnotationResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToWorkflow(tctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	var annotation *ent.Annotation

	key := req.GetKey()

	var newVar bool
	annotation, newVar, err = flow.SetAnnotation(tctx, &entWorkflowAnnotationQuerier{clients: flow.edb.Clients(tctx), cached: cached}, key, req.GetMimeType(), req.GetData())
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if newVar {
		flow.logger.Infof(ctx, cached.Workflow.ID, cached.GetAttributes(recipient.Workflow), "Created workflow annotation '%s'.", key)
	} else {
		flow.logger.Infof(ctx, cached.Workflow.ID, cached.GetAttributes(recipient.Workflow), "Updated workflow annotation '%s'.", key)
	}

	flow.pubsub.NotifyWorkflowAnnotations(cached.Workflow)

	var resp grpc.SetWorkflowAnnotationResponse

	resp.Namespace = cached.Namespace.Name
	resp.Path = cached.Path()
	resp.Key = key
	resp.CreatedAt = timestamppb.New(annotation.CreatedAt)
	resp.UpdatedAt = timestamppb.New(annotation.UpdatedAt)
	resp.Checksum = annotation.Hash
	resp.Size = int64(annotation.Size)
	resp.MimeType = annotation.MimeType

	return &resp, nil
}

func (flow *flow) SetWorkflowAnnotationParcels(srv grpc.Flow_SetWorkflowAnnotationParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	req, err := srv.Recv()
	if err != nil {
		return err
	}

	namespace := req.GetNamespace()
	path := req.GetPath()
	key := req.GetKey()

	totalSize := int(req.GetSize())

	buf := new(bytes.Buffer)

	for {

		_, err = io.Copy(buf, bytes.NewReader(req.Data))
		if err != nil {
			return err
		}

		if req.Size <= 0 {
			if buf.Len() >= totalSize {
				break
			}
		}

		req, err = srv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		if req.Size <= 0 {
			if buf.Len() >= totalSize {
				break
			}
		} else {
			if req == nil {
				break
			}
		}

		if int(req.GetSize()) != totalSize {
			return errors.New("totalSize changed mid stream")
		}

	}

	if buf.Len() > totalSize {
		return errors.New("received more data than expected")
	}

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	cached, err := flow.traverseToWorkflow(tctx, namespace, path)
	if err != nil {
		return err
	}

	var annotation *ent.Annotation

	var newVar bool
	annotation, newVar, err = flow.SetAnnotation(tctx, &entWorkflowAnnotationQuerier{clients: flow.edb.Clients(tctx), cached: cached}, key, req.GetMimeType(), buf.Bytes())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newVar {
		flow.logger.Infof(ctx, cached.Workflow.ID, cached.GetAttributes(recipient.Workflow), "Created workflow annotation '%s'.", key)
	} else {
		flow.logger.Infof(ctx, cached.Workflow.ID, cached.GetAttributes(recipient.Workflow), "Updated workflow annotation '%s'.", key)
	}

	flow.pubsub.NotifyWorkflowAnnotations(cached.Workflow)

	var resp grpc.SetWorkflowAnnotationResponse

	resp.Namespace = cached.Namespace.Name
	resp.Path = cached.Path()
	resp.Key = key
	resp.CreatedAt = timestamppb.New(annotation.CreatedAt)
	resp.UpdatedAt = timestamppb.New(annotation.UpdatedAt)
	resp.Checksum = annotation.Hash
	resp.Size = int64(annotation.Size)
	resp.MimeType = annotation.MimeType

	err = srv.SendAndClose(&resp)
	if err != nil {
		return err
	}

	return nil
}

func (flow *flow) DeleteWorkflowAnnotation(ctx context.Context, req *grpc.DeleteWorkflowAnnotationRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, annotation, err := flow.traverseToWorkflowAnnotation(tctx, req.GetNamespace(), req.GetPath(), req.GetKey())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tctx)

	err = clients.Annotation.DeleteOneID(annotation.ID).Exec(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, cached.Workflow.ID, cached.GetAttributes(recipient.Workflow), "Deleted workflow annotation '%s'.", annotation.Name)
	flow.pubsub.NotifyWorkflowAnnotations(cached.Workflow)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) RenameWorkflowAnnotation(ctx context.Context, req *grpc.RenameWorkflowAnnotationRequest) (*grpc.RenameWorkflowAnnotationResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, annotation, err := flow.traverseToWorkflowAnnotation(tctx, req.GetNamespace(), req.GetPath(), req.GetOld())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tctx)

	x, err := clients.Annotation.UpdateOneID(annotation.ID).SetName(req.GetNew()).Save(ctx)
	if err != nil {
		return nil, err
	}

	annotation.Name = x.Name

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, cached.Workflow.ID, cached.GetAttributes(recipient.Workflow), "Renamed workflow annotation from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyWorkflowAnnotations(cached.Workflow)

	var resp grpc.RenameWorkflowAnnotationResponse

	resp.Checksum = annotation.Hash
	resp.CreatedAt = timestamppb.New(annotation.CreatedAt)
	resp.Key = annotation.Name
	resp.Namespace = cached.Namespace.Name
	resp.Size = int64(annotation.Size)
	resp.UpdatedAt = timestamppb.New(annotation.UpdatedAt)
	resp.MimeType = annotation.MimeType

	return &resp, nil
}
