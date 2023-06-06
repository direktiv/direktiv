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
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	"github.com/google/uuid"

	entns "github.com/direktiv/direktiv/pkg/flow/ent/namespace"
	"github.com/direktiv/direktiv/pkg/flow/ent/vardata"
	entvardata "github.com/direktiv/direktiv/pkg/flow/ent/vardata"
	"github.com/direktiv/direktiv/pkg/flow/ent/varref"
	entvar "github.com/direktiv/direktiv/pkg/flow/ent/varref"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/gabriel-vasile/mimetype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (srv *server) getWorkflowVariable(ctx context.Context, instance *enginerefactor.Instance, key string, load bool) (*database.VarRef, *database.VarData, error) {
	vref, err := srv.database.WorkflowVariable(ctx, instance.Instance.WorkflowID, key)
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

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback()

	f, err = tx.FileStore().ForRootID(ns.ID).GetFile(ctx, path)
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

	instance, err := internal.getInstance(ctx, req.GetInstance())
	if err != nil {
		return err
	}

	vref, vdata, err := internal.getWorkflowVariable(ctx, instance, req.GetKey(), true)
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

func (flow *flow) WorkflowVariables(ctx context.Context, req *grpc.WorkflowVariablesRequest) (*grpc.WorkflowVariablesResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, f, err := flow.getWorkflow(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	clients := flow.edb.Clients(ctx)

	query := clients.VarRef.Query().Where(entvar.WorkflowID(f.ID))

	results, pi, err := paginate[*ent.VarRefQuery, *ent.VarRef](ctx, req.Pagination, query, variablesOrderings, variablesFilters)
	if err != nil {
		return nil, err
	}

	resp := new(grpc.WorkflowVariablesResponse)
	resp.Namespace = ns.Name
	resp.Path = f.Path
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

func (flow *flow) WorkflowVariablesStream(req *grpc.WorkflowVariablesRequest, srv grpc.Flow_WorkflowVariablesStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	ns, f, err := flow.getWorkflow(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflowVariables(f.ID)
	defer flow.cleanup(sub.Close)

resend:

	clients := flow.edb.Clients(ctx)

	query := clients.VarRef.Query().Where(entvar.WorkflowID(f.ID))

	results, pi, err := paginate[*ent.VarRefQuery, *ent.VarRef](ctx, req.Pagination, query, variablesOrderings, variablesFilters)
	if err != nil {
		return err
	}

	resp := new(grpc.WorkflowVariablesResponse)
	resp.Namespace = ns.Name
	resp.Path = f.Path
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

type varQuerier interface {
	QueryVars() *ent.VarRefQuery
}

type entNamespaceVarQuerier struct {
	clients *entwrapper.EntClients
	ns      *database.Namespace
}

func (x *entNamespaceVarQuerier) QueryVars() *ent.VarRefQuery {
	return x.clients.VarRef.Query().Where(entvar.HasNamespaceWith(entns.ID(x.ns.ID)))
}

type entWorkflowVarQuerier struct {
	clients *entwrapper.EntClients
	ns      *database.Namespace
	wfID    uuid.UUID
	path    string
}

func (x *entWorkflowVarQuerier) QueryVars() *ent.VarRefQuery {
	return x.clients.VarRef.Query().Where(entvar.WorkflowID(x.wfID))
}

type entInstanceVarQuerier struct {
	clients  *entwrapper.EntClients
	instance *enginerefactor.Instance
}

func (x *entInstanceVarQuerier) QueryVars() *ent.VarRefQuery {
	return x.clients.VarRef.Query().Where(entvar.InstanceID(x.instance.Instance.ID))
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

	var ns *database.Namespace
	var wfID uuid.UUID

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
			ns = v.ns
			query = query.SetNamespaceID(v.ns.ID)
		case *entWorkflowVarQuerier:
			ns = v.ns
			wfID = v.wfID
			query = query.SetWorkflowID(wfID)
		case *entInstanceVarQuerier:
			ns = &database.Namespace{ID: v.instance.Instance.NamespaceID, Name: v.instance.TelemetryInfo.NamespaceName}
			query = query.SetInstanceID(v.instance.Instance.ID)
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
		ns = v.ns
	case *entWorkflowVarQuerier:
		broadcastInput.Scope = BroadcastEventScopeWorkflow
		ns = v.ns
		wfID = v.wfID
		broadcastInput.WorkflowPath = v.path
	case *entInstanceVarQuerier:
		broadcastInput.Scope = BroadcastEventScopeInstance
		ns = &database.Namespace{ID: v.instance.Instance.NamespaceID, Name: v.instance.TelemetryInfo.NamespaceName}
		broadcastInput.InstanceID = v.instance.Instance.ID.String()
	}

	if newVar {
		err = flow.BroadcastVariable(ctx, BroadcastEventTypeCreate, broadcastInput.Scope, broadcastInput, ns)
	} else {
		err = flow.BroadcastVariable(ctx, BroadcastEventTypeUpdate, broadcastInput.Scope, broadcastInput, ns)
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
	var ns *database.Namespace

	broadcastInput := broadcastVariableInput{
		Key:       key,
		TotalSize: int64(vdata.Size),
	}
	switch v := q.(type) {
	case *entNamespaceVarQuerier:
		broadcastInput.Scope = BroadcastEventScopeNamespace
		ns = v.ns
	case *entWorkflowVarQuerier:
		broadcastInput.Scope = BroadcastEventScopeWorkflow
		ns = v.ns
		broadcastInput.WorkflowPath = v.path
	case *entInstanceVarQuerier:
		broadcastInput.Scope = BroadcastEventScopeInstance
		broadcastInput.InstanceID = v.instance.Instance.ID.String()
		ns = &database.Namespace{ID: v.instance.Instance.NamespaceID, Name: v.instance.TelemetryInfo.NamespaceName}
	}

	err = flow.BroadcastVariable(ctx, BroadcastEventTypeDelete, broadcastInput.Scope, broadcastInput, ns)

	return vdata, newVar, err
}

func (flow *flow) SetWorkflowVariable(ctx context.Context, req *grpc.SetWorkflowVariableRequest) (*grpc.SetWorkflowVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, file, err := flow.getWorkflow(ctx, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	var vdata *ent.VarData

	key := req.GetKey()

	var newVar bool
	vdata, newVar, err = flow.SetVariable(tctx, &entWorkflowVarQuerier{clients: flow.edb.Clients(tctx), ns: ns, wfID: file.ID, path: file.Path}, key, req.GetData(), req.GetMimeType(), false)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if newVar {
		flow.logger.Infof(ctx, file.ID, database.GetAttributes(recipient.Workflow, ns, fileAttributes(*file)), "Created workflow variable '%s'.", key)
	} else {
		flow.logger.Infof(ctx, file.ID, database.GetAttributes(recipient.Workflow, ns, fileAttributes(*file)), "Updated workflow variable '%s'.", key)
	}

	flow.pubsub.NotifyWorkflowVariables(file.ID)

	var resp grpc.SetWorkflowVariableResponse

	resp.Namespace = ns.Name
	resp.Path = file.Path
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
	vdata, newVar, err = internal.flow.SetVariable(tctx, &entWorkflowVarQuerier{clients: internal.edb.Clients(tctx), ns: &database.Namespace{ID: instance.Instance.NamespaceID, Name: instance.TelemetryInfo.NamespaceName}, wfID: instance.Instance.WorkflowID, path: instance.Instance.CalledAs}, key, buf.Bytes(), mimeType, false)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newVar {
		internal.logger.Infof(ctx, instance.Instance.WorkflowID, instance.GetAttributes(recipient.Workflow), "Created workflow variable '%s'.", key)
	} else {
		internal.logger.Infof(ctx, instance.Instance.WorkflowID, instance.GetAttributes(recipient.Workflow), "Updated workflow variable '%s'.", key)
	}

	internal.pubsub.NotifyWorkflowVariables(instance.Instance.WorkflowID)

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

	namespace := req.GetNamespace()
	path := req.GetPath()

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

	ns, file, err := flow.getWorkflow(ctx, namespace, path)
	if err != nil {
		return err
	}

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	var vdata *ent.VarData

	var newVar bool
	vdata, newVar, err = flow.SetVariable(tctx, &entWorkflowVarQuerier{clients: flow.edb.Clients(tctx), ns: ns, wfID: file.ID, path: file.Path}, key, buf.Bytes(), mimeType, false)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newVar {
		flow.logger.Infof(ctx, file.ID, database.GetAttributes(recipient.Workflow, ns, fileAttributes(*file)), "Created workflow variable '%s'.", key)
	} else {
		flow.logger.Infof(ctx, file.ID, database.GetAttributes(recipient.Workflow, ns, fileAttributes(*file)), "Updated workflow variable '%s'.", key)
	}

	flow.pubsub.NotifyWorkflowVariables(file.ID)

	var resp grpc.SetWorkflowVariableResponse

	resp.Namespace = ns.Name
	resp.Path = file.Path
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

	ns, file, vref, vdata, err := flow.getWorkflowVariable(ctx, req.GetNamespace(), req.GetPath(), req.GetKey(), false)
	if err != nil {
		return nil, err
	}

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

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

	flow.logger.Infof(ctx, file.ID, database.GetAttributes(recipient.Workflow, ns, fileAttributes(*file)), "Deleted workflow variable '%s'.", vref.Name)
	flow.pubsub.NotifyWorkflowVariables(file.ID)

	// Broadcast Event
	broadcastInput := broadcastVariableInput{
		WorkflowPath: req.GetPath(),
		Key:          req.GetKey(),
		TotalSize:    int64(vdata.Size),
		Scope:        BroadcastEventScopeWorkflow,
	}
	err = flow.BroadcastVariable(ctx, BroadcastEventTypeDelete, BroadcastEventScopeNamespace, broadcastInput, ns)
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) RenameWorkflowVariable(ctx context.Context, req *grpc.RenameWorkflowVariableRequest) (*grpc.RenameWorkflowVariableResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, file, vref, vdata, err := flow.getWorkflowVariable(ctx, req.GetNamespace(), req.GetPath(), req.GetOld(), false)
	if err != nil {
		return nil, err
	}

	tctx, tx, err := flow.database.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

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

	flow.logger.Infof(ctx, file.ID, database.GetAttributes(recipient.Workflow, ns, fileAttributes(*file)), "Renamed workflow variable from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyWorkflowVariables(file.ID)

	var resp grpc.RenameWorkflowVariableResponse

	resp.Checksum = vdata.Hash
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.Key = vref.Name
	resp.Namespace = ns.Name
	resp.TotalSize = int64(vdata.Size)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.MimeType = vdata.MimeType

	return &resp, nil
}
