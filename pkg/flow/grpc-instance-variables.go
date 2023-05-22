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
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entvardata "github.com/direktiv/direktiv/pkg/flow/ent/vardata"
	entvar "github.com/direktiv/direktiv/pkg/flow/ent/varref"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

		srv.loggerBeta.Log(addTraceFrom(ctx, srv.flow.GetAttributes()), logengine.Error, "Failed to resolve instance %s", instance)
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
		srv.loggerBeta.Log(addTraceFrom(ctx, srv.flow.GetAttributes()), logengine.Error, "Failed to resolve variable")
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

func (flow *flow) InstanceVariables(ctx context.Context, req *grpc.InstanceVariablesRequest) (*grpc.InstanceVariablesResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	query := clients.VarRef.Query().Where(entvar.HasInstanceWith(entinst.ID(cached.Instance.ID)))

	results, pi, err := paginate[*ent.VarRefQuery, *ent.VarRef](ctx, req.Pagination, query, variablesOrderings, variablesFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.InstanceVariablesResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Instance = cached.Instance.ID.String()
	resp.Variables = new(grpc.Variables)
	resp.Variables.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Variables.Results)
	if err != nil {
		return nil, err
	}

	for i := range results {
		vref := results[i]

		vdata, err := vref.QueryVardata().Select(entvardata.FieldCreatedAt, entvardata.FieldHash, entvardata.FieldSize, entvardata.FieldUpdatedAt).Only(ctx)
		if err != nil {
			return nil, err
		}

		v := resp.Variables.Results[i]
		v.Checksum = vdata.Hash
		v.CreatedAt = timestamppb.New(vdata.CreatedAt)
		v.Size = int64(vdata.Size)
		v.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
		v.MimeType = vdata.MimeType
	}

	return resp, nil
}

func (flow *flow) InstanceVariablesStream(req *grpc.InstanceVariablesRequest, srv grpc.Flow_InstanceVariablesStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	cached, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeInstanceVariables(cached)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(ctx)

	query := clients.VarRef.Query().Where(entvar.HasInstanceWith(entinst.ID(cached.Instance.ID)))

	results, pi, err := paginate[*ent.VarRefQuery, *ent.VarRef](ctx, req.Pagination, query, variablesOrderings, variablesFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.InstanceVariablesResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Instance = cached.Instance.ID.String()
	resp.Variables = new(grpc.Variables)
	resp.Variables.PageInfo = pi

	err = bytedata.ConvertDataForOutput(results, &resp.Variables.Results)
	if err != nil {
		return err
	}

	for i := range results {
		vref := results[i]

		vdata, err := vref.QueryVardata().Select(entvardata.FieldCreatedAt, entvardata.FieldHash, entvardata.FieldSize, entvardata.FieldUpdatedAt).Only(ctx)
		if err != nil {
			return err
		}

		v := resp.Variables.Results[i]
		v.Checksum = vdata.Hash
		v.CreatedAt = timestamppb.New(vdata.CreatedAt)
		v.Size = int64(vdata.Size)
		v.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
		v.MimeType = vdata.MimeType
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

func (flow *flow) SetInstanceVariable(ctx context.Context, req *grpc.SetInstanceVariableRequest) (*grpc.SetInstanceVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	key := req.GetKey()

	cached, err := flow.getInstance(tctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		flow.loggerBeta.Log(addTraceFrom(ctx, flow.GetAttributes()), logengine.Error, "Failed to resolve instance '%s'.", req.GetInstance())
		return nil, err
	}

	var vdata *ent.VarData
	var newVar bool

	vdata, newVar, err = flow.SetVariable(tctx, &entInstanceVarQuerier{clients: flow.edb.Clients(tctx), cached: cached}, key, req.GetData(), req.GetMimeType(), false)
	if err != nil {
		flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Error, "Could not create / change instance variable.")
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if newVar {
		flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Info, "Created instance variable '%s'.", key)
	} else {
		flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Info, "Updated instance variable '%s'.", key)
	}
	flow.pubsub.NotifyInstanceVariables(cached.Instance)

	var resp grpc.SetInstanceVariableResponse

	resp.Namespace = cached.Namespace.Name
	resp.Instance = cached.Instance.ID.String()
	resp.Key = key
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.Checksum = vdata.Hash
	resp.TotalSize = int64(vdata.Size)
	resp.MimeType = vdata.MimeType

	return &resp, nil
}

func (internal *internal) SetThreadVariableParcels(srv grpc.Internal_SetThreadVariableParcelsServer) error {
	internal.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	req, err := srv.Recv()
	if err != nil {
		return err
	}

	mimeType := req.GetMimeType()
	instance := req.GetInstance()
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

	tctx, tx, err := internal.database.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	cached, err := internal.getInstance(tctx, instance)
	if err != nil {
		return err
	}

	var vdata *ent.VarData
	var newVar bool

	vdata, newVar, err = internal.flow.SetVariable(tctx, &entInstanceVarQuerier{clients: internal.edb.Clients(tctx), cached: cached}, key, buf.Bytes(), mimeType, true)
	if err != nil {
		internal.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Error, "Could not create / change thread variable")
		return err
	}

	err = tx.Commit()
	if err != nil {
		internal.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Error, "Could not create / change thread variable '%s'.", key)
		return err
	}

	if newVar {
		internal.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Error, "Created thread variable '%s'.", key)
	} else {
		internal.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Error, "Updated thread variable '%s'.", key)
	}

	internal.pubsub.NotifyInstanceVariables(cached.Instance) // what do we do about this for thread variables?

	var resp grpc.SetVariableInternalResponse

	resp.Instance = cached.Instance.ID.String()
	resp.Key = key
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.Checksum = vdata.Hash
	resp.TotalSize = int64(vdata.Size)
	resp.MimeType = vdata.MimeType

	err = srv.SendAndClose(&resp)
	if err != nil {
		return err
	}

	return nil
}

func (internal *internal) SetInstanceVariableParcels(srv grpc.Internal_SetInstanceVariableParcelsServer) error {
	internal.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	req, err := srv.Recv()
	if err != nil {
		return err
	}

	mimeType := req.GetMimeType()
	instance := req.GetInstance()
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

	tctx, tx, err := internal.database.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	cached, err := internal.getInstance(tctx, instance)
	if err != nil {
		internal.loggerBeta.Log(addTraceFrom(ctx, internal.flow.GetAttributes()), logengine.Error, "Failed to resolve instance %s", req.GetInstance())
		return err
	}

	var vdata *ent.VarData
	var newVar bool

	vdata, newVar, err = internal.flow.SetVariable(tctx, &entInstanceVarQuerier{clients: internal.edb.Clients(tctx), cached: cached}, key, buf.Bytes(), mimeType, false)
	if err != nil {
		internal.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Error, "Could not create or update instance variable '%s'.", key)
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newVar {
		internal.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Info, "Created instance variable '%s'.", key)
	} else {
		internal.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Info, "Updated instance variable '%s'.", key)
	}

	internal.pubsub.NotifyInstanceVariables(cached.Instance)

	var resp grpc.SetVariableInternalResponse

	resp.Instance = cached.Instance.ID.String()
	resp.Key = key
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.Checksum = vdata.Hash
	resp.TotalSize = int64(vdata.Size)
	resp.MimeType = vdata.MimeType

	err = srv.SendAndClose(&resp)
	if err != nil {
		return err
	}

	return nil
}

func (flow *flow) SetInstanceVariableParcels(srv grpc.Flow_SetInstanceVariableParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	req, err := srv.Recv()
	if err != nil {
		return err
	}

	mimeType := req.GetMimeType()
	namespace := req.GetNamespace()
	instance := req.GetInstance()
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

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	cached, err := flow.getInstance(tctx, namespace, instance)
	if err != nil {
		flow.loggerBeta.Log(addTraceFrom(ctx, flow.GetAttributes()), logengine.Error, "Failed to resolve instance %s", req.GetInstance())
		return err
	}

	var vdata *ent.VarData
	var newVar bool

	vdata, newVar, err = flow.SetVariable(tctx, &entInstanceVarQuerier{clients: flow.edb.Clients(tctx), cached: cached}, key, req.GetData(), mimeType, false)
	if err != nil {
		flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Error, "Could not create / change instance variable '%s'.", key)
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newVar {
		flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Info, "Created instance variable '%s'.", key)
	} else {
		flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Info, "Updated instance variable '%s'.", key)
	}

	flow.pubsub.NotifyInstanceVariables(cached.Instance)

	var resp grpc.SetInstanceVariableResponse

	resp.Namespace = cached.Namespace.Name
	resp.Instance = cached.Instance.ID.String()
	resp.Key = key
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.Checksum = vdata.Hash
	resp.TotalSize = int64(vdata.Size)
	resp.MimeType = vdata.MimeType

	err = srv.SendAndClose(&resp)
	if err != nil {
		return err
	}

	return nil
}

func (flow *flow) DeleteInstanceVariable(ctx context.Context, req *grpc.DeleteInstanceVariableRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, vref, vdata, err := flow.traverseToInstanceVariable(tctx, req.GetNamespace(), req.GetInstance(), req.GetKey(), false)
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tctx)

	err = clients.VarRef.DeleteOneID(vref.ID).Exec(ctx)
	if err != nil {
		flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Error, "Failed to delete instance variable ID '%s'.", vref.Name)
		return nil, err
	}

	if vdata.RefCount == 0 {
		err = clients.VarData.DeleteOneID(vdata.ID).Exec(ctx)
		if err != nil {
			flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Error, "Failed to delete instance variable data '%s'.", vref.Name)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Info, "Deleted instance variable '%s'.", vref.Name)
	flow.pubsub.NotifyInstanceVariables(cached.Instance)

	// Broadcast Event
	broadcastInput := broadcastVariableInput{
		Key:        req.GetKey(),
		InstanceID: req.GetInstance(),
		TotalSize:  int64(vdata.Size),
		Scope:      BroadcastEventScopeInstance,
	}
	err = flow.BroadcastVariable(ctx, BroadcastEventTypeDelete, BroadcastEventScopeInstance, broadcastInput, cached.Namespace)
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) RenameInstanceVariable(ctx context.Context, req *grpc.RenameInstanceVariableRequest) (*grpc.RenameInstanceVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, vref, vdata, err := flow.traverseToInstanceVariable(tctx, req.GetNamespace(), req.GetInstance(), req.GetOld(), false)
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tctx)

	x, err := clients.VarRef.UpdateOneID(vref.ID).SetName(req.GetNew()).Save(tctx)
	if err != nil {
		flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Error, "Failed to store new instance variable name")
		return nil, err
	}

	vref.Name = x.Name

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	flow.loggerBeta.Log(addTraceFrom(ctx, cached.GetAttributes("instance")), logengine.Info, "Renamed instance variable from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyInstanceVariables(cached.Instance)

	var resp grpc.RenameInstanceVariableResponse

	resp.Checksum = vdata.Hash
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.Key = vref.Name
	resp.Namespace = cached.Namespace.Name
	resp.TotalSize = int64(vdata.Size)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.MimeType = vdata.MimeType

	return &resp, nil
}
