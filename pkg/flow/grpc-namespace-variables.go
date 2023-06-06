package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	entvardata "github.com/direktiv/direktiv/pkg/flow/ent/vardata"
	entvar "github.com/direktiv/direktiv/pkg/flow/ent/varref"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (srv *server) getNamespaceVariable(ctx context.Context, nsID uuid.UUID, key string, load bool) (*database.VarRef, *database.VarData, error) {
	vref, err := srv.database.NamespaceVariable(ctx, nsID, key)
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
	ns, err := srv.edb.NamespaceByName(ctx, namespace)
	if err != nil {
		return nil, nil, nil, err
	}

	vref, vdata, err := srv.getNamespaceVariable(ctx, ns.ID, key, load)
	if err != nil {
		return nil, nil, nil, err
	}

	cached := &database.CacheData{
		Namespace: ns,
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

	instance, err := internal.getInstance(ctx, req.GetInstance())
	if err != nil {
		return err
	}

	vref, vdata, err := internal.getNamespaceVariable(ctx, instance.Instance.NamespaceID, req.GetKey(), true)
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

var variablesOrderings = []*orderingInfo{
	{
		db:           entvar.FieldName,
		req:          "UPDATED",
		defaultOrder: ent.Asc,
	},
}

var variablesFilters = map[*filteringInfo]func(query *ent.VarRefQuery, v string) (*ent.VarRefQuery, error){
	{
		field: util.PaginationKeyName,
		ftype: "CONTAINS",
	}: func(query *ent.VarRefQuery, v string) (*ent.VarRefQuery, error) {
		return query.Where(entvar.NameContains(v)), nil
	},
}

func (flow *flow) NamespaceVariables(ctx context.Context, req *grpc.NamespaceVariablesRequest) (*grpc.NamespaceVariablesResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	query := clients.VarRef.Query().Where(entvar.HasNamespaceWith(entns.ID(cached.Namespace.ID)))

	results, pi, err := paginate[*ent.VarRefQuery, *ent.VarRef](ctx, req.Pagination, query, variablesOrderings, variablesFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.NamespaceVariablesResponse)
	resp.Namespace = cached.Namespace.Name
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

func (flow *flow) NamespaceVariablesStream(req *grpc.NamespaceVariablesRequest, srv grpc.Flow_NamespaceVariablesStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	cached := new(database.CacheData)

	err := flow.database.NamespaceByName(ctx, cached, req.GetNamespace())
	if err != nil {
		return err
	}

	clients := flow.edb.Clients(ctx)

	sub := flow.pubsub.SubscribeNamespaceVariables(cached.Namespace)
	defer flow.cleanup(sub.Close)

resend:

	query := clients.VarRef.Query().Where(entvar.HasNamespaceWith(entns.ID(cached.Namespace.ID)))

	results, pi, err := paginate[*ent.VarRefQuery, *ent.VarRef](ctx, req.Pagination, query, variablesOrderings, variablesFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.NamespaceVariablesResponse)
	resp.Namespace = cached.Namespace.Name
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

func (flow *flow) SetNamespaceVariable(ctx context.Context, req *grpc.SetNamespaceVariableRequest) (*grpc.SetNamespaceVariableResponse, error) {
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

	var vdata *ent.VarData

	key := req.GetKey()

	var newVar bool
	vdata, newVar, err = flow.SetVariable(tctx, &entNamespaceVarQuerier{ns: cached.Namespace, clients: flow.edb.Clients(tctx)}, key, req.GetData(), req.GetMimeType(), false)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if newVar {
		flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Created namespace variable '%s'.", key)
	} else {
		flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Updated namespace variable '%s'.", key)
	}

	flow.pubsub.NotifyNamespaceVariables(cached.Namespace.ID)

	var resp grpc.SetNamespaceVariableResponse

	resp.Namespace = cached.Namespace.Name
	resp.Key = key
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.Checksum = vdata.Hash
	resp.TotalSize = int64(vdata.Size)
	resp.MimeType = vdata.MimeType

	return &resp, nil
}

func (internal *internal) SetNamespaceVariableParcels(srv grpc.Internal_SetNamespaceVariableParcelsServer) error {
	internal.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	req, err := srv.Recv()
	if err != nil {
		return err
	}

	instance, err := internal.getInstance(ctx, req.GetInstance())
	if err != nil {
		return err
	}

	mimeType := req.GetMimeType()
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

	var vdata *ent.VarData

	var newVar bool
	vdata, newVar, err = internal.flow.SetVariable(tctx, &entNamespaceVarQuerier{ns: &database.Namespace{ID: instance.Instance.NamespaceID, Name: instance.TelemetryInfo.NamespaceName}, clients: internal.edb.Clients(tctx)}, key, buf.Bytes(), mimeType, false)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newVar {
		internal.logger.Infof(ctx, instance.Instance.NamespaceID, instance.GetAttributes(recipient.Namespace), "Created namespace variable '%s'.", key)
	} else {
		internal.logger.Infof(ctx, instance.Instance.NamespaceID, instance.GetAttributes(recipient.Namespace), "Updated namespace variable '%s'.", key)
	}

	internal.pubsub.NotifyNamespaceVariables(instance.Instance.NamespaceID)

	var resp grpc.SetVariableInternalResponse

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

func (flow *flow) SetNamespaceVariableParcels(srv grpc.Flow_SetNamespaceVariableParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	req, err := srv.Recv()
	if err != nil {
		return err
	}

	mimeType := req.GetMimeType()
	namespace := req.GetNamespace()
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

	cached := new(database.CacheData)

	err = flow.database.NamespaceByName(tctx, cached, namespace)
	if err != nil {
		return err
	}

	var vdata *ent.VarData

	var newVar bool
	vdata, newVar, err = flow.SetVariable(tctx, &entNamespaceVarQuerier{ns: cached.Namespace, clients: flow.edb.Clients(tctx)}, key, buf.Bytes(), mimeType, false)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newVar {
		flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Created namespace variable '%s'.", key)
	} else {
		flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Updated namespace variable '%s'.", key)
	}

	flow.pubsub.NotifyNamespaceVariables(cached.Namespace.ID)

	var resp grpc.SetNamespaceVariableResponse

	resp.Namespace = cached.Namespace.Name
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

func (flow *flow) DeleteNamespaceVariable(ctx context.Context, req *grpc.DeleteNamespaceVariableRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, vref, vdata, err := flow.traverseToNamespaceVariable(tctx, req.GetNamespace(), req.GetKey(), false)
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tctx)

	err = clients.VarRef.DeleteOneID(vref.ID).Exec(tctx)
	if err != nil {
		return nil, err
	}

	if vdata.RefCount == 0 {
		err = clients.VarData.DeleteOneID(vdata.ID).Exec(tctx)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Deleted namespace variable '%s'.", vref.Name)
	flow.pubsub.NotifyNamespaceVariables(cached.Namespace.ID)

	// Broadcast Event
	broadcastInput := broadcastVariableInput{
		WorkflowPath: "",
		Key:          req.GetKey(),
		TotalSize:    int64(vdata.Size),
		Scope:        BroadcastEventScopeNamespace,
	}
	err = flow.BroadcastVariable(ctx, BroadcastEventTypeDelete, BroadcastEventScopeNamespace, broadcastInput, cached.Namespace)
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) RenameNamespaceVariable(ctx context.Context, req *grpc.RenameNamespaceVariableRequest) (*grpc.RenameNamespaceVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, vref, vdata, err := flow.traverseToNamespaceVariable(tctx, req.GetNamespace(), req.GetOld(), false)
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(tctx)

	x, err := clients.VarRef.UpdateOneID(vref.ID).SetName(req.GetNew()).Save(tctx)
	if err != nil {
		return nil, err
	}

	vref.Name = x.Name

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logger.Infof(ctx, cached.Namespace.ID, cached.GetAttributes(recipient.Namespace), "Renamed namespace variable from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyNamespaceVariables(cached.Namespace.ID)

	var resp grpc.RenameNamespaceVariableResponse

	resp.Checksum = vdata.Hash
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.Key = vref.Name
	resp.Namespace = cached.Namespace.Name
	resp.TotalSize = int64(vdata.Size)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.MimeType = vdata.MimeType

	return &resp, nil
}
