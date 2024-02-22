package flow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (srv *server) getInstance(ctx context.Context, namespace, instanceID string) (*enginerefactor.Instance, error) {
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return nil, err
	}

	tx, err := srv.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, namespace)
	if err != nil {
		return nil, err
	}

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

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	idata, err := tx.InstanceStore().ForInstanceID(instID).GetSummaryWithInput(ctx)
	if err != nil {
		return nil, err
	}

	tx.Rollback()

	instance, err := enginerefactor.ParseInstanceData(idata)
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceInputResponse
	resp.Instance = bytedata.ConvertInstanceToGrpcInstance(instance)
	resp.Data = idata.Input
	resp.Namespace = ns.Name

	return &resp, nil
}

func (flow *flow) InstanceOutput(ctx context.Context, req *grpc.InstanceOutputRequest) (*grpc.InstanceOutputResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	instID, err := uuid.Parse(req.GetInstance())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	idata, err := tx.InstanceStore().ForInstanceID(instID).GetSummaryWithOutput(ctx)
	if err != nil {
		return nil, err
	}

	tx.Rollback()

	instance, err := enginerefactor.ParseInstanceData(idata)
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceOutputResponse
	resp.Instance = bytedata.ConvertInstanceToGrpcInstance(instance)
	resp.Data = idata.Output
	resp.Namespace = ns.Name

	return &resp, nil
}

func (flow *flow) InstanceMetadata(ctx context.Context, req *grpc.InstanceMetadataRequest) (*grpc.InstanceMetadataResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	instID, err := uuid.Parse(req.GetInstance())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

	idata, err := tx.InstanceStore().ForInstanceID(instID).GetSummaryWithMetadata(ctx)
	if err != nil {
		return nil, err
	}

	tx.Rollback()

	instance, err := enginerefactor.ParseInstanceData(idata)
	if err != nil {
		return nil, err
	}

	var resp grpc.InstanceMetadataResponse
	resp.Instance = bytedata.ConvertInstanceToGrpcInstance(instance)
	resp.Data = idata.Metadata
	resp.Namespace = ns.Name

	return &resp, nil
}

func (flow *flow) Instances(ctx context.Context, req *grpc.InstancesRequest) (*grpc.InstancesResponse, error) {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	opts := new(instancestore.ListOpts)
	if req.Pagination != nil {
		opts.Limit = int(req.Pagination.Limit)
		opts.Offset = int(req.Pagination.Offset)

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

			switch x.Type {
			case "CONTAINS":
				filter.Kind = instancestore.FilterKindContains
			case "WORKFLOW":
				filter.Kind = instancestore.FilterKindMatch
			case "PREFIX":
				filter.Kind = instancestore.FilterKindPrefix
			case "MATCH": //nolint:goconst
				filter.Kind = instancestore.FilterKindMatch
			case "AFTER":
				filter.Kind = instancestore.FilterKindAfter
			case "BEFORE":
				filter.Kind = instancestore.FilterKindBefore
			default:
				filter.Kind = x.Type
			}

			switch x.Field {
			case "AS":
				filter.Field = instancestore.FieldWorkflowPath
				filter.Value = x.Val
			case "CREATED":
				filter.Field = instancestore.FieldCreatedAt
				t, err := time.Parse(time.RFC3339, x.Val)
				if err != nil {
					return nil, instancestore.ErrBadListOpts
				}
				filter.Value = t.UTC()
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

			opts.Filters = append(opts.Filters, filter)
		}
	}

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

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

	resp.Instances.Results = bytedata.ConvertInstancesToGrpcInstances(results.Results)

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

	tx, err := flow.beginSqlTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
	if err != nil {
		return nil, err
	}

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
	resp.Instance = bytedata.ConvertInstanceToGrpcInstance(instance)
	resp.Flow = instance.RuntimeInfo.Flow

	if l := len(instance.DescentInfo.Descent); l > 0 {
		resp.InvokedBy = instance.DescentInfo.Descent[l-1].ID.String()
	}

	resp.Namespace = instance.TelemetryInfo.NamespaceName

	rwf := new(grpc.InstanceWorkflow)
	rwf.Name = filepath.Base(instance.Instance.WorkflowPath)
	rwf.Parent = filepath.Dir(instance.Instance.WorkflowPath)
	rwf.Path = instance.Instance.WorkflowPath
	resp.Workflow = rwf

	return &resp, nil
}

func (flow *flow) InstanceStream(req *grpc.InstanceRequest, srv grpc.Flow_InstanceStreamServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	phash := ""
	nhash := ""

	var err error
	var ns *database.Namespace
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
		return err
	})
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
	resp.Instance = bytedata.ConvertInstanceToGrpcInstance(instance)
	resp.Flow = instance.RuntimeInfo.Flow

	if l := len(instance.DescentInfo.Descent); l > 0 {
		resp.InvokedBy = instance.DescentInfo.Descent[l-1].ID.String()
	}

	resp.Namespace = ns.Name

	rwf := new(grpc.InstanceWorkflow)
	rwf.Name = filepath.Base(instance.Instance.WorkflowPath)
	rwf.Parent = filepath.Dir(instance.Instance.WorkflowPath)
	rwf.Path = instance.Instance.WorkflowPath
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

	var err error
	var ns *database.Namespace
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
		return err
	})
	if err != nil {
		return nil, err
	}

	calledAs := req.GetPath()

	span := trace.SpanFromContext(ctx)

	input := req.GetInput()
	if input == nil {
		input = make([]byte, 0)
	}

	args := &newInstanceArgs{
		ID:        uuid.New(),
		Namespace: ns,
		CalledAs:  calledAs,
		Input:     input,
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

type grpcMetadataTMC struct {
	md *metadata.MD
}

func (tmc *grpcMetadataTMC) Get(k string) string {
	array := tmc.md.Get(k)
	if len(array) == 0 {
		return ""
	}
	return array[0]
}

func (tmc *grpcMetadataTMC) Keys() []string {
	keys := tmc.md.Get("oteltmckeys")
	if keys == nil {
		keys = make([]string, 0)
	}
	return keys
}

func (tmc *grpcMetadataTMC) Set(k, v string) {
	newKey := len(tmc.md.Get(k)) == 0
	tmc.md.Set(k, v)
	if newKey {
		tmc.md.Append("oteltmckeys", k)
	}
}

func (flow *flow) AwaitWorkflow(req *grpc.AwaitWorkflowRequest, srv grpc.Flow_AwaitWorkflowServer) error {
	flow.sugar.Debugf("Handling gRPC request: %s", this())

	ctx := srv.Context()
	prop := otel.GetTextMapPropagator()
	requestMetadata, _ := metadata.FromIncomingContext(ctx)
	metadataCopy := requestMetadata.Copy()
	carrier := &grpcMetadataTMC{&metadataCopy}
	ctx = prop.Extract(ctx, carrier)

	phash := ""
	nhash := ""

	var err error
	var ns *database.Namespace
	err = flow.runSqlTx(ctx, func(tx *sqlTx) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, req.GetNamespace())
		return err
	})
	if err != nil {
		return err
	}

	calledAs := req.GetPath()

	span := trace.SpanFromContext(ctx)
	cctx, cSpan := span.TracerProvider().Tracer("flow/direktiv").Start(ctx, "working")
	defer cSpan.End()
	span = trace.SpanFromContext(cctx)

	input := req.GetInput()
	if input == nil {
		input = make([]byte, 0)
	}

	args := &newInstanceArgs{
		ID:        uuid.New(),
		Namespace: ns,
		CalledAs:  calledAs,
		Input:     input,
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
	rwf.Path = instance.Instance.WorkflowPath
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

		resp.Data = idata.Output
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
