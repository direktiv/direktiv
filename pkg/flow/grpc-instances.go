package flow

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"go.opentelemetry.io/otel/trace"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (srv *server) getInstance(ctx context.Context, namespace, instanceID string) (*enginerefactor.Instance, error) {
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, err
	}

	ns, err := srv.edb.NamespaceByName(ctx, namespace)
	if err != nil {
		return nil, err
	}

	tx, err := srv.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	idata, err := tx.InstanceStore().ForInstanceID(id).GetSummary(ctx)
	if err != nil {
		return nil, err
	}

	if ns.ID != idata.NamespaceID {
		return nil, os.ErrNotExist
	}

	instance, err := enginerefactor.ParseInstanceData(idata)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (internal *internal) getInstance(ctx context.Context, instanceID string) (*enginerefactor.Instance, error) {
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, err
	}

	tx, err := internal.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	idata, err := tx.InstanceStore().ForInstanceID(id).GetSummary(ctx)
	if err != nil {
		return nil, err
	}

	instance, err := enginerefactor.ParseInstanceData(idata)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (flow *flow) InstanceInput(ctx context.Context, req *grpc.InstanceInputRequest) (*grpc.InstanceInputResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	instID, err := uuid.Parse(req.GetInstance())
	if err != nil {
		return nil, err
	}

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	idata, err := tx.InstanceStore().ForInstanceID(instID).GetSummaryWithInput(ctx)
	if err != nil {
		return nil, err
	}

	tx.Rollback()

	var resp grpc.InstanceInputResponse

	err = bytedata.ConvertDataForOutput(idata, &resp.Instance)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(idata.Input, &m)
	if err != nil {
		return nil, err
	}
	delete(m, "private")
	input := bytedata.Marshal(m)

	resp.Data = []byte(input)
	resp.Namespace = ns.Name

	return &resp, nil
}

func (flow *flow) InstanceOutput(ctx context.Context, req *grpc.InstanceOutputRequest) (*grpc.InstanceOutputResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	instID, err := uuid.Parse(req.GetInstance())
	if err != nil {
		return nil, err
	}

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	idata, err := tx.InstanceStore().ForInstanceID(instID).GetSummaryWithOutput(ctx)
	if err != nil {
		return nil, err
	}

	tx.Rollback()

	var resp grpc.InstanceOutputResponse

	err = bytedata.ConvertDataForOutput(idata, &resp.Instance)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal([]byte(idata.Output), &m)
	if err != nil {
		return nil, err
	}
	delete(m, "private")
	output := bytedata.Marshal(m)

	resp.Data = []byte(output)
	resp.Namespace = ns.Name

	return &resp, nil
}

func (flow *flow) InstanceMetadata(ctx context.Context, req *grpc.InstanceMetadataRequest) (*grpc.InstanceMetadataResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	instID, err := uuid.Parse(req.GetInstance())
	if err != nil {
		return nil, err
	}

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	idata, err := tx.InstanceStore().ForInstanceID(instID).GetSummaryWithMetadata(ctx)
	if err != nil {
		return nil, err
	}

	tx.Rollback()

	var resp grpc.InstanceMetadataResponse

	err = bytedata.ConvertDataForOutput(idata, &resp.Instance)
	if err != nil {
		return nil, err
	}

	resp.Data = []byte(idata.Metadata)
	resp.Namespace = ns.Name

	return &resp, nil
}

func (flow *flow) Instances(ctx context.Context, req *grpc.InstancesRequest) (*grpc.InstancesResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	opts := new(instancestore.ListOpts)
	if req.Pagination != nil {
		opts.Limit = int(req.Pagination.Limit)
		opts.Limit = int(req.Pagination.Offset)

		for idx := range req.Pagination.Order {
			x := req.Pagination.Order[idx]
			var order instancestore.Order
			switch x.Direction {
			case "":
				fallthrough
			case "DESC":
				order.Descending = true
			case "ASC":
			default:
				return nil, instancestore.ErrBadListOpts
			}

			switch x.Field {
			case "CREATED":
				order.Field = instancestore.FieldCreatedAt
			default:
				order.Field = x.Field
			}

			opts.Orders = append(opts.Orders, order)
		}

		var err error

		for idx := range req.Pagination.Filter {
			x := req.Pagination.Filter[idx]
			var filter instancestore.Filter
			filter.Kind = x.Type

			switch x.Field {
			case "AS":
				filter.Field = instancestore.FieldCalledAs
				filter.Value = x.Val
			case "CREATED":
				filter.Field = instancestore.FieldCreatedAt
				t, err := time.Parse(time.RFC3339, x.Val)
				if err != nil {
					return nil, instancestore.ErrBadListOpts
				}
				filter.Value = t
			case "STATUS":
				filter.Field = instancestore.FieldStatus
				filter.Value, err = instancestore.InstanceStatusFromString(x.Val)
				if err != nil {
					return nil, instancestore.ErrBadListOpts
				}
			case "TRIGGER":
				filter.Field = instancestore.FieldInvoker
				filter.Value = x.Val
			default:
				filter.Field = x.Field
				filter.Value = x.Val
			}
		}
	}

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	results, err := tx.InstanceStore().GetNamespaceInstances(ctx, ns.ID, opts)
	if err != nil {
		return nil, err
	}

	tx.Rollback()

	resp := new(grpc.InstancesResponse)
	resp.Namespace = ns.Name
	resp.Instances = new(grpc.Instances)
	resp.Instances.PageInfo = &grpc.PageInfo{
		Total: int32(results.Total),
		// Limit: ,
		// Offset: ,
		// Order: ,
		// Filter: ,
	}

	err = bytedata.ConvertDataForOutput(results, &resp.Instances.Results)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (flow *flow) InstancesStream(req *grpc.InstancesRequest, srv grpc.Flow_InstancesStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())
	ctx := srv.Context()

	resp, err := flow.Instances(ctx, req)
	if err != nil {
		return err
	}
	// mock streaming response.
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			err = srv.Send(resp)
			if err != nil {
				return err
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func (flow *flow) Instance(ctx context.Context, req *grpc.InstanceRequest) (*grpc.InstanceResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	instID, err := uuid.Parse(req.GetInstance())
	if err != nil {
		return nil, err
	}

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	idata, err := tx.InstanceStore().ForInstanceID(instID).GetSummary(ctx)
	if err != nil {
		return nil, err
	}

	tx.Rollback()

	if ns.ID != idata.NamespaceID {
		return nil, instancestore.ErrNotFound
	}

	instance, err := enginerefactor.ParseInstanceData(idata)
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceResponse

	err = bytedata.ConvertDataForOutput(idata, &resp.Instance)
	if err != nil {
		return nil, err
	}

	resp.Flow = instance.RuntimeInfo.Flow
	// TODO: alan
	// if rt.Caller != uuid.Nil {
	// 	resp.InvokedBy = rt.Caller.String()
	// }

	resp.Namespace = instance.TelemetryInfo.NamespaceName

	rwf := new(grpc.InstanceWorkflow)
	// TODO: alan
	// rwf.Name = instance.Instance.CalledAs
	// rwf.Parent = strings.TrimPrefix(cached.File.Dir(), "/") // TODO: get rid of the trim?
	// rwf.Path = strings.TrimPrefix(cached.File.Path, "/")    // TODO: get rid of the trim?
	rwf.Revision = instance.Instance.RevisionID.String()
	resp.Workflow = rwf

	return &resp, nil
}

func (flow *flow) InstanceStream(req *grpc.InstanceRequest, srv grpc.Flow_InstanceStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return err
	}

	instID, err := uuid.Parse(req.GetInstance())
	if err != nil {
		return err
	}

	var sub *pubsub.Subscription

resend:

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	idata, err := tx.InstanceStore().ForInstanceID(instID).GetSummary(ctx)
	if err != nil {
		return err
	}

	tx.Rollback()

	if sub == nil {
		sub = flow.pubsub.SubscribeInstance(idata.ID)
		defer flow.cleanup(sub.Close)
		goto resend
	}

	instance, err := enginerefactor.ParseInstanceData(idata)
	if err != nil {
		return err
	}

	resp := new(grpc.InstanceResponse)

	err = bytedata.ConvertDataForOutput(instance, &resp.Instance)
	if err != nil {
		return err
	}

	resp.Flow = instance.RuntimeInfo.Flow
	// TODO: alan
	// resp.InvokedBy = rt.Caller.String()

	resp.Namespace = ns.Name

	rwf := new(grpc.InstanceWorkflow)
	// TODO: alan
	// if cached.File != nil {
	// 	rwf.Name = cached.File.Name()
	// 	rwf.Parent = strings.TrimPrefix(cached.File.Dir(), "/") // TODO: get rid of the trim?
	// 	rwf.Path = strings.TrimPrefix(cached.File.Path, "/")    // TODO: get rid of the trim?
	// }
	rwf.Revision = instance.Instance.RevisionID.String()
	resp.Workflow = rwf

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

func (flow *flow) StartWorkflow(ctx context.Context, req *grpc.StartWorkflowRequest) (*grpc.StartWorkflowResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	calledAs := req.GetPath()
	if req.GetRef() != "" {
		calledAs += ":" + req.GetRef()
	}

	span := trace.SpanFromContext(ctx)

	args := &newInstanceArgs{
		ID:        uuid.New(),
		Namespace: ns,
		CalledAs:  calledAs,
		Input:     req.GetInput(),
		Invoker:   apiCaller,
		TelemetryInfo: &enginerefactor.InstanceTelemetryInfo{
			TraceID: span.SpanContext().TraceID().String(),
			SpanID:  span.SpanContext().SpanID().String(),
			// TODO: alan, CallPath: ,
			NamespaceName: ns.Name,
		},
	}

	im, err := flow.engine.NewInstance(ctx, args)
	if err != nil {
		flow.sugar.Debugf("Error returned to gRPC request %s: %v", this(), err)
		flow.logger.Errorf(ctx, flow.ID, flow.GetAttributes(), "Failed starting a Workflow")
		return nil, err
	}

	if !req.GetHold() {
		flow.engine.queue(im)
	}

	var resp grpc.StartWorkflowResponse

	resp.Namespace = req.GetNamespace()
	resp.Instance = im.ID().String()

	return &resp, nil
}

func (flow *flow) ReleaseInstance(ctx context.Context, req *grpc.ReleaseInstanceRequest) (*grpc.ReleaseInstanceResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	im, err := flow.engine.getInstanceMemory(ctx, req.GetInstance())
	if err != nil {
		return nil, err
	}

	if im.instance.TelemetryInfo.NamespaceName != req.GetNamespace() {
		return nil, errors.New("instance not found")
	}

	if im.instance.Instance.Status != instancestore.InstanceStatusPending {
		return nil, errors.New("instance already released")
	}

	flow.engine.queue(im)

	var resp grpc.ReleaseInstanceResponse

	resp.Namespace = req.GetNamespace()
	resp.Instance = im.ID().String()

	return &resp, nil
}

func (flow *flow) CancelInstance(ctx context.Context, req *grpc.CancelInstanceRequest) (*emptypb.Empty, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	instance, err := flow.getInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		flow.logger.Errorf(ctx, flow.ID, flow.GetAttributes(), "Failed to resolve instance %s", req.GetInstance())
		return nil, err
	}

	flow.engine.cancelInstance(instance.Instance.ID.String(), "direktiv.cancels.api", "cancelled by api request", false)

	var resp emptypb.Empty

	return &resp, nil
}

func (flow *flow) AwaitWorkflow(req *grpc.AwaitWorkflowRequest, srv grpc.Flow_AwaitWorkflowServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	ns, err := flow.edb.NamespaceByName(ctx, req.GetNamespace())
	if err != nil {
		return err
	}

	calledAs := req.GetPath()
	if req.GetRef() != "" {
		calledAs += ":" + req.GetRef()
	}

	span := trace.SpanFromContext(ctx)

	args := &newInstanceArgs{
		ID:        uuid.New(),
		Namespace: ns,
		CalledAs:  calledAs,
		Input:     req.GetInput(),
		Invoker:   apiCaller,
		TelemetryInfo: &enginerefactor.InstanceTelemetryInfo{
			TraceID: span.SpanContext().TraceID().String(),
			SpanID:  span.SpanContext().SpanID().String(),
			// TODO: alan, CallPath: ,
			NamespaceName: ns.Name,
		},
	}

	im, err := flow.engine.NewInstance(ctx, args)
	if err != nil {
		flow.logger.Errorf(ctx, flow.ID, flow.GetAttributes(), "Failed to create instance: %v", err)
		flow.sugar.Debugf("Error returned to gRPC request %s: %v", this(), err)
		return err
	}

	sub := flow.pubsub.SubscribeInstance(im.instance.Instance.ID)
	defer flow.cleanup(sub.Close)

	flow.engine.queue(im)

	var instance *enginerefactor.Instance

resend:

	instance, err = flow.getInstance(ctx, req.GetNamespace(), im.instance.Instance.ID.String())
	if err != nil {
		flow.sugar.Debugf("Error returned to gRPC request %s: %v", this(), err)
		return err
	}

	resp := new(grpc.AwaitWorkflowResponse)
	resp.Namespace = req.GetNamespace()
	resp.Instance = bytedata.ConvertInstanceToGrpcInstance(instance)
	resp.InvokedBy = instance.Instance.Invoker // TODO: is this accurate?
	resp.Flow = instance.RuntimeInfo.Flow
	resp.Data = instance.Instance.Output
	rwf := new(grpc.InstanceWorkflow)
	// rwf.Name = cached.File.Name()
	// rwf.Parent = cached.Dir()
	rwf.Path = instance.Instance.CalledAs
	rwf.Revision = instance.Instance.RevisionID.String()
	resp.Workflow = rwf

	if instance.Instance.Status == instancestore.InstanceStatusComplete {
		tx, err := flow.beginSqlTx(ctx)
		if err != nil {
			flow.sugar.Debugf("Error returned to gRPC request %s: %v", this(), err)
			return err
		}
		defer tx.Rollback()

		idata, err := tx.InstanceStore().ForInstanceID(instance.Instance.ID).GetSummaryWithOutput(ctx)
		if err != nil {
			flow.sugar.Debugf("Error returned to gRPC request %s: %v", this(), err)
			return err
		}

		tx.Rollback()

		resp.Data = []byte(idata.Output)
	}

	nhash = bytedata.Checksum(resp)
	if nhash != phash {
		err = srv.Send(resp)
		if err != nil {
			flow.sugar.Debugf("Error returned to gRPC request %s: %v", this(), err)
			return err
		}
	}
	phash = nhash

	if instance.Instance.Status != instancestore.InstanceStatusPending {
		return nil
	}

	more := sub.Wait(ctx)
	if !more {
		return nil
	}

	goto resend
}
