package flow

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entnote "github.com/direktiv/direktiv/pkg/flow/ent/annotation"
	entino "github.com/direktiv/direktiv/pkg/flow/ent/inode"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (flow *flow) traverseToInodeAnnotation(ctx context.Context, tx Transaction, namespace, path, key string) (*CacheData, *Annotation, error) {

	cached := new(CacheData)

	err := flow.database.NamespaceByName(ctx, tx, cached, namespace)
	if err != nil {
		return nil, nil, err
	}

	err = flow.database.InodeByPath(ctx, tx, cached, path)
	if err != nil {
		return nil, nil, err
	}

	annotation, err := flow.database.InodeAnnotation(ctx, tx, cached.Inode().ID, key)
	if err != nil {
		return nil, nil, err
	}

	return cached, annotation, nil

}

func (flow *flow) NodeAnnotation(ctx context.Context, req *grpc.NodeAnnotationRequest) (*grpc.NodeAnnotationResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, annotation, err := flow.traverseToInodeAnnotation(ctx, nil, req.GetNamespace(), req.GetPath(), req.GetKey())
	if err != nil {
		return nil, err
	}

	var resp grpc.NodeAnnotationResponse

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

func (flow *flow) NodeAnnotationParcels(req *grpc.NodeAnnotationRequest, srv grpc.Flow_NodeAnnotationParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	cached, annotation, err := flow.traverseToInodeAnnotation(ctx, nil, req.GetNamespace(), req.GetPath(), req.GetKey())
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(annotation.Data)

	for {

		resp := new(grpc.NodeAnnotationResponse)

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

func (flow *flow) NodeAnnotations(ctx context.Context, req *grpc.NodeAnnotationsRequest) (*grpc.NodeAnnotationsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, err := flow.traverseToInode(ctx, nil, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	clients := flow.entClients(nil)

	query := clients.Annotation.Query().Where(entnote.HasInodeWith(entino.ID(cached.Inode().ID)))

	results, pi, err := paginate[*ent.AnnotationQuery, *ent.Annotation](ctx, req.Pagination, query, annotationsOrderings, annotationsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.NodeAnnotationsResponse)
	resp.Namespace = cached.Namespace.Name
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

	cached, err := flow.traverseToInode(ctx, nil, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeInodeAnnotations(cached)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.entClients(nil)

	query := clients.Annotation.Query().Where(entnote.HasInodeWith(entino.ID(cached.Inode().ID)))

	results, pi, err := paginate[*ent.AnnotationQuery, *ent.Annotation](ctx, req.Pagination, query, annotationsOrderings, annotationsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.NodeAnnotationsResponse)
	resp.Namespace = cached.Namespace.Name
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

	cached, err := flow.traverseToInode(ctx, tx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	var annotation *ent.Annotation

	key := req.GetKey()

	annotation, _, err = flow.SetAnnotation(ctx, tx.Annotation, &entInodeAnnotationQuerier{clients: flow.entClients(tx), cached: cached}, key, req.GetMimeType(), req.GetData())
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.pubsub.NotifyInodeAnnotations(cached.Inode())

	var resp grpc.SetNodeAnnotationResponse

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

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	cached, err := flow.traverseToInode(ctx, tx, namespace, path)
	if err != nil {
		return err
	}

	var annotation *ent.Annotation

	annotation, _, err = flow.SetAnnotation(ctx, tx.Annotation, &entInodeAnnotationQuerier{clients: flow.entClients(tx), cached: cached}, key, req.GetMimeType(), buf.Bytes())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	flow.pubsub.NotifyInodeAnnotations(cached.Inode())

	var resp grpc.SetNodeAnnotationResponse

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

func (flow *flow) DeleteNodeAnnotation(ctx context.Context, req *grpc.DeleteNodeAnnotationRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, annotation, err := flow.traverseToInodeAnnotation(ctx, tx, req.GetNamespace(), req.GetPath(), req.GetKey())
	if err != nil {
		return nil, err
	}

	annotationc := tx.Annotation

	err = annotationc.DeleteOneID(annotation.ID).Exec(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.pubsub.NotifyInodeAnnotations(cached.Inode())

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

	cached, annotation, err := flow.traverseToInodeAnnotation(ctx, tx, req.GetNamespace(), req.GetPath(), req.GetOld())
	if err != nil {
		return nil, err
	}

	anno, err := tx.Annotation.UpdateOneID(annotation.ID).SetName(req.GetNew()).Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.pubsub.NotifyInodeAnnotations(cached.Inode())

	var resp grpc.RenameNodeAnnotationResponse

	resp.Checksum = anno.Hash
	resp.CreatedAt = timestamppb.New(anno.CreatedAt)
	resp.Key = annotation.Name
	resp.Namespace = cached.Namespace.Name
	resp.Size = int64(anno.Size)
	resp.UpdatedAt = timestamppb.New(anno.UpdatedAt)
	resp.MimeType = anno.MimeType

	return &resp, nil

}
