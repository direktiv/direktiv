package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entnote "github.com/direktiv/direktiv/pkg/flow/ent/annotation"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var annotationsOrderings = []*orderingInfo{
	{
		db:           entnote.FieldName,
		req:          util.PaginationKeyName,
		defaultOrder: ent.Asc,
	},
}

var annotationsFilters = map[*filteringInfo]func(query *ent.AnnotationQuery, v string) (*ent.AnnotationQuery, error){
	{
		field: util.PaginationKeyName,
		ftype: "CONTAINS",
	}: func(query *ent.AnnotationQuery, v string) (*ent.AnnotationQuery, error) {
		return query.Where(entnote.NameContains(v)), nil
	},
}

func (flow *flow) NamespaceAnnotation(ctx context.Context, req *grpc.NamespaceAnnotationRequest) (*grpc.NamespaceAnnotationResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	nsc := flow.db.Namespace

	d, err := flow.traverseToNamespaceAnnotation(ctx, nsc, req.GetNamespace(), req.GetKey())
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespaceAnnotationResponse

	resp.Namespace = d.ns().Name
	resp.Key = d.annotation.Name
	resp.CreatedAt = timestamppb.New(d.annotation.CreatedAt)
	resp.UpdatedAt = timestamppb.New(d.annotation.UpdatedAt)
	resp.Checksum = d.annotation.Hash
	resp.Size = int64(d.annotation.Size)
	resp.MimeType = d.annotation.MimeType

	if resp.Size > parcelSize {
		return nil, status.Error(codes.ResourceExhausted, "annotation too large to return without using the parcelling API")
	}

	resp.Data = d.annotation.Data

	return &resp, nil

}

func (flow *flow) NamespaceAnnotationParcels(req *grpc.NamespaceAnnotationRequest, srv grpc.Flow_NamespaceAnnotationParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	nsc := flow.db.Namespace

	d, err := flow.traverseToNamespaceAnnotation(ctx, nsc, req.GetNamespace(), req.GetKey())
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(d.annotation.Data)

	for {

		resp := new(grpc.NamespaceAnnotationResponse)

		resp.Namespace = d.ns().Name
		resp.Key = d.annotation.Name
		resp.CreatedAt = timestamppb.New(d.annotation.CreatedAt)
		resp.UpdatedAt = timestamppb.New(d.annotation.UpdatedAt)
		resp.Checksum = d.annotation.Hash
		resp.Size = int64(d.annotation.Size)
		resp.MimeType = d.annotation.MimeType

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

func (flow *flow) NamespaceAnnotations(ctx context.Context, req *grpc.NamespaceAnnotationsRequest) (*grpc.NamespaceAnnotationsResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.getNamespace(ctx, flow.db.Namespace, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	query := ns.QueryAnnotations()

	results, pi, err := paginate[*ent.AnnotationQuery, *ent.Annotation](ctx, req.Pagination, query, annotationsOrderings, annotationsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.NamespaceAnnotationsResponse)
	resp.Namespace = ns.Name
	resp.Annotations = new(grpc.Annotations)
	resp.Annotations.PageInfo = pi

	err = atob(results, &resp.Annotations.Results)
	if err != nil {
		return nil, err
	}

	return resp, nil

}

func (flow *flow) NamespaceAnnotationsStream(req *grpc.NamespaceAnnotationsRequest, srv grpc.Flow_NamespaceAnnotationsStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	ns, err := flow.getNamespace(ctx, flow.db.Namespace, req.GetNamespace())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeNamespaceAnnotations(ns)
	defer flow.cleanup(sub.Close)

resend:

	query := ns.QueryAnnotations()

	results, pi, err := paginate[*ent.AnnotationQuery, *ent.Annotation](ctx, req.Pagination, query, annotationsOrderings, annotationsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.NamespaceAnnotationsResponse)
	resp.Namespace = ns.Name
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

func (flow *flow) SetNamespaceAnnotation(ctx context.Context, req *grpc.SetNamespaceAnnotationRequest) (*grpc.SetNamespaceAnnotationResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	annotationc := tx.Annotation

	ns, err := flow.getNamespace(ctx, nsc, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	var annotation *ent.Annotation

	key := req.GetKey()

	var newAnnotation bool
	annotation, newAnnotation, err = flow.SetAnnotation(ctx, annotationc, ns, key, req.GetMimeType(), req.GetData())
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if newAnnotation {
		flow.logToNamespace(ctx, time.Now(), ns, "Created namespace annotation '%s'.", key)
	} else {
		flow.logToNamespace(ctx, time.Now(), ns, "Updated namespace annotation '%s'.", key)
	}

	flow.pubsub.NotifyNamespaceAnnotations(ns)

	var resp grpc.SetNamespaceAnnotationResponse

	resp.Namespace = ns.Name
	resp.Key = key
	resp.CreatedAt = timestamppb.New(annotation.CreatedAt)
	resp.UpdatedAt = timestamppb.New(annotation.UpdatedAt)
	resp.Checksum = annotation.Hash
	resp.Size = int64(annotation.Size)
	resp.MimeType = annotation.MimeType

	return &resp, nil

}

type annotationQuerier interface {
	QueryAnnotations() *ent.AnnotationQuery
}

func (flow *flow) SetAnnotation(ctx context.Context, annotationc *ent.AnnotationClient, q annotationQuerier, key string, mimetype string, data []byte) (*ent.Annotation, bool, error) {

	hash, err := computeHash(data)
	if err != nil {
		flow.sugar.Error(err)
	}

	if mimetype == "" {
		mimetype = http.DetectContentType(data)
	}

	var annotation *ent.Annotation
	var newAnnotation bool

	annotation, err = q.QueryAnnotations().Where(entnote.NameEQ(key)).Only(ctx)

	if err != nil {

		if !derrors.IsNotFound(err) {
			return nil, false, err
		}

		query := annotationc.Create().SetSize(len(data)).SetHash(hash).SetData(data).SetName(key).SetMimeType(mimetype)

		switch v := q.(type) {
		case *ent.Namespace:
			query = query.SetNamespace(v)
		case *ent.Workflow:
			query = query.SetWorkflow(v)
		case *ent.Instance:
			query = query.SetInstance(v)
		default:
			panic(errors.New("bad querier"))
		}

		annotation, err = query.Save(ctx)
		if err != nil {
			return nil, false, err
		}

		newAnnotation = true

	} else {

		query := annotation.Update().SetSize(len(data)).SetHash(hash).SetData(data).SetMimeType(mimetype)

		annotation, err = query.Save(ctx)
		if err != nil {
			return nil, false, err
		}

	}

	return annotation, newAnnotation, err

}

func (flow *flow) SetNamespaceAnnotationParcels(srv grpc.Flow_SetNamespaceAnnotationParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	req, err := srv.Recv()
	if err != nil {
		return err
	}

	namespace := req.GetNamespace()
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

	nsc := tx.Namespace
	annotationc := tx.Annotation

	ns, err := flow.getNamespace(ctx, nsc, namespace)
	if err != nil {
		return err
	}

	var annotation *ent.Annotation

	var newAnnotation bool
	annotation, newAnnotation, err = flow.SetAnnotation(ctx, annotationc, ns, key, req.GetMimeType(), buf.Bytes())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newAnnotation {
		flow.logToNamespace(ctx, time.Now(), ns, "Created namespace annotation '%s'.", key)
	} else {
		flow.logToNamespace(ctx, time.Now(), ns, "Updated namespace annotation '%s'.", key)
	}

	flow.pubsub.NotifyNamespaceAnnotations(ns)

	var resp grpc.SetNamespaceAnnotationResponse

	resp.Namespace = ns.Name
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

func (flow *flow) DeleteNamespaceAnnotation(ctx context.Context, req *grpc.DeleteNamespaceAnnotationRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace

	d, err := flow.traverseToNamespaceAnnotation(ctx, nsc, req.GetNamespace(), req.GetKey())
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

	flow.logWithTagsToNamespace(ctx, time.Now(), d, "Deleted namespace annotation '%s'.", d.annotation.Name)
	flow.pubsub.NotifyNamespaceAnnotations(d.ns())

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) RenameNamespaceAnnotation(ctx context.Context, req *grpc.RenameNamespaceAnnotationRequest) (*grpc.RenameNamespaceAnnotationResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToNamespaceAnnotation(ctx, nsc, req.GetNamespace(), req.GetOld())
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

	flow.logWithTagsToNamespace(ctx, time.Now(), d, "Renamed namespace annotation from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyNamespaceAnnotations(d.ns())

	var resp grpc.RenameNamespaceAnnotationResponse

	resp.Checksum = d.annotation.Hash
	resp.CreatedAt = timestamppb.New(d.annotation.CreatedAt)
	resp.Key = annotation.Name
	resp.Namespace = d.ns().Name
	resp.Size = int64(d.annotation.Size)
	resp.UpdatedAt = timestamppb.New(d.annotation.UpdatedAt)
	resp.MimeType = d.annotation.MimeType

	return &resp, nil

}
