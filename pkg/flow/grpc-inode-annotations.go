package flow

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (flow *flow) NodeAnnotation(ctx context.Context, req *grpc.NodeAnnotationRequest) (*grpc.NodeAnnotationResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	nsc := flow.db.Namespace

	d, err := flow.traverseToInodeAnnotation(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetKey())
	if err != nil {
		return nil, err
	}

	var resp grpc.NodeAnnotationResponse

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

func (flow *flow) NodeAnnotationParcels(req *grpc.NodeAnnotationRequest, srv grpc.Flow_NodeAnnotationParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	nsc := flow.db.Namespace

	d, err := flow.traverseToInodeAnnotation(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetKey())
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(d.annotation.Data)

	for {

		resp := new(grpc.NodeAnnotationResponse)

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

func (flow *flow) NodeAnnotations(ctx context.Context, req *grpc.NodeAnnotationsRequest) (*grpc.NodeAnnotationsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	d, err := flow.traverseToInode(ctx, flow.db.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	query := d.ino.QueryAnnotations()

	results, pi, err := paginate[*ent.AnnotationQuery, *ent.Annotation](ctx, req.Pagination, query, annotationsOrderings, annotationsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.NodeAnnotationsResponse)
	resp.Namespace = d.namespace()
	resp.Annotations = new(grpc.Annotations)
	resp.Annotations.PageInfo = pi

	err = atob(results, &resp.Annotations.Results)
	if err != nil {
		return nil, err
	}

	return resp, nil

}

func (flow *flow) NodeAnnotationsStream(req *grpc.NodeAnnotationsRequest, srv grpc.Flow_NodeAnnotationsStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	d, err := flow.traverseToInode(ctx, flow.db.Namespace, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeInodeAnnotations(d.ino)
	defer flow.cleanup(sub.Close)

resend:

	query := d.ino.QueryAnnotations()

	results, pi, err := paginate[*ent.AnnotationQuery, *ent.Annotation](ctx, req.Pagination, query, annotationsOrderings, annotationsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.NodeAnnotationsResponse)
	resp.Namespace = d.namespace()
	resp.Annotations = new(grpc.Annotations)
	resp.Annotations.PageInfo = pi

	err = atob(results, &resp.Annotations.Results)
	if err != nil {
		return err
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

func (flow *flow) SetNodeAnnotation(ctx context.Context, req *grpc.SetNodeAnnotationRequest) (*grpc.SetNodeAnnotationResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	annotationc := tx.Annotation

	d, err := flow.traverseToInode(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	var annotation *ent.Annotation

	key := req.GetKey()

	// var newVar bool
	annotation /*newVar*/, _, err = flow.SetAnnotation(ctx, annotationc, d.ino, key, req.GetData())
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// if newVar {
	// 	flow.logToInode(ctx, time.Now(), d, "Created inode annotation '%s'.", key)
	// } else {
	// 	flow.logToInode(ctx, time.Now(), d, "Updated inode annotation '%s'.", key)
	// }

	flow.pubsub.NotifyInodeAnnotations(d.ino)

	var resp grpc.SetNodeAnnotationResponse

	resp.Namespace = d.ns().Name
	resp.Path = d.path
	resp.Key = key
	resp.CreatedAt = timestamppb.New(annotation.CreatedAt)
	resp.UpdatedAt = timestamppb.New(annotation.UpdatedAt)
	resp.Checksum = annotation.Hash
	resp.TotalSize = int64(annotation.Size)

	return &resp, nil

}

func (flow *flow) SetNodeAnnotationParcels(srv grpc.Flow_SetNodeAnnotationParcelsServer) error {

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

	d, err := flow.traverseToInode(ctx, nsc, namespace, path)
	if err != nil {
		return err
	}

	var annotation *ent.Annotation

	// var newVar bool
	annotation /*newVar*/, _, err = flow.SetAnnotation(ctx, annotationc, d.ino, key, buf.Bytes())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	// if newVar {
	// 	flow.logToInode(ctx, time.Now(), d, "Created inode annotation '%s'.", key)
	// } else {
	// 	flow.logToInode(ctx, time.Now(), d, "Updated inode annotation '%s'.", key)
	// }

	flow.pubsub.NotifyInodeAnnotations(d.ino)

	var resp grpc.SetNodeAnnotationResponse

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

func (flow *flow) DeleteNodeAnnotation(ctx context.Context, req *grpc.DeleteNodeAnnotationRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace

	d, err := flow.traverseToInodeAnnotation(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetKey())
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

	// flow.logToInode(ctx, time.Now(), d.wfData, "Deleted inode annotation '%s'.", d.annotation.Name)
	flow.pubsub.NotifyInodeAnnotations(d.ino)

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) RenameNodeAnnotation(ctx context.Context, req *grpc.RenameNodeAnnotationRequest) (*grpc.RenameNodeAnnotationResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToInodeAnnotation(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetOld())
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

	// flow.logToInode(ctx, time.Now(), d.wfData, "Renamed inode annotation from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyInodeAnnotations(d.ino)

	var resp grpc.RenameNodeAnnotationResponse

	resp.Checksum = d.annotation.Hash
	resp.CreatedAt = timestamppb.New(d.annotation.CreatedAt)
	resp.Key = annotation.Name
	resp.Namespace = d.ns().Name
	resp.TotalSize = int64(d.annotation.Size)
	resp.UpdatedAt = timestamppb.New(d.annotation.UpdatedAt)

	return &resp, nil

}
