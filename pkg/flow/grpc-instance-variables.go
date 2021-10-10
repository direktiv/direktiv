package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vorteil/direktiv/pkg/flow/ent"
	entvardata "github.com/vorteil/direktiv/pkg/flow/ent/vardata"
	"github.com/vorteil/direktiv/pkg/flow/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (flow *flow) InstanceVariable(ctx context.Context, req *grpc.InstanceVariableRequest) (*grpc.InstanceVariableResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	nsc := flow.db.Namespace

	d, err := flow.traverseToInstanceVariable(ctx, nsc, req.GetNamespace(), req.GetInstance(), req.GetKey(), true)
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceVariableResponse

	resp.Namespace = d.ns().Name
	resp.Instance = d.in.ID.String()
	resp.Key = d.vref.Name
	resp.CreatedAt = timestamppb.New(d.vdata.CreatedAt)
	resp.UpdatedAt = timestamppb.New(d.vdata.UpdatedAt)
	resp.Checksum = d.vdata.Hash
	resp.TotalSize = int64(d.vdata.Size)

	if resp.TotalSize > parcelSize {
		return nil, status.Error(codes.ResourceExhausted, "variable too large to return without using the parcelling API")
	}

	resp.Data = d.vdata.Data

	return &resp, nil

}

func (internal *internal) InstanceVariableParcels(req *grpc.VariableInternalRequest, srv grpc.Internal_InstanceVariableParcelsServer) error {

	internal.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	nsc := internal.db.Namespace
	inc := internal.db.Instance

	id, err := internal.getInstance(ctx, inc, req.GetInstance(), false)
	if err != nil {
		return err
	}

	d, err := internal.traverseToInstanceVariable(ctx, nsc, id.namespace(), req.GetInstance(), req.GetKey(), true)
	if err != nil && !IsNotFound(err) {
		return err
	}

	if IsNotFound(err) {
		d = new(instvarData)
		d.vdata = new(ent.VarData)
		t := time.Now()
		d.vdata.Data = make([]byte, 0)
		hash, err := computeHash(d.vdata.Data)
		if err != nil {
			internal.sugar.Error(err)
		}
		d.vdata.CreatedAt = t
		d.vdata.UpdatedAt = t
		d.vdata.Hash = hash
		d.vdata.Size = 0
	}

	rdr := bytes.NewReader(d.vdata.Data)

	for {

		resp := new(grpc.VariableInternalResponse)

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

func (flow *flow) InstanceVariableParcels(req *grpc.InstanceVariableRequest, srv grpc.Flow_InstanceVariableParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	nsc := flow.db.Namespace

	d, err := flow.traverseToInstanceVariable(ctx, nsc, req.GetNamespace(), req.GetInstance(), req.GetKey(), true)
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(d.vdata.Data)

	for {

		resp := new(grpc.InstanceVariableResponse)

		resp.Namespace = d.ns().Name
		resp.Instance = d.in.ID.String()
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

func (flow *flow) InstanceVariables(ctx context.Context, req *grpc.InstanceVariablesRequest) (*grpc.InstanceVariablesResponse, error) {

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

	d, err := flow.getInstance(ctx, nsc, req.GetNamespace(), req.GetInstance(), false)
	if err != nil {
		return nil, err
	}

	query := d.in.QueryVars()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceVariablesResponse

	resp.Namespace = d.ns().Name
	resp.Instance = d.in.ID.String()

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

func (flow *flow) InstanceVariablesStream(req *grpc.InstanceVariablesRequest, srv grpc.Flow_InstanceVariablesStreamServer) error {

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
	d, err := flow.getInstance(ctx, nsc, req.GetNamespace(), req.GetInstance(), false)
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeInstanceVariables(d.in)
	defer flow.cleanup(sub.Close)

resend:

	query := d.in.QueryVars()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	resp := new(grpc.InstanceVariablesResponse)

	resp.Namespace = d.ns().Name
	resp.Instance = d.in.ID.String()

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

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend

}

func (flow *flow) SetInstanceVariable(ctx context.Context, req *grpc.SetInstanceVariableRequest) (*grpc.SetInstanceVariableResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	vrefc := tx.VarRef
	vdatac := tx.VarData

	key := req.GetKey()

	d, err := flow.getInstance(ctx, nsc, req.GetNamespace(), req.GetInstance(), false)
	if err != nil {
		return nil, err
	}

	var vdata *ent.VarData

	vdata, err = flow.SetVariable(ctx, vrefc, vdatac, d.in, key, req.GetData())
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToInstance(ctx, time.Now(), d.in, "Created instance variable '%s'.", key)
	flow.pubsub.NotifyInstanceVariables(d.in)

	var resp grpc.SetInstanceVariableResponse

	resp.Namespace = d.ns().Name
	resp.Instance = d.in.ID.String()
	resp.Key = key
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.Checksum = vdata.Hash
	resp.TotalSize = int64(vdata.Size)

	return &resp, nil

}

func (internal *internal) SetInstanceVariableParcels(srv grpc.Internal_SetInstanceVariableParcelsServer) error {

	internal.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	req, err := srv.Recv()
	if err != nil {
		return err
	}

	inc := internal.db.Instance

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

	tx, err := internal.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	inc = tx.Instance
	vrefc := tx.VarRef
	vdatac := tx.VarData

	d, err := internal.getInstance(ctx, inc, instance, false)
	if err != nil {
		return err
	}

	var vdata *ent.VarData

	vdata, err = internal.flow.SetVariable(ctx, vrefc, vdatac, d.in, key, req.GetData())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	internal.logToInstance(ctx, time.Now(), d.in, "Created instance variable '%s'.", key)
	internal.pubsub.NotifyInstanceVariables(d.in)

	var resp grpc.SetVariableInternalResponse

	resp.Instance = d.in.ID.String()
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

func (flow *flow) SetInstanceVariableParcels(srv grpc.Flow_SetInstanceVariableParcelsServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	req, err := srv.Recv()
	if err != nil {
		return err
	}

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

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	vrefc := tx.VarRef
	vdatac := tx.VarData

	d, err := flow.getInstance(ctx, nsc, namespace, instance, false)
	if err != nil {
		return err
	}

	var vdata *ent.VarData

	vdata, err = flow.SetVariable(ctx, vrefc, vdatac, d.in, key, req.GetData())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	flow.logToInstance(ctx, time.Now(), d.in, "Created instance variable '%s'.", key)
	flow.pubsub.NotifyInstanceVariables(d.in)

	var resp grpc.SetInstanceVariableResponse

	resp.Namespace = d.ns().Name
	resp.Instance = d.in.ID.String()
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

func (flow *flow) DeleteInstanceVariable(ctx context.Context, req *grpc.DeleteInstanceVariableRequest) (*emptypb.Empty, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace

	d, err := flow.traverseToInstanceVariable(ctx, nsc, req.GetNamespace(), req.GetInstance(), req.GetKey(), false)
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

	flow.logToInstance(ctx, time.Now(), d.in, "Deleted instance variable '%s'.", d.vref.Name)
	flow.pubsub.NotifyInstanceVariables(d.in)

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) RenameInstanceVariable(ctx context.Context, req *grpc.RenameInstanceVariableRequest) (*grpc.RenameInstanceVariableResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToInstanceVariable(ctx, nsc, req.GetNamespace(), req.GetInstance(), req.GetOld(), false)
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

	flow.logToInstance(ctx, time.Now(), d.in, "Renamed instance variable from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyInstanceVariables(d.in)

	var resp grpc.RenameInstanceVariableResponse

	resp.Checksum = d.vdata.Hash
	resp.CreatedAt = timestamppb.New(d.vdata.CreatedAt)
	resp.Key = vref.Name
	resp.Namespace = d.ns().Name
	resp.TotalSize = int64(d.vdata.Size)
	resp.UpdatedAt = timestamppb.New(d.vdata.UpdatedAt)

	return &resp, nil

}
