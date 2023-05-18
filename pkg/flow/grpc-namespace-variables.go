package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (srv *server) getNamespaceVariable(ctx context.Context, cached *database.CacheData, key string, load bool) (*database.VarRef, *database.VarData, error) {
	vref, err := srv.database.NamespaceVariable(ctx, cached.Namespace.ID, key)
	if err != nil {
		return nil, nil, err
	}

	vdata, err := srv.database.VariableData(ctx, vref.VarData, load)
	if err != nil {
		return nil, nil, err
	}

	return vref, vdata, nil
}

func (srv *server) traverseToNamespaceVariable(ctx context.Context, namespace, key string, load bool) (*database.CacheData, *database.VarRef, *database.VarData, error) {
	cached := new(database.CacheData)

	err := srv.database.NamespaceByName(ctx, cached, namespace)
	if err != nil {
		return nil, nil, nil, err
	}

	vref, vdata, err := srv.getNamespaceVariable(ctx, cached, key, load)
	if err != nil {
		return nil, nil, nil, err
	}

	return cached, vref, vdata, nil
}

func (flow *flow) NamespaceVariable(ctx context.Context, req *grpc.NamespaceVariableRequest) (*grpc.NamespaceVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, vref, vdata, err := flow.traverseToNamespaceVariable(ctx, req.GetNamespace(), req.GetKey(), true)
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespaceVariableResponse

	resp.Namespace = cached.Namespace.Name
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

func (internal *internal) NamespaceVariableParcels(req *grpc.VariableInternalRequest, srv grpc.Internal_NamespaceVariableParcelsServer) error {
	internal.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	cached, err := internal.getInstance(ctx, req.GetInstance())
	if err != nil {
		return err
	}

	vref, vdata, err := internal.getNamespaceVariable(ctx, cached, req.GetKey(), true)
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

func (flow *flow) NamespaceVariableParcels(req *grpc.NamespaceVariableRequest, srv grpc.Flow_NamespaceVariableParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	cached, vref, vdata, err := flow.traverseToNamespaceVariable(ctx, req.GetNamespace(), req.GetKey(), true)
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(vdata.Data)

	for {
		resp := new(grpc.NamespaceVariableResponse)

		resp.Namespace = cached.Namespace.Name
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
