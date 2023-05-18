package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (srv *server) getWorkflowVariable(ctx context.Context, cached *database.CacheData, key string, load bool) (*database.VarRef, *database.VarData, error) {
	vref, err := srv.database.WorkflowVariable(ctx, cached.File.ID, key)
	if err != nil {
		return nil, nil, err
	}

	vdata, err := srv.database.VariableData(ctx, vref.VarData, load)
	if err != nil {
		return nil, nil, err
	}

	return vref, vdata, nil
}

func (flow *flow) getWorkflow(ctx context.Context, namespace, path string) (ns *database.Namespace, f *filestore.File, err error) {
	ns, err = flow.edb.NamespaceByName(ctx, namespace)
	if err != nil {
		return
	}

	fStore, _, _, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return
	}
	defer rollback()

	f, err = fStore.ForRootID(ns.ID).GetFile(ctx, path)
	if err != nil {
		return
	}

	if f.Typ != filestore.FileTypeWorkflow {
		err = ErrNotWorkflow
		return
	}

	return
}

func (flow *flow) getWorkflowVariable(ctx context.Context, namespace, path, key string, loadData bool) (ns *database.Namespace, f *filestore.File, vref *database.VarRef, vdata *database.VarData, err error) {
	ns, f, err = flow.getWorkflow(ctx, namespace, path)
	if err != nil {
		return
	}

	vref, err = flow.database.WorkflowVariable(ctx, f.ID, key)
	if err != nil {
		return
	}

	vdata, err = flow.database.VariableData(ctx, vref.VarData, loadData)
	if err != nil {
		return
	}

	return
}

func (flow *flow) WorkflowVariable(ctx context.Context, req *grpc.WorkflowVariableRequest) (*grpc.WorkflowVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, f, vref, vdata, err := flow.getWorkflowVariable(ctx, req.GetNamespace(), req.GetPath(), req.GetKey(), true)
	if err != nil {
		return nil, err
	}

	var resp grpc.WorkflowVariableResponse

	resp.Namespace = ns.Name
	resp.Path = f.Path
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

func (internal *internal) WorkflowVariableParcels(req *grpc.VariableInternalRequest, srv grpc.Internal_WorkflowVariableParcelsServer) error {
	internal.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	cached, err := internal.getInstance(ctx, req.GetInstance())
	if err != nil {
		return err
	}

	vref, vdata, err := internal.getWorkflowVariable(ctx, cached, req.GetKey(), true)
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

func (flow *flow) WorkflowVariableParcels(req *grpc.WorkflowVariableRequest, srv grpc.Flow_WorkflowVariableParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	ns, f, vref, vdata, err := flow.getWorkflowVariable(ctx, req.GetNamespace(), req.GetPath(), req.GetKey(), true)
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(vdata.Data)

	for {
		resp := new(grpc.WorkflowVariableResponse)

		resp.Namespace = ns.Name
		resp.Path = f.Path
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

type varQuerier interface {
}

type entNamespaceVarQuerier struct {
	clients *entwrapper.EntClients
	cached  *database.CacheData
}

type entWorkflowVarQuerier struct {
	clients *entwrapper.EntClients
	// cached  *database.CacheData
	ns *database.Namespace
	f  *filestore.File
}

type entInstanceVarQuerier struct {
	clients *entwrapper.EntClients
	cached  *database.CacheData
}
