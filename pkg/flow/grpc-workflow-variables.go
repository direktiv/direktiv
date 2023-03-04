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
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	"github.com/direktiv/direktiv/pkg/flow/ent/vardata"
	entvardata "github.com/direktiv/direktiv/pkg/flow/ent/vardata"
	"github.com/direktiv/direktiv/pkg/flow/ent/varref"
	entvar "github.com/direktiv/direktiv/pkg/flow/ent/varref"
	entwf "github.com/direktiv/direktiv/pkg/flow/ent/workflow"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/gabriel-vasile/mimetype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (srv *server) getWorkflowVariable(ctx context.Context, cached *database.CacheData, key string, load bool) (*database.VarRef, *database.VarData, error) {
	vref, err := srv.database.WorkflowVariable(ctx, cached.Workflow.ID, key)
	if err != nil {
		return nil, nil, err
	}

	vdata, err := srv.database.VariableData(ctx, vref.VarData, load)
	if err != nil {
		return nil, nil, err
	}

	return vref, vdata, nil
}

func (srv *server) traverseToWorkflowVariable(ctx context.Context, namespace, path, key string, load bool) (*database.CacheData, *database.VarRef, *database.VarData, error) {
	cached, err := srv.traverseToWorkflow(ctx, namespace, path)
	if err != nil {
		return nil, nil, nil, err
	}

	vref, vdata, err := srv.getWorkflowVariable(ctx, cached, key, load)
	if err != nil {
		return nil, nil, nil, err
	}

	return cached, vref, vdata, nil
}

func (flow *flow) WorkflowVariable(ctx context.Context, req *grpc.WorkflowVariableRequest) (*grpc.WorkflowVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, vref, vdata, err := flow.traverseToWorkflowVariable(ctx, req.GetNamespace(), req.GetPath(), req.GetKey(), true)
	if err != nil {
		return nil, err
	}

	var resp grpc.WorkflowVariableResponse

	resp.Namespace = cached.Namespace.Name
	resp.Path = cached.Path()
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

	cached, vref, vdata, err := flow.traverseToWorkflowVariable(ctx, req.GetNamespace(), req.GetPath(), req.GetKey(), true)
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(vdata.Data)

	for {

		resp := new(grpc.WorkflowVariableResponse)

		resp.Namespace = cached.Namespace.Name
		resp.Path = cached.Path()
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

func (flow *flow) WorkflowVariables(ctx context.Context, req *grpc.WorkflowVariablesRequest) (*grpc.WorkflowVariablesResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	cached, err := flow.traverseToWorkflow(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	query := clients.VarRef.Query().Where(entvar.HasWorkflowWith(entwf.ID(cached.Workflow.ID)))

	results, pi, err := paginate[*ent.VarRefQuery, *ent.VarRef](ctx, req.Pagination, query, variablesOrderings, variablesFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.WorkflowVariablesResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Path = cached.Path()
	resp.Variables = new(grpc.Variables)
	resp.Variables.PageInfo = pi

	err = bytedata.Atob(results, &resp.Variables.Results)
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

func (flow *flow) WorkflowVariablesStream(req *grpc.WorkflowVariablesRequest, srv grpc.Flow_WorkflowVariablesStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	cached, err := flow.traverseToWorkflow(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflowVariables(cached)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(ctx)

	query := clients.VarRef.Query().Where(entvar.HasWorkflowWith(entwf.ID(cached.Workflow.ID)))

	results, pi, err := paginate[*ent.VarRefQuery, *ent.VarRef](ctx, req.Pagination, query, variablesOrderings, variablesFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.WorkflowVariablesResponse)
	resp.Namespace = cached.Namespace.Name
	resp.Path = cached.Path()
	resp.Variables = new(grpc.Variables)
	resp.Variables.PageInfo = pi

	err = bytedata.Atob(results, &resp.Variables.Results)
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

type varQuerier interface {
	QueryVars() *ent.VarRefQuery
}

type entNamespaceVarQuerier struct {
	clients *entwrapper.EntClients
	cached  *database.CacheData
}

func (x *entNamespaceVarQuerier) QueryVars() *ent.VarRefQuery {
	return x.clients.VarRef.Query().Where(entvar.HasNamespaceWith(entns.ID(x.cached.Namespace.ID)))
}

type entWorkflowVarQuerier struct {
	clients *entwrapper.EntClients
	cached  *database.CacheData
}

func (x *entWorkflowVarQuerier) QueryVars() *ent.VarRefQuery {
	return x.clients.VarRef.Query().Where(entvar.HasWorkflowWith(entwf.ID(x.cached.Workflow.ID)))
}

type entInstanceVarQuerier struct {
	clients *entwrapper.EntClients
	cached  *database.CacheData
}

func (x *entInstanceVarQuerier) QueryVars() *ent.VarRefQuery {
	return x.clients.VarRef.Query().Where(entvar.HasInstanceWith(entinst.ID(x.cached.Instance.ID)))
}

func (flow *flow) SetVariable(ctx context.Context, q varQuerier, key string, data []byte, vMimeType string, thread bool) (*ent.VarData, bool, error) {
	hash, err := bytedata.ComputeHash(data)
	if err != nil {
		flow.sugar.Error(err)
	}

	var vdata *ent.VarData
	var newVar bool

	var vref *ent.VarRef

	if thread {
		vref, err = q.QueryVars().Where(varref.NameEQ(key), varref.BehaviourContains("thread")).Only(ctx)
	} else {
		vref, err = q.QueryVars().Where(varref.NameEQ(key), varref.BehaviourIsNil()).Only(ctx)
	}

	clients := flow.edb.Clients(ctx)

	var cached *database.CacheData

	if err != nil {

		if !derrors.IsNotFound(err) {
			return nil, false, err
		}

		vdataBuilder := clients.VarData.Create().SetSize(len(data)).SetHash(hash)
		// set mime type if provided
		if vMimeType != "" {
			vdataBuilder.SetMimeType(vMimeType)
		} else {
			// auto detect
			mtype := mimetype.Detect(data)
			vdataBuilder.SetMimeType(mtype.String())
		}

		vdata, err = vdataBuilder.SetData(data).Save(ctx)
		if err != nil {
			return nil, false, err
		}

		query := clients.VarRef.Create().SetVardata(vdata).SetName(key)

		switch v := q.(type) {
		case *entNamespaceVarQuerier:
			cached = v.cached
			query = query.SetNamespaceID(v.cached.Namespace.ID)
		case *entWorkflowVarQuerier:
			cached = v.cached
			query = query.SetWorkflowID(v.cached.Workflow.ID)
		case *entInstanceVarQuerier:
			cached = v.cached
			query = query.SetInstanceID(v.cached.Instance.ID)
			if thread {
				query = query.SetBehaviour("thread")
			}
		default:
			panic(errors.New("bad querier"))
		}

		_, err = query.Save(ctx)
		if err != nil {
			return nil, false, err
		}

		newVar = true
	} else {

		vdata, err = vref.QueryVardata().Select(vardata.FieldID).Only(ctx)
		if err != nil {
			return nil, false, err
		}

		vdataBuilder := vdata.Update().SetSize(len(data)).SetHash(hash).SetData(data)

		// Update mime type if provided
		if vMimeType != "" {
			vdataBuilder.SetMimeType(vMimeType)
		}

		vdata, err = vdataBuilder.Save(ctx)
		if err != nil {
			return nil, false, err
		}

	}

	// Broadcast Event
	broadcastInput := broadcastVariableInput{
		Key:       key,
		TotalSize: int64(vdata.Size),
	}
	switch v := q.(type) {
	case *entNamespaceVarQuerier:
		broadcastInput.Scope = BroadcastEventScopeNamespace
		cached = v.cached
	case *entWorkflowVarQuerier:
		broadcastInput.Scope = BroadcastEventScopeWorkflow
		cached = v.cached
		broadcastInput.WorkflowPath = cached.Path()
	case *entInstanceVarQuerier:
		broadcastInput.Scope = BroadcastEventScopeInstance
		cached = v.cached
		broadcastInput.InstanceID = v.cached.Instance.ID.String()
	}

	if newVar {
		err = flow.BroadcastVariable(ctx, BroadcastEventTypeCreate, broadcastInput.Scope, broadcastInput, cached)
	} else {
		err = flow.BroadcastVariable(ctx, BroadcastEventTypeUpdate, broadcastInput.Scope, broadcastInput, cached)
	}

	return vdata, newVar, err
}

func (flow *flow) DeleteVariable(ctx context.Context, q varQuerier, key string, data []byte, vMimeType string, thread bool) (*ent.VarData, bool, error) {
	var err error
	var vdata *ent.VarData
	var newVar bool

	var vref *ent.VarRef

	if thread {
		vref, err = q.QueryVars().Where(varref.NameEQ(key), varref.BehaviourContains("thread")).Only(ctx)
	} else {
		vref, err = q.QueryVars().Where(varref.NameEQ(key), varref.BehaviourIsNil()).Only(ctx)
	}

	if err != nil {
		return nil, false, err
	}

	clients := flow.edb.Clients(ctx)

	vdata, err = vref.QueryVardata().Select(vardata.FieldID).Only(ctx)
	if err != nil {
		return nil, false, err
	}
	_, err = clients.VarRef.Delete().Where(varref.NameEQ(key)).Exec(ctx)
	if err != nil {
		return nil, false, err
	}

	k, err := vdata.QueryVarrefs().Count(ctx)
	if err != nil {
		return nil, false, err
	}

	if k == 0 {
		err = clients.VarData.DeleteOne(vdata).Exec(ctx)
		if err != nil {
			return nil, false, err
		}
	}

	// Broadcast Event
	var cached *database.CacheData
	broadcastInput := broadcastVariableInput{
		Key:       key,
		TotalSize: int64(vdata.Size),
	}
	switch v := q.(type) {
	case *entNamespaceVarQuerier:
		broadcastInput.Scope = BroadcastEventScopeNamespace
		cached = v.cached
	case *entWorkflowVarQuerier:
		broadcastInput.Scope = BroadcastEventScopeWorkflow
		cached = v.cached
		broadcastInput.WorkflowPath = cached.Path()
	case *entInstanceVarQuerier:
		broadcastInput.Scope = BroadcastEventScopeInstance
		broadcastInput.InstanceID = v.cached.Instance.ID.String()
		cached = v.cached
	}

	err = flow.BroadcastVariable(ctx, BroadcastEventTypeDelete, broadcastInput.Scope, broadcastInput, cached)

	return vdata, newVar, err
}

func (flow *flow) SetWorkflowVariable(ctx context.Context, req *grpc.SetWorkflowVariableRequest) (*grpc.SetWorkflowVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, err := flow.traverseToWorkflow(tctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	var vdata *ent.VarData

	key := req.GetKey()

	var newVar bool
	vdata, newVar, err = flow.SetVariable(tctx, &entWorkflowVarQuerier{clients: flow.edb.Clients(tctx), cached: cached}, key, req.GetData(), req.GetMimeType(), false)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if newVar {
		flow.logToWorkflow(ctx, time.Now(), cached, "Created workflow variable '%s'.", key)
	} else {
		flow.logToWorkflow(ctx, time.Now(), cached, "Updated workflow variable '%s'.", key)
	}

	flow.pubsub.NotifyWorkflowVariables(cached.Workflow)

	var resp grpc.SetWorkflowVariableResponse

	resp.Namespace = cached.Namespace.Name
	resp.Path = cached.Path()
	resp.Key = key
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.Checksum = vdata.Hash
	resp.TotalSize = int64(vdata.Size)
	resp.MimeType = vdata.MimeType

	return &resp, nil
}

func (internal *internal) SetWorkflowVariableParcels(srv grpc.Internal_SetWorkflowVariableParcelsServer) error {
	internal.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	req, err := srv.Recv()
	if err != nil {
		return err
	}

	cached, err := internal.getInstance(ctx, req.GetInstance())
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
	vdata, newVar, err = internal.flow.SetVariable(tctx, &entWorkflowVarQuerier{clients: internal.edb.Clients(tctx), cached: cached}, key, buf.Bytes(), mimeType, false)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newVar {
		internal.logToWorkflow(ctx, time.Now(), cached, "Created workflow variable '%s'.", key)
	} else {
		internal.logToWorkflow(ctx, time.Now(), cached, "Updated workflow variable '%s'.", key)
	}

	internal.pubsub.NotifyWorkflowVariables(cached.Workflow)

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

func (flow *flow) SetWorkflowVariableParcels(srv grpc.Flow_SetWorkflowVariableParcelsServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	req, err := srv.Recv()
	if err != nil {
		return err
	}

	mimeType := req.GetMimeType()
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

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	cached, err := flow.traverseToWorkflow(tctx, namespace, path)
	if err != nil {
		return err
	}

	var vdata *ent.VarData

	var newVar bool
	vdata, newVar, err = flow.SetVariable(tctx, &entWorkflowVarQuerier{clients: flow.edb.Clients(tctx), cached: cached}, key, buf.Bytes(), mimeType, false)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newVar {
		flow.logToWorkflow(ctx, time.Now(), cached, "Created workflow variable '%s'.", key)
	} else {
		flow.logToWorkflow(ctx, time.Now(), cached, "Updated workflow variable '%s'.", key)
	}

	flow.pubsub.NotifyWorkflowVariables(cached.Workflow)

	var resp grpc.SetWorkflowVariableResponse

	resp.Namespace = cached.Namespace.Name
	resp.Path = cached.Path()
	resp.Key = key
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.Checksum = vdata.Hash
	resp.TotalSize = int64(vdata.Size)

	err = srv.SendAndClose(&resp)
	if err != nil {
		return err
	}

	return nil
}

func (flow *flow) DeleteWorkflowVariable(ctx context.Context, req *grpc.DeleteWorkflowVariableRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, vref, vdata, err := flow.traverseToWorkflowVariable(tctx, req.GetNamespace(), req.GetPath(), req.GetKey(), false)
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

	flow.logToWorkflow(ctx, time.Now(), cached, "Deleted workflow variable '%s'.", vref.Name)
	flow.pubsub.NotifyWorkflowVariables(cached.Workflow)

	// Broadcast Event
	broadcastInput := broadcastVariableInput{
		WorkflowPath: req.GetPath(),
		Key:          req.GetKey(),
		TotalSize:    int64(vdata.Size),
		Scope:        BroadcastEventScopeWorkflow,
	}
	err = flow.BroadcastVariable(ctx, BroadcastEventTypeDelete, BroadcastEventScopeNamespace, broadcastInput, cached)
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) RenameWorkflowVariable(ctx context.Context, req *grpc.RenameWorkflowVariableRequest) (*grpc.RenameWorkflowVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	cached, vref, vdata, err := flow.traverseToWorkflowVariable(tctx, req.GetNamespace(), req.GetPath(), req.GetOld(), false)
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

	flow.logToWorkflow(ctx, time.Now(), cached, "Renamed workflow variable from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyWorkflowVariables(cached.Workflow)

	var resp grpc.RenameWorkflowVariableResponse

	resp.Checksum = vdata.Hash
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.Key = vref.Name
	resp.Namespace = cached.Namespace.Name
	resp.TotalSize = int64(vdata.Size)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.MimeType = vdata.MimeType

	return &resp, nil
}
