package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	"github.com/vorteil/direktiv/pkg/flow/ent"
	"github.com/vorteil/direktiv/pkg/flow/ent/vardata"
	entvardata "github.com/vorteil/direktiv/pkg/flow/ent/vardata"
	"github.com/vorteil/direktiv/pkg/flow/ent/varref"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (flow *flow) WorkflowVariable(ctx context.Context, req *grpc.WorkflowVariableRequest) (*grpc.WorkflowVariableResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	nsc := flow.db.Namespace

	d, err := flow.traverseToWorkflowVariable(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetKey(), true)
	if err != nil {
		return nil, err
	}

	var resp grpc.WorkflowVariableResponse

	resp.Namespace = d.ns().Name
	resp.Path = d.path
	resp.Key = d.vref.Name
	resp.CreatedAt = timestamppb.New(d.vdata.CreatedAt)
	resp.UpdatedAt = timestamppb.New(d.vdata.UpdatedAt)
	resp.Checksum = d.vdata.Hash
	resp.TotalSize = int64(d.vdata.Size)

	if resp.TotalSize > parcelSize {
		return nil, errors.New("variable too large to return without using the parcelling API")
	}

	resp.Data = d.vdata.Data

	return &resp, nil

}

func (flow *flow) WorkflowVariableParcels(req *grpc.WorkflowVariableRequest, srv grpc.Flow_WorkflowVariableParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	nsc := flow.db.Namespace

	d, err := flow.traverseToWorkflowVariable(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetKey(), true)
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(d.vdata.Data)

	for {

		resp := new(grpc.WorkflowVariableResponse)

		resp.Namespace = d.ns().Name
		resp.Path = d.path
		resp.Key = d.vref.Name
		resp.CreatedAt = timestamppb.New(d.vdata.CreatedAt)
		resp.UpdatedAt = timestamppb.New(d.vdata.UpdatedAt)
		resp.Checksum = d.vdata.Hash
		resp.TotalSize = int64(d.vdata.Size)

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

func (flow *flow) WorkflowVariables(ctx context.Context, req *grpc.WorkflowVariablesRequest) (*grpc.WorkflowVariablesResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.VarRefPaginateOption{}
	opts = append(opts, variablesOrder(p))
	filter := variablesFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace

	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	query := d.wf.QueryVars()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.WorkflowVariablesResponse

	resp.Namespace = d.ns().Name
	resp.Path = d.path

	err = atob(cx, &resp.Variables)
	if err != nil {
		return nil, err
	}

	for i := range cx.Edges {

		edge := cx.Edges[i]
		vref := edge.Node

		vdata, err := vref.QueryVardata().Select(entvardata.FieldCreatedAt, entvardata.FieldHash, entvardata.FieldSize, entvardata.FieldUpdatedAt).Only(ctx)
		if err != nil {
			return nil, err
		}

		v := resp.Variables.Edges[i].Node
		v.Checksum = vdata.Hash
		v.CreatedAt = timestamppb.New(vdata.CreatedAt)
		v.Size = int64(vdata.Size)
		v.UpdatedAt = timestamppb.New(vdata.UpdatedAt)

	}

	return &resp, nil

}

func (flow *flow) WorkflowVariablesStream(req *grpc.WorkflowVariablesRequest, srv grpc.Flow_WorkflowVariablesStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	opts := []ent.VarRefPaginateOption{}
	opts = append(opts, variablesOrder(p))
	filter := variablesFilter(p)
	if filter != nil {
		opts = append(opts, filter)
	}

	nsc := flow.db.Namespace
	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeWorkflowVariables(d.wf)
	defer flow.cleanup(sub.Close)

resend:

	query := d.wf.QueryVars()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	resp := new(grpc.WorkflowVariablesResponse)

	resp.Namespace = d.ns().Name
	resp.Path = d.path

	err = atob(cx, &resp.Variables)
	if err != nil {
		return err
	}

	for i := range cx.Edges {

		edge := cx.Edges[i]
		vref := edge.Node

		vdata, err := vref.QueryVardata().Select(entvardata.FieldCreatedAt, entvardata.FieldHash, entvardata.FieldSize, entvardata.FieldUpdatedAt).Only(ctx)
		if err != nil {
			return err
		}

		v := resp.Variables.Edges[i].Node
		v.Checksum = vdata.Hash
		v.CreatedAt = timestamppb.New(vdata.CreatedAt)
		v.Size = int64(vdata.Size)
		v.UpdatedAt = timestamppb.New(vdata.UpdatedAt)

	}

	nhash = checksum(resp)
	if nhash != phash {
		err = srv.Send(resp)
		if err != nil {
			return err
		}
	}
	phash = nhash

	more := sub.Wait()
	if !more {
		return nil
	}

	goto resend

}

type varQuerier interface {
	QueryVars() *ent.VarRefQuery
}

func (flow *flow) SetVariable(ctx context.Context, vrefc *ent.VarRefClient, vdatac *ent.VarDataClient, q varQuerier, key string, data []byte) error {

	hash := checksum(data)

	vref, err := q.QueryVars().Where(varref.NameEQ(key)).Only(ctx)
	if err != nil {

		if !ent.IsNotFound(err) {
			return err
		}

		vdata, err := vdatac.Create().SetSize(len(data)).SetHash(hash).SetData(data).Save(ctx)
		if err != nil {
			return err
		}

		query := vrefc.Create().SetVardata(vdata).SetName(key)

		switch q.(type) {
		case *ent.Namespace:
			query = query.SetNamespace(q.(*ent.Namespace))
		case *ent.Workflow:
			query = query.SetWorkflow(q.(*ent.Workflow))
		case *ent.Instance:
			query = query.SetInstance(q.(*ent.Instance))
		default:
			panic(errors.New("bad querier"))
		}

		_, err = query.Save(ctx)
		if err != nil {
			return err
		}

	} else {

		vdata, err := vref.QueryVardata().Select(vardata.FieldID).Only(ctx)
		if err != nil {
			return err
		}

		vdata, err = vdata.Update().SetSize(len(data)).SetHash(hash).SetData(data).Save(ctx)
		if err != nil {
			return err
		}

	}

	return nil

}

func (flow *flow) SetWorkflowVariable(ctx context.Context, req *grpc.SetWorkflowVariableRequest) (*grpc.SetWorkflowVariableResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	vrefc := tx.VarRef
	vdatac := tx.VarData

	d, err := flow.traverseToWorkflow(ctx, nsc, req.GetNamespace(), req.GetPath())
	if err != nil {
		return nil, err
	}

	var vdata *ent.VarData

	key := req.GetKey()

	err = flow.SetVariable(ctx, vrefc, vdatac, d.wf, key, req.GetData())
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), d.wf, "Created workflow variable '%s'.", key)
	flow.pubsub.NotifyWorkflowVariables(d.wf)

	var resp grpc.SetWorkflowVariableResponse

	resp.Namespace = d.ns().Name
	resp.Path = d.path
	resp.Key = key
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.Checksum = vdata.Hash
	resp.TotalSize = int64(vdata.Size)

	return &resp, nil

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

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	vrefc := tx.VarRef
	vdatac := tx.VarData

	d, err := flow.traverseToWorkflow(ctx, nsc, namespace, path)
	if err != nil {
		return err
	}

	var vdata *ent.VarData

	err = flow.SetVariable(ctx, vrefc, vdatac, d.wf, key, buf.Bytes())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	flow.logToWorkflow(ctx, time.Now(), d.wf, "Created workflow variable '%s'.", key)
	flow.pubsub.NotifyWorkflowVariables(d.wf)

	var resp grpc.SetWorkflowVariableResponse

	resp.Namespace = d.ns().Name
	resp.Path = d.path
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

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace

	d, err := flow.traverseToWorkflowVariable(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetKey(), false)
	if err != nil {
		return nil, err
	}

	vrefc := tx.VarRef
	vdatac := tx.VarData

	err = vrefc.DeleteOne(d.vref).Exec(ctx)
	if err != nil {
		return nil, err
	}

	k, err := d.vdata.QueryVarrefs().Count(ctx)
	if err != nil {
		return nil, err
	}

	if k == 0 {
		err = vdatac.DeleteOne(d.vdata).Exec(ctx)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), d.wf, "Deleted workflow variable '%s'.", d.vref.Name)
	flow.pubsub.NotifyWorkflowVariables(d.wf)

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) RenameWorkflowVariable(ctx context.Context, req *grpc.RenameWorkflowVariableRequest) (*grpc.RenameWorkflowVariableResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToWorkflowVariable(ctx, nsc, req.GetNamespace(), req.GetPath(), req.GetOld(), false)
	if err != nil {
		return nil, err
	}

	vref, err := d.vref.Update().SetName(req.GetNew()).Save(ctx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToWorkflow(ctx, time.Now(), d.wf, "Renamed workflow variable from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyWorkflowVariables(d.wf)

	var resp grpc.RenameWorkflowVariableResponse

	resp.Checksum = d.vdata.Hash
	resp.CreatedAt = timestamppb.New(d.vdata.CreatedAt)
	resp.Key = vref.Name
	resp.Namespace = d.ns().Name
	resp.TotalSize = int64(d.vdata.Size)
	resp.UpdatedAt = timestamppb.New(d.vdata.UpdatedAt)

	return &resp, nil

}
