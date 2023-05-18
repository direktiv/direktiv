package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (srv *server) getInstanceVariable(ctx context.Context, cached *database.CacheData, key string, load bool) (*database.VarRef, *database.VarData, error) {
	vref, err := srv.database.InstanceVariable(ctx, cached.Instance.ID, key)
	if err != nil {
		return nil, nil, err
	}

	vdata, err := srv.database.VariableData(ctx, vref.VarData, load)
	if err != nil {
		return nil, nil, err
	}

	return vref, vdata, nil
}

func (srv *server) getThreadVariable(ctx context.Context, cached *database.CacheData, key string, load bool) (*database.VarRef, *database.VarData, error) {
	vref, err := srv.database.ThreadVariable(ctx, cached.Instance.ID, key)
	if err != nil {
		return nil, nil, err
	}

	vdata, err := srv.database.VariableData(ctx, vref.VarData, load)
	if err != nil {
		return nil, nil, err
	}

	return vref, vdata, nil
}

func (srv *server) traverseToInstanceVariable(ctx context.Context, namespace, instance, key string, load bool) (*database.CacheData, *database.VarRef, *database.VarData, error) {
	id, err := uuid.Parse(instance)
	if err != nil {
		return nil, nil, nil, err
	}

	cached := new(database.CacheData)

	err = srv.database.Instance(ctx, cached, id)
	if err != nil {
		srv.logger.Errorf(ctx, srv.ID, srv.flow.GetAttributes(), "Failed to resolve instance %s", instance)
		return nil, nil, nil, err
	}

	fStore, _, _, rollback, err := srv.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rollback()

	file, revision, err := fStore.GetRevision(ctx, cached.Instance.Revision)
	if err != nil {
		return nil, nil, nil, err
	}

	cached.File = file
	cached.Revision = revision

	if cached.Namespace.Name != namespace {
		return nil, nil, nil, os.ErrNotExist
	}

	vref, vdata, err := srv.getInstanceVariable(ctx, cached, key, load)
	if err != nil {
		srv.logger.Errorf(ctx, srv.ID, srv.flow.GetAttributes(), "Failed to resolve variable")
		return nil, nil, nil, err
	}

	return cached, vref, vdata, nil
}

func (flow *flow) InstanceVariable(ctx context.Context, req *grpc.InstanceVariableRequest) (*grpc.InstanceVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, vref, vdata, err := flow.traverseToInstanceVariable(ctx, req.GetNamespace(), req.GetInstance(), req.GetKey(), true)
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceVariableResponse

	resp.Namespace = cached.Namespace.Name
	resp.Instance = cached.Instance.ID.String()
	resp.Key = vref.Name
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.Checksum = vdata.Hash
	resp.TotalSize = int64(vdata.Size)
	resp.MimeType = vdata.MimeType

	if resp.TotalSize > parcelSize {
		return nil, status.Error(codes.ResourceExhausted, "variable too large to return without using the parcelling API")
	}

	resp.Data = vdata.Data

	return &resp, nil
}

func (internal *internal) InstanceVariableParcels(req *grpc.VariableInternalRequest, srv grpc.Internal_InstanceVariableParcelsServer) error {
	internal.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	instID, err := uuid.Parse(req.GetInstance())
	if err != nil {
		return err
	}

	cached := new(database.CacheData)

	err = internal.database.Instance(ctx, cached, instID)
	if err != nil {
		return err
	}

	fStore, _, _, rollback, err := internal.flow.beginSqlTx(ctx)
	if err != nil {
		return err
	}
	defer rollback()

	file, revision, err := fStore.GetRevision(ctx, cached.Instance.Revision)
	if err != nil {
		return err
	}

	cached.File = file
	cached.Revision = revision

	vref, vdata, err := internal.getInstanceVariable(ctx, cached, req.GetKey(), true)
	if err != nil && !derrors.IsNotFound(err) {
		return err
	}

	if derrors.IsNotFound(err) {
		vref = new(database.VarRef)
		vref.Name = req.GetKey()
		vdata = new(database.VarData)
		t := time.Now()
		vdata.Data = make([]byte, 0)
		hash, err := bytedata.ComputeHash(vdata.Data)
		if err != nil {
			internal.sugar.Error(err)
		}
		vdata.CreatedAt = t
		vdata.UpdatedAt = t
		vdata.Hash = hash
		vdata.Size = 0
	}

	rdr := bytes.NewReader(vdata.Data)

	for {
		resp := new(grpc.VariableInternalResponse)

		resp.Key = vref.Name
		resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
		resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
		resp.Checksum = vdata.Hash
		resp.TotalSize = int64(vdata.Size)
		resp.MimeType = vdata.MimeType

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

func (internal *internal) ThreadVariableParcels(req *grpc.VariableInternalRequest, srv grpc.Internal_ThreadVariableParcelsServer) error {
	internal.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	instID, err := uuid.Parse(req.GetInstance())
	if err != nil {
		return err
	}

	cached := new(database.CacheData)

	err = internal.database.Instance(ctx, cached, instID)
	if err != nil {
		return err
	}

	fStore, _, _, rollback, err := internal.flow.beginSqlTx(ctx)
	if err != nil {
		return err
	}
	defer rollback()

	file, revision, err := fStore.GetRevision(ctx, cached.Instance.Revision)
	if err != nil {
		return err
	}

	cached.File = file
	cached.Revision = revision

	vref, vdata, err := internal.getThreadVariable(ctx, cached, req.GetKey(), true)
	if err != nil && !derrors.IsNotFound(err) {
		return err
	}

	if derrors.IsNotFound(err) {
		vref = new(database.VarRef)
		vref.Name = req.GetKey()
		vdata = new(database.VarData)
		t := time.Now()
		vdata.Data = make([]byte, 0)
		hash, err := bytedata.ComputeHash(vdata.Data)
		if err != nil {
			internal.sugar.Error(err)
		}
		vdata.CreatedAt = t
		vdata.UpdatedAt = t
		vdata.Hash = hash
		vdata.Size = 0
	}

	rdr := bytes.NewReader(vdata.Data)

	for {
		resp := new(grpc.VariableInternalResponse)

		resp.Key = vref.Name
		resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
		resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
		resp.Checksum = vdata.Hash
		resp.TotalSize = int64(vdata.Size)
		resp.MimeType = vdata.MimeType

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

func (flow *flow) InstanceVariableParcels(req *grpc.InstanceVariableRequest, srv grpc.Flow_InstanceVariableParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	cached, vref, vdata, err := flow.traverseToInstanceVariable(ctx, req.GetNamespace(), req.GetInstance(), req.GetKey(), true)
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(vdata.Data)

	for {
		resp := new(grpc.InstanceVariableResponse)

		resp.Namespace = cached.Namespace.Name
		resp.Instance = cached.Instance.ID.String()
		resp.Key = vref.Name
		resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
		resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
		resp.Checksum = vdata.Hash
		resp.TotalSize = int64(vdata.Size)
		resp.MimeType = vdata.MimeType

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
