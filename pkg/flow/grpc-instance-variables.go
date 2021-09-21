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
		return nil, errors.New("variable too large to return without using the parcelling API")
	}

	resp.Data = d.vdata.Data

	return &resp, nil

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

	more := sub.Wait()
	if !more {
		return nil
	}

	goto resend

}

func (flow *flow) SetInstanceVariable(ctx context.Context, req *grpc.SetInstanceVariableRequest) (*grpc.SetInstanceVariableResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	hash := checksum(req.Data)

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	vrefc := tx.VarRef
	vdatac := tx.VarData

	d, err := flow.getInstance(ctx, nsc, req.GetNamespace(), req.GetInstance(), false)
	if err != nil {
		return nil, err
	}

	var vdata *ent.VarData

	vref, err := d.in.QueryVars().Where(varref.NameEQ(req.GetKey())).Only(ctx)
	if err != nil {

		if !ent.IsNotFound(err) {
			return nil, err
		}

		vdata, err = vdatac.Create().SetSize(len(req.Data)).SetHash(hash).SetData(req.Data).Save(ctx)
		if err != nil {
			return nil, err
		}

		_, err = vrefc.Create().SetVardata(vdata).SetInstance(d.in).SetName(req.GetKey()).Save(ctx)
		if err != nil {
			return nil, err
		}

	} else {

		vdata, err = vref.QueryVardata().Select(vardata.FieldID).Only(ctx)
		if err != nil {
			return nil, err
		}

		vdata, err = vdata.Update().SetSize(len(req.Data)).SetHash(hash).SetData(req.Data).Save(ctx)
		if err != nil {
			return nil, err
		}

	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	flow.logToInstance(ctx, time.Now(), d.in, "Created instance variable '%s'.", vref.Name)
	flow.pubsub.NotifyInstanceVariables(d.in)

	var resp grpc.SetInstanceVariableResponse

	resp.Namespace = d.ns().Name
	resp.Instance = d.in.ID.String()
	resp.Key = vref.Name
	resp.CreatedAt = timestamppb.New(vdata.CreatedAt)
	resp.UpdatedAt = timestamppb.New(vdata.UpdatedAt)
	resp.Checksum = vdata.Hash
	resp.TotalSize = int64(vdata.Size)

	return &resp, nil

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

	hash := checksum(buf.Bytes())

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

	vref, err := d.in.QueryVars().Where(varref.NameEQ(key)).Only(ctx)
	if err != nil {

		if !ent.IsNotFound(err) {
			return err
		}

		vdata, err = vdatac.Create().SetSize(buf.Len()).SetHash(hash).SetData(buf.Bytes()).Save(ctx)
		if err != nil {
			return err
		}

		_, err = vrefc.Create().SetVardata(vdata).SetInstance(d.in).SetName(key).Save(ctx)
		if err != nil {
			return err
		}

	} else {

		vdata, err = vref.QueryVardata().Select(vardata.FieldID).Only(ctx)
		if err != nil {
			return err
		}

		vdata, err = vdata.Update().SetSize(buf.Len()).SetHash(hash).SetData(buf.Bytes()).Save(ctx)
		if err != nil {
			return err
		}

	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	flow.logToInstance(ctx, time.Now(), d.in, "Created instnace variable '%s'.", vref.Name)
	flow.pubsub.NotifyInstanceVariables(d.in)

	var resp grpc.SetInstanceVariableResponse

	resp.Namespace = d.ns().Name
	resp.Instance = d.in.ID.String()
	resp.Key = vref.Name
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
