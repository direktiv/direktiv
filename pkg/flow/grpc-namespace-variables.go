package flow

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/ent"
	entvardata "github.com/direktiv/direktiv/pkg/flow/ent/vardata"
	entvar "github.com/direktiv/direktiv/pkg/flow/ent/varref"
	derrors "github.com/direktiv/direktiv/pkg/flow/errors"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (flow *flow) NamespaceVariable(ctx context.Context, req *grpc.NamespaceVariableRequest) (*grpc.NamespaceVariableResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	nsc := flow.db.Namespace

	d, err := flow.traverseToNamespaceVariable(ctx, nsc, req.GetNamespace(), req.GetKey(), true)
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespaceVariableResponse

	resp.Namespace = d.ns().Name
	resp.Key = d.vref.Name
	resp.CreatedAt = timestamppb.New(d.vdata.CreatedAt)
	resp.UpdatedAt = timestamppb.New(d.vdata.UpdatedAt)
	resp.Checksum = d.vdata.Hash
	resp.TotalSize = int64(d.vdata.Size)
	resp.MimeType = d.vdata.MimeType

	if resp.TotalSize > parcelSize {
		return nil, status.Error(codes.ResourceExhausted, "variable too large to return without using the parcelling API")
	}

	resp.Data = d.vdata.Data

	return &resp, nil

}

func (internal *internal) NamespaceVariableParcels(req *grpc.VariableInternalRequest, srv grpc.Internal_NamespaceVariableParcelsServer) error {

	internal.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()

	nsc := internal.db.Namespace
	inc := internal.db.Instance

	id, err := internal.getInstance(ctx, inc, req.GetInstance(), false)
	if err != nil {
		return err
	}

	d, err := internal.traverseToNamespaceVariable(ctx, nsc, id.namespace(), req.GetKey(), true)
	if err != nil && !derrors.IsNotFound(err) {
		return err
	}

	if derrors.IsNotFound(err) {
		d = new(nsvarData)
		d.vref = new(ent.VarRef)
		d.vref.Name = req.GetKey()
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
		resp.MimeType = d.vdata.MimeType

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

	nsc := flow.db.Namespace

	d, err := flow.traverseToNamespaceVariable(ctx, nsc, req.GetNamespace(), req.GetKey(), true)
	if err != nil {
		return err
	}

	rdr := bytes.NewReader(d.vdata.Data)

	for {

		resp := new(grpc.NamespaceVariableResponse)

		resp.Namespace = d.ns().Name
		resp.Key = d.vref.Name
		resp.CreatedAt = timestamppb.New(d.vdata.CreatedAt)
		resp.UpdatedAt = timestamppb.New(d.vdata.UpdatedAt)
		resp.Checksum = d.vdata.Hash
		resp.TotalSize = int64(d.vdata.Size)
		resp.MimeType = d.vdata.MimeType

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

func variablesOrder(p *pagination) []ent.VarRefPaginateOption {

	var opts []ent.VarRefPaginateOption

	for _, o := range p.order {

		if o == nil {
			continue
		}

		field := ent.VarRefOrderFieldName
		direction := ent.OrderDirectionAsc

		if x := o.Field; x != "" && x == "NAME" {
			field = ent.VarRefOrderFieldName
		}

		if x := o.Direction; x != "" && x == "DESC" {
			direction = ent.OrderDirectionDesc
		}

		opts = append(opts, ent.WithVarRefOrder(&ent.VarRefOrder{
			Direction: direction,
			Field:     field,
		}))
	}

	if len(opts) == 0 {
		opts = append(opts, ent.WithVarRefOrder(&ent.VarRefOrder{
			Direction: ent.OrderDirectionAsc,
			Field:     ent.VarRefOrderFieldName,
		}))
	}

	return opts

}

func variablesFilter(p *pagination) []ent.VarRefPaginateOption {

	var filters []func(query *ent.VarRefQuery) (*ent.VarRefQuery, error)
	var opts []ent.VarRefPaginateOption

	if p.filter == nil {
		return nil
	}

	for i := range p.filter {

		f := p.filter[i]

		if f == nil {
			continue
		}

		filter := f.Val

		filters = append(filters, func(query *ent.VarRefQuery) (*ent.VarRefQuery, error) {

			if filter == "" {
				return query, nil
			}

			field := f.Field
			if field == "" {
				return query, nil
			}

			switch field {
			case "NAME":

				ftype := f.Type
				if ftype == "" {
					return query, nil
				}

				switch ftype {
				case "CONTAINS":
					return query.Where(entvar.NameContains(filter)), nil
				}
			}

			return query, nil

		})

	}

	if len(filters) > 0 {
		opts = append(opts, ent.WithVarRefFilter(func(query *ent.VarRefQuery) (*ent.VarRefQuery, error) {
			var err error
			for _, filter := range filters {
				query, err = filter(query)
				if err != nil {
					return nil, err
				}
			}
			return query, nil
		}))
	}

	return opts

}

func (flow *flow) NamespaceVariables(ctx context.Context, req *grpc.NamespaceVariablesRequest) (*grpc.NamespaceVariablesResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	p, err := getPagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	opts := []ent.VarRefPaginateOption{}
	opts = append(opts, variablesOrder(p)...)
	opts = append(opts, variablesFilter(p)...)

	nsc := flow.db.Namespace
	ns, err := flow.getNamespace(ctx, nsc, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	query := ns.QueryVars()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return nil, err
	}

	var resp grpc.NamespaceVariablesResponse

	resp.Namespace = ns.Name

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
		v.MimeType = vdata.MimeType

	}

	return &resp, nil

}

func (flow *flow) NamespaceVariablesStream(req *grpc.NamespaceVariablesRequest, srv grpc.Flow_NamespaceVariablesStreamServer) error {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	p, err := getPagination(req.Pagination)
	if err != nil {
		return err
	}

	opts := []ent.VarRefPaginateOption{}
	opts = append(opts, variablesOrder(p)...)
	opts = append(opts, variablesFilter(p)...)

	nsc := flow.db.Namespace
	ns, err := flow.getNamespace(ctx, nsc, req.GetNamespace())
	if err != nil {
		return err
	}

	sub := flow.pubsub.SubscribeNamespaceVariables(ns)
	defer flow.cleanup(sub.Close)

resend:

	query := ns.QueryVars()
	cx, err := query.Paginate(ctx, p.After(), p.First(), p.Before(), p.Last(), opts...)
	if err != nil {
		return err
	}

	resp := new(grpc.NamespaceVariablesResponse)

	resp.Namespace = ns.Name

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

func (flow *flow) SetNamespaceVariable(ctx context.Context, req *grpc.SetNamespaceVariableRequest) (*grpc.SetNamespaceVariableResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	vrefc := tx.VarRef
	vdatac := tx.VarData

	ns, err := flow.getNamespace(ctx, nsc, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	var vdata *ent.VarData

	key := req.GetKey()

	var newVar bool
	vdata, newVar, err = flow.SetVariable(ctx, vrefc, vdatac, ns, key, req.GetData(), req.GetMimeType(), false)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	if newVar {
		flow.logToNamespace(ctx, time.Now(), ns, "Created namespace variable '%s'.", key)
	} else {
		flow.logToNamespace(ctx, time.Now(), ns, "Updated namespace variable '%s'.", key)
	}

	flow.pubsub.NotifyNamespaceVariables(ns)

	var resp grpc.SetNamespaceVariableResponse

	resp.Namespace = ns.Name
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

	inc := internal.db.Instance

	id, err := internal.getInstance(ctx, inc, req.GetInstance(), false)
	if err != nil {
		return err
	}

	mimeType := req.GetMimeType()
	namespace := id.namespace()
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

	nsc := tx.Namespace
	vrefc := tx.VarRef
	vdatac := tx.VarData

	ns, err := internal.getNamespace(ctx, nsc, namespace)
	if err != nil {
		return err
	}

	var vdata *ent.VarData

	var newVar bool
	vdata, newVar, err = internal.flow.SetVariable(ctx, vrefc, vdatac, ns, key, buf.Bytes(), mimeType, false)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newVar {
		internal.logToNamespace(ctx, time.Now(), ns, "Created namespace variable '%s'.", key)
	} else {
		internal.logToNamespace(ctx, time.Now(), ns, "Updated namespace variable '%s'.", key)
	}

	internal.pubsub.NotifyNamespaceVariables(ns)

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

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	vrefc := tx.VarRef
	vdatac := tx.VarData

	ns, err := flow.getNamespace(ctx, nsc, namespace)
	if err != nil {
		return err
	}

	var vdata *ent.VarData

	var newVar bool
	vdata, newVar, err = flow.SetVariable(ctx, vrefc, vdatac, ns, key, buf.Bytes(), mimeType, false)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if newVar {
		flow.logToNamespace(ctx, time.Now(), ns, "Created namespace variable '%s'.", key)
	} else {
		flow.logToNamespace(ctx, time.Now(), ns, "Updated namespace variable '%s'.", key)
	}

	flow.pubsub.NotifyNamespaceVariables(ns)

	var resp grpc.SetNamespaceVariableResponse

	resp.Namespace = ns.Name
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

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace

	d, err := flow.traverseToNamespaceVariable(ctx, nsc, req.GetNamespace(), req.GetKey(), false)
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

	flow.logToNamespace(ctx, time.Now(), d.ns(), "Deleted namespace variable '%s'.", d.vref.Name)
	flow.pubsub.NotifyNamespaceVariables(d.ns())

	// Broadcast Event
	broadcastInput := broadcastVariableInput{
		WorkflowPath: "",
		Key:          req.GetKey(),
		TotalSize:    int64(d.vdata.Size),
		Scope:        BroadcastEventScopeNamespace,
	}
	err = flow.BroadcastVariable(ctx, BroadcastEventTypeDelete, BroadcastEventScopeNamespace, broadcastInput, d.ns())
	if err != nil {
		return nil, err
	}

	var resp emptypb.Empty

	return &resp, nil

}

func (flow *flow) RenameNamespaceVariable(ctx context.Context, req *grpc.RenameNamespaceVariableRequest) (*grpc.RenameNamespaceVariableResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	tx, err := flow.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	nsc := tx.Namespace
	d, err := flow.traverseToNamespaceVariable(ctx, nsc, req.GetNamespace(), req.GetOld(), false)
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

	flow.logToNamespace(ctx, time.Now(), d.ns(), "Renamed namespace variable from '%s' to '%s'.", req.GetOld(), req.GetNew())
	flow.pubsub.NotifyNamespaceVariables(d.ns())

	var resp grpc.RenameNamespaceVariableResponse

	resp.Checksum = d.vdata.Hash
	resp.CreatedAt = timestamppb.New(d.vdata.CreatedAt)
	resp.Key = vref.Name
	resp.Namespace = d.ns().Name
	resp.TotalSize = int64(d.vdata.Size)
	resp.UpdatedAt = timestamppb.New(d.vdata.UpdatedAt)
	resp.MimeType = d.vdata.MimeType

	return &resp, nil

}
