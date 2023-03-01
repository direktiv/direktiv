package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entnote "github.com/direktiv/direktiv/pkg/flow/ent/annotation"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (flow *flow) traverseToInstanceAnnotation(ctx context.Context, namespace, instance, key string) (*database.CacheData, *database.Annotation, error) {
	id, err := uuid.Parse(instance)
	if err != nil {
		return nil, nil, err
	}

	cached := new(database.CacheData)

	err = flow.database.Instance(ctx, cached, id)
	if err != nil {
		return nil, nil, err
	}

	if cached.Namespace.Name != namespace {
		return nil, nil, os.ErrNotExist
	}

	annotation, err := flow.database.InstanceAnnotation(ctx, cached.Instance.ID, key)
	if err != nil {
		return nil, nil, err
	}

	return cached, annotation, nil
}

func (flow *flow) InstanceAnnotation(ctx context.Context, req *grpc.InstanceAnnotationRequest) (*grpc.InstanceAnnotationResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, annotation, err := flow.traverseToInstanceAnnotation(ctx, req.GetNamespace(), req.GetInstance(), req.GetKey())
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceAnnotationResponse

	resp.Namespace = cached.Namespace.Name
	resp.Instance = cached.Instance.ID.String()
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

func (flow *flow) InstanceAnnotationParcels(req *grpc.InstanceAnnotationRequest, srv grpc.Flow_InstanceAnnotationParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	cached, annotation, err := flow.traverseToInstanceAnnotation(ctx, req.GetNamespace(), req.GetInstance(), req.GetKey())
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(annotation.Data)

	for {

		resp := new(grpc.InstanceAnnotationResponse)

		resp.Namespace = cached.Namespace.Name
		resp.Instance = cached.Instance.ID.String()
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

func (flow *flow) InstanceAnnotations(ctx context.Context, req *grpc.InstanceAnnotationsRequest) (*grpc.InstanceAnnotationsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	query := clients.Annotation.Query().Where(entnote.HasInstanceWith(entinst.ID(cached.Instance.ID)))

	results, pi, err := paginate[*ent.AnnotationQuery, *ent.Annotation](ctx, req.Pagination, query, annotationsOrderings, annotationsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.InstanceAnnotationsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Annotations = new(grpc.Annotations)
	resp.Annotations.PageInfo = pi

	err = atob(results, &resp.Annotations.Results)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (flow *flow) InstanceAnnotationsStream(req *grpc.InstanceAnnotationsRequest, srv grpc.Flow_InstanceAnnotationsStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	cached, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeInstanceAnnotations(cached)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(ctx)

	query := clients.Annotation.Query().Where(entnote.HasInstanceWith(entinst.ID(cached.Instance.ID)))

	results, pi, err := paginate[*ent.AnnotationQuery, *ent.Annotation](ctx, req.Pagination, query, annotationsOrderings, annotationsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.InstanceAnnotationsResponse)
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

func (flow *flow) SetInstanceAnnotation(ctx context.Context, req *grpc.SetInstanceAnnotationRequest) (*grpc.SetInstanceAnnotationResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	key := req.GetKey()

	cached, err := flow.getInstance(tctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	var annotation *ent.Annotation
	var newVar bool

	annotation, newVar, err = flow.SetAnnotation(tctx, &entInstanceAnnotationQuerier{clients: flow.edb.Clients(tctx), cached: cached}, key, req.GetMimeType(), req.GetData())
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if newVar {
		flow.logToInstance(ctx, time.Now(), cached, "Created instance annotation '%s'.", key)
	} else {
		flow.logToInstance(ctx, time.Now(), cached, "Updated instance annotation '%s'.", key)
	}
	flow.pubsub.NotifyInstanceAnnotations(cached.Instance)

	var resp grpc.SetInstanceAnnotationResponse

	resp.Namespace = cached.Namespace.Name
	resp.Instance = cached.Instance.ID.String()
	resp.Key = key
	resp.CreatedAt = timestamppb.New(annotation.CreatedAt)
	resp.UpdatedAt = timestamppb.New(annotation.UpdatedAt)
	resp.Checksum = annotation.Hash
	resp.Size = int64(annotation.Size)
	resp.MimeType = annotation.MimeType

	return &resp, nil
}

func (flow *flow) SetInstanceAnnotationParcels(srv grpc.Flow_SetInstanceAnnotationParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	req, err := srv.Recv()
	if err != nil {
		return err
	}

	namespace := req.GetNamespace()
	instance := req.GetInstance()
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

	cached, err := flow.getInstance(tctx, namespace, instance)
	if err != nil {
		return err
	}

	var annotation *ent.Annotation
	var newVar bool

	annotation, newVar, err = flow.SetAnnotation(tctx, &entInstanceAnnotationQuerier{clients: flow.edb.Clients(tctx), cached: cached}, key, req.GetMimeType(), req.GetData())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newVar {
		flow.logToInstance(ctx, time.Now(), cached, "Created instance annotation '%s'.", key)
	} else {
		flow.logToInstance(ctx, time.Now(), cached, "Updated instance annotation '%s'.", key)
	}

	flow.pubsub.NotifyInstanceAnnotations(cached.Instance)

	var resp grpc.SetInstanceAnnotationResponse

	resp.Namespace = cached.Namespace.Name
	resp.Instance = cached.Instance.ID.String()
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

func (flow *flow) DeleteInstanceAnnotation(ctx context.Context, req *grpc.DeleteInstanceAnnotationRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, annotation, err := flow.traverseToInstanceAnnotation(tctx, req.GetNamespace(), req.GetInstance(), req.GetKey())
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

	flow.logToInstance(ctx, time.Now(), cached, "Deleted instance annotation '%s'.", annotation.Name)
	flow.pubsub.NotifyInstanceAnnotations(cached.Instance)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) RenameInstanceAnnotation(ctx context.Context, req *grpc.RenameInstanceAnnotationRequest) (*grpc.RenameInstanceAnnotationResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, annotation, err := flow.traverseToInstanceAnnotation(tctx, req.GetNamespace(), req.GetInstance(), req.GetOld())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tctx)

	anno, err := clients.Annotation.UpdateOneID(annotation.ID).SetName(req.GetNew()).Save(tctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToInstance(ctx, time.Now(), cached, "Renamed instance annotation from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyInstanceAnnotations(cached.Instance)

	var resp grpc.RenameInstanceAnnotationResponse

	resp.Checksum = anno.Hash
	resp.CreatedAt = timestamppb.New(anno.CreatedAt)
	resp.Key = anno.Name
	resp.Namespace = cached.Namespace.Name
	resp.Size = int64(anno.Size)
	resp.UpdatedAt = timestamppb.New(anno.UpdatedAt)
	resp.MimeType = anno.MimeType

	return &resp, nil
}
