package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entnote "github.com/direktiv/direktiv/pkg/flow/ent/annotation"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
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

func (flow *flow) traverseToNamespaceAnnotation(ctx context.Context, namespace, key string) (*database.CacheData, *database.Annotation, error) {
	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, namespace)
	if err != nil {
		return nil, nil, err
	}

	annotation, err := flow.database.NamespaceAnnotation(ctx, cached.Namespace.ID, key)
	if err != nil {
		return nil, nil, err
	}

	return cached, annotation, nil
}

func (flow *flow) NamespaceAnnotation(ctx context.Context, req *grpc.NamespaceAnnotationRequest) (*grpc.NamespaceAnnotationResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, annotation, err := flow.traverseToNamespaceAnnotation(ctx, req.GetNamespace(), req.GetKey())
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespaceAnnotationResponse

	resp.Namespace = cached.Namespace.Name
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

func (flow *flow) NamespaceAnnotationParcels(req *grpc.NamespaceAnnotationRequest, srv grpc.Flow_NamespaceAnnotationParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	cached, annotation, err := flow.traverseToNamespaceAnnotation(ctx, req.GetNamespace(), req.GetKey())
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(annotation.Data)

	for {
		resp := new(grpc.NamespaceAnnotationResponse)

		resp.Namespace = cached.Namespace.Name
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

func (flow *flow) NamespaceAnnotations(ctx context.Context, req *grpc.NamespaceAnnotationsRequest) (*grpc.NamespaceAnnotationsResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	query := clients.Annotation.Query().Where(entnote.HasNamespaceWith(entns.ID(cached.Namespace.ID)))

	results, pi, err := paginate[*ent.AnnotationQuery, *ent.Annotation](ctx, req.Pagination, query, annotationsOrderings, annotationsFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.NamespaceAnnotationsResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Annotations = new(grpc.Annotations)
	resp.Annotations.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Annotations.Results)
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

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeNamespaceAnnotations(cached.Namespace)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(ctx)

	query := clients.Annotation.Query().Where(entnote.HasNamespaceWith(entns.ID(cached.Namespace.ID)))

	results, pi, err := paginate[*ent.AnnotationQuery, *ent.Annotation](ctx, req.Pagination, query, annotationsOrderings, annotationsFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.NamespaceAnnotationsResponse)
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

func (flow *flow) SetNamespaceAnnotation(ctx context.Context, req *grpc.SetNamespaceAnnotationRequest) (*grpc.SetNamespaceAnnotationResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached := new(database.CacheData)

	err = flow.database.NamespaceByName(tctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	var annotation *ent.Annotation

	key := req.GetKey()

	var newAnnotation bool
	annotation, newAnnotation, err = flow.SetAnnotation(tctx, &entNamespaceAnnotationQuerier{cached: cached, clients: flow.edb.Clients(tctx)}, key, req.GetMimeType(), req.GetData())
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if newAnnotation {
		flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("namespace")), logengine.Info, "Created namespace annotation '%s'.", key)
	} else {
		flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("namespace")), logengine.Info, "Updated namespace annotation '%s'.", key)
	}

	flow.pubsub.NotifyNamespaceAnnotations(cached.Namespace)

	var resp grpc.SetNamespaceAnnotationResponse

	resp.Namespace = cached.Namespace.Name
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

type entNamespaceAnnotationQuerier struct {
	clients *entwrapper.EntClients
	cached  *database.CacheData
}

func (x *entNamespaceAnnotationQuerier) QueryAnnotations() *ent.AnnotationQuery {
	return x.clients.Annotation.Query().Where(entnote.HasNamespaceWith(entns.ID(x.cached.Namespace.ID)))
}

type entInstanceAnnotationQuerier struct {
	clients *entwrapper.EntClients
	cached  *database.CacheData
}

func (x *entInstanceAnnotationQuerier) QueryAnnotations() *ent.AnnotationQuery {
	return x.clients.Annotation.Query().Where(entnote.HasInstanceWith(entinst.ID(x.cached.Instance.ID)))
}

func (flow *flow) SetAnnotation(ctx context.Context, q annotationQuerier, key string, mimetype string, data []byte) (*ent.Annotation, bool, error) {
	hash, err := bytedata.ComputeHash(data)
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
		clients := flow.edb.Clients(ctx)

		query := clients.Annotation.Create().SetSize(len(data)).SetHash(hash).SetData(data).SetName(key).SetMimeType(mimetype)

		switch v := q.(type) {
		case *entNamespaceAnnotationQuerier:
			query = query.SetNamespace(v.clients.Namespace.GetX(ctx, v.cached.Namespace.ID))
		case *entInstanceAnnotationQuerier:
			query = query.SetInstance(v.clients.Instance.GetX(ctx, v.cached.Instance.ID))
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

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	cached := new(database.CacheData)

	err = flow.database.NamespaceByName(tctx, cached, namespace)
	if err != nil {
		return err
	}

	var annotation *ent.Annotation

	var newAnnotation bool
	annotation, newAnnotation, err = flow.SetAnnotation(tctx, &entNamespaceAnnotationQuerier{cached: cached, clients: flow.edb.Clients(tctx)}, key, req.GetMimeType(), buf.Bytes())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newAnnotation {
		flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("namespace")), logengine.Info, "Created namespace annotation '%s'.", key)
	} else {
		flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("namespace")), logengine.Info, "Updated namespace annotation '%s'.", key)
	}

	flow.pubsub.NotifyNamespaceAnnotations(cached.Namespace)

	var resp grpc.SetNamespaceAnnotationResponse

	resp.Namespace = cached.Namespace.Name
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

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, annotation, err := flow.traverseToNamespaceAnnotation(tctx, req.GetNamespace(), req.GetKey())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tctx)

	err = clients.Annotation.DeleteOneID(annotation.ID).Exec(tctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("namespace")), logengine.Info, "Deleted namespace annotation '%s'.", annotation.Name)
	flow.pubsub.NotifyNamespaceAnnotations(cached.Namespace)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) RenameNamespaceAnnotation(ctx context.Context, req *grpc.RenameNamespaceAnnotationRequest) (*grpc.RenameNamespaceAnnotationResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, annotation, err := flow.traverseToNamespaceAnnotation(tctx, req.GetNamespace(), req.GetOld())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tctx)

	x, err := clients.Annotation.UpdateOneID(annotation.ID).SetName(req.GetNew()).Save(tctx)
	if err != nil {
		flow.loggerBeta.Log(addTraceFrom(ctx, flow.GetAttributes()), logengine.Error, "Failed to rename a namespace annotation '%s'.", req.GetOld())
		return nil, err
	}

	annotation.Name = x.Name

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("namespace")), logengine.Info, "Renamed namespace annotation from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyNamespaceAnnotations(cached.Namespace)

	var resp grpc.RenameNamespaceAnnotationResponse

	resp.Checksum = annotation.Hash
	resp.CreatedAt = timestamppb.New(annotation.CreatedAt)
	resp.Key = annotation.Name
	resp.Namespace = cached.Namespace.Name
	resp.Size = int64(annotation.Size)
	resp.UpdatedAt = timestamppb.New(annotation.UpdatedAt)
	resp.MimeType = annotation.MimeType

	return &resp, nil
}
