package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (flow *flow) WorkflowAnnotation(ctx context.Context, req *grpc.WorkflowAnnotationRequest) (*grpc.WorkflowAnnotationResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	nsc := flow.db.Namespace

	d, err := flow.traverseToWorkflowAnnotation(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetKey())
	if err != nil {
		return nil, err
	}

	var resp grpc.WorkflowAnnotationResponse

	resp.Namespace = d.ns().Name
	resp.Path = d.path
	resp.Key = d.annotation.Name
	resp.CreatedAt = timestamppb.New(d.annotation.CreatedAt)
	resp.UpdatedAt = timestamppb.New(d.annotation.UpdatedAt)
	resp.Checksum = d.annotation.Hash
	resp.TotalSize = int64(d.annotation.Size)

	if resp.TotalSize > parcelSize {
		return nil, status.Error(codes.ResourceExhausted, "annotation too large to return without using the parcelling API")
	}

	resp.Data = d.annotation.Data

	return &resp, nil

}

func (flow *flow) WorkflowAnnotationParcels(req *grpc.WorkflowAnnotationRequest, srv grpc.Flow_WorkflowAnnotationParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	nsc := flow.db.Namespace

	d, err := flow.traverseToWorkflowAnnotation(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetKey())
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(d.annotation.Data)

	for {

		resp := new(grpc.WorkflowAnnotationResponse)

		resp.Namespace = d.ns().Name
		resp.Path = d.path
		resp.Key = d.annotation.Name
		resp.CreatedAt = timestamppb.New(d.annotation.CreatedAt)
		resp.UpdatedAt = timestamppb.New(d.annotation.UpdatedAt)
		resp.Checksum = d.annotation.Hash
		resp.TotalSize = int64(d.annotation.Size)

		buf := new(bytes.Buffer)
		k, err := io.CopyN(buf, rdr, parcelSize)
		if err != nil {

			if errors.Is(err, io.EOF) {
				err = nil
			}

			if err == nil && k == 0 {
				if resp.TotalSize == 0 {
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

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.AnnotationPaginateOption{}
	opts = append(opts, annotationsOrder(p)...)
	opts = append(opts, annotationsFilter(p)...)

	nsc := flow.db.Namespace

	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	query := d.wf.QueryAnnotations()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.WorkflowAnnotationsResponse

	resp.Namespace = d.ns().Name
	resp.Path = d.path

	err = atob(cx, &resp.Annotations)
	if err != nil {
		return nil, err
	}

	for i := range cx.Edges {

		edge := cx.Edges[i]
		annotation := edge.Node

		v := resp.Annotations.Edges[i].Node
		v.Checksum = annotation.Hash
		v.CreatedAt = timestamppb.New(annotation.CreatedAt)
		v.Size = int64(annotation.Size)
		v.UpdatedAt = timestamppb.New(annotation.UpdatedAt)

	}

	return &resp, nil

}

func (flow *flow) WorkflowAnnotationsStream(req *grpc.WorkflowAnnotationsRequest, srv grpc.Flow_WorkflowAnnotationsStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	opts := []ent.AnnotationPaginateOption{}
	opts = append(opts, annotationsOrder(p)...)
	opts = append(opts, annotationsFilter(p)...)

	nsc := flow.db.Namespace
	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflowAnnotations(d.wf)
	defer flow.cleanup(sub.Close)

resend:

	query := d.wf.QueryAnnotations()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	resp := new(grpc.WorkflowAnnotationsResponse)

	resp.Namespace = d.ns().Name
	resp.Path = d.path

	err = atob(cx, &resp.Annotations)
	if err != nil {
		return err
	}

	for i := range cx.Edges {

		edge := cx.Edges[i]
		annotation := edge.Node

		v := resp.Annotations.Edges[i].Node
		v.Checksum = annotation.Hash
		v.CreatedAt = timestamppb.New(annotation.CreatedAt)
		v.Size = int64(annotation.Size)
		v.UpdatedAt = timestamppb.New(annotation.UpdatedAt)

	}

	nhash = checksum(resp)
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

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	annotationc := tx.Annotation

	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	var annotation *ent.Annotation

	key := req.GetKey()

	var newVar bool
	annotation, newVar, err = flow.SetAnnotation(ctx, annotationc, d.wf, key, req.GetData())
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if newVar {
		flow.logToWorkflow(ctx, time.Now(), d, "Created workflow annotation '%s'.", key)
	} else {
		flow.logToWorkflow(ctx, time.Now(), d, "Updated workflow annotation '%s'.", key)
	}

	flow.pubsub.NotifyWorkflowAnnotations(d.wf)

	var resp grpc.SetWorkflowAnnotationResponse

	resp.Namespace = d.ns().Name
	resp.Path = d.path
	resp.Key = key
	resp.CreatedAt = timestamppb.New(annotation.CreatedAt)
	resp.UpdatedAt = timestamppb.New(annotation.UpdatedAt)
	resp.Checksum = annotation.Hash
	resp.TotalSize = int64(annotation.Size)

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

	totalSize := int(req.GetTotalSize())

	buf := new(bytes.Buffer)

	for {

		_, err = io.Copy(buf, bytes.NewReader(req.Data))
		if err != nil {
			return err
		}

		if req.TotalSize <= 0 {
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

		if req.TotalSize <= 0 {
			if buf.Len() >= totalSize {
				break
			}
		} else {
			if req == nil {
				break
			}
		}

		if int(req.GetTotalSize()) != totalSize {
			return errors.New("totalSize changed mid stream")
		}

	}

	if buf.Len() > totalSize {
		return errors.New("received more data than expected")
	}

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	annotationc := tx.Annotation

	d, err := flow.traverseToWorkflow(ctx, nsc, namespace, path)
	if err != nil {
		return err
	}

	var annotation *ent.Annotation

	var newVar bool
	annotation, newVar, err = flow.SetAnnotation(ctx, annotationc, d.wf, key, buf.Bytes())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newVar {
		flow.logToWorkflow(ctx, time.Now(), d, "Created workflow annotation '%s'.", key)
	} else {
		flow.logToWorkflow(ctx, time.Now(), d, "Updated workflow annotation '%s'.", key)
	}

	flow.pubsub.NotifyWorkflowAnnotations(d.wf)

	var resp grpc.SetWorkflowAnnotationResponse

	resp.Namespace = d.ns().Name
	resp.Path = d.path
	resp.Key = key
	resp.CreatedAt = timestamppb.New(annotation.CreatedAt)
	resp.UpdatedAt = timestamppb.New(annotation.UpdatedAt)
	resp.Checksum = annotation.Hash
	resp.TotalSize = int64(annotation.Size)

	err = srv.SendAndClose(&resp)
	if err != nil {
		return err
	}

	return nil

}

func (flow *flow) DeleteWorkflowAnnotation(ctx context.Context, req *grpc.DeleteWorkflowAnnotationRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace

	d, err := flow.traverseToWorkflowAnnotation(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetKey())
	if err != nil {
		return nil, err
	}

	annotationc := tx.Annotation

	err = annotationc.DeleteOne(d.annotation).Exec(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), d.wfData, "Deleted workflow annotation '%s'.", d.annotation.Name)
	flow.pubsub.NotifyWorkflowAnnotations(d.wf)

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) RenameWorkflowAnnotation(ctx context.Context, req *grpc.RenameWorkflowAnnotationRequest) (*grpc.RenameWorkflowAnnotationResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToWorkflowAnnotation(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetOld())
	if err != nil {
		return nil, err
	}

	annotation, err := d.annotation.Update().SetName(req.GetNew()).Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), d.wfData, "Renamed workflow annotation from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyWorkflowAnnotations(d.wf)

	var resp grpc.RenameWorkflowAnnotationResponse

	resp.Checksum = d.annotation.Hash
	resp.CreatedAt = timestamppb.New(d.annotation.CreatedAt)
	resp.Key = annotation.Name
	resp.Namespace = d.ns().Name
	resp.TotalSize = int64(d.annotation.Size)
	resp.UpdatedAt = timestamppb.New(d.annotation.UpdatedAt)

	return &resp, nil

}
