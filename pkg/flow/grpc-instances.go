package flow

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
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

	tx, err := srv.flow.beginSQLTx(ctx)
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

	tx, err := internal.flow.beginSQLTx(ctx)
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

//nolint:dupl
func (flow *flow) InstanceInput(ctx context.Context, req *grpc.InstanceInputRequest) (*grpc.InstanceInputResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	instID, err := uuid.Parse(req.GetInstance())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSQLTx(ctx)
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

//nolint:dupl
func (flow *flow) InstanceOutput(ctx context.Context, req *grpc.InstanceOutputRequest) (*grpc.InstanceOutputResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	instID, err := uuid.Parse(req.GetInstance())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSQLTx(ctx)
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

//nolint:dupl
func (flow *flow) InstanceMetadata(ctx context.Context, req *grpc.InstanceMetadataRequest) (*grpc.InstanceMetadataResponse, error) {
	slog.Debug("Handling gRPC request", "this", this())

	instID, err := uuid.Parse(req.GetInstance())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSQLTx(ctx)
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
	slog.Debug("Handling gRPC request", "this", this())

	opts := new(instancestore.ListOpts)
	if req.GetPagination() != nil {
		opts.Limit = int(req.GetPagination().GetLimit())
		opts.Offset = int(req.GetPagination().GetOffset())

		for idx := range req.GetPagination().GetOrder() {
			x := req.GetPagination().GetOrder()[idx]
			var order instancestore.Order
			switch x.GetDirection() {
			case "":
				fallthrough
			case "DESC":
				order.Descending = true
			case "ASC":
			default:
				return nil, instancestore.ErrBadListOpts
			}

			switch x.GetField() {
			case "CREATED":
				order.Field = instancestore.FieldCreatedAt
			default:
				order.Field = x.GetField()
			}

			opts.Orders = append(opts.Orders, order)
		}

		var err error

		for idx := range req.GetPagination().GetFilter() {
			x := req.GetPagination().GetFilter()[idx]
			var filter instancestore.Filter

			switch x.GetType() {
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
				filter.Kind = x.GetType()
			}

			switch x.GetField() {
			case "AS":
				filter.Field = instancestore.FieldWorkflowPath
				filter.Value = x.GetVal()
			case "CREATED":
				filter.Field = instancestore.FieldCreatedAt
				t, err := time.Parse(time.RFC3339, x.GetVal())
				if err != nil {
					return nil, instancestore.ErrBadListOpts
				}
				filter.Value = t.UTC()
			case "STATUS":
				filter.Field = instancestore.FieldStatus
				filter.Value, err = instancestore.InstanceStatusFromString(x.GetVal())
				if err != nil {
					return nil, instancestore.ErrBadListOpts
				}
			case "TRIGGER":
				filter.Field = instancestore.FieldInvoker
				filter.Value = x.GetVal()
			default:
				filter.Field = x.GetField()
				filter.Value = x.GetVal()
			}

			opts.Filters = append(opts.Filters, filter)
		}
	}

	tx, err := flow.beginSQLTx(ctx)
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
	slog.Debug("Handling gRPC request", "this", this())
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
	slog.Debug("Handling gRPC request", "this", this())

	instID, err := uuid.Parse(req.GetInstance())
	if err != nil {
		return nil, err
	}

	tx, err := flow.beginSQLTx(ctx)
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
	slog.Debug("Handling gRPC request", "this", this())

	ctx := srv.Context()
	var phash, nhash string

	var err error
	var ns *datastore.Namespace
	err = flow.runSQLTx(ctx, func(tx *database.SQLStore) error {
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

	tx, err := flow.beginSQLTx(ctx)
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

func (engine *engine) StartWorkflow(ctx context.Context, namespace, path string, input []byte) (*instancestore.InstanceData, error) {
	var err error
	var ns *datastore.Namespace

	err = engine.runSQLTx(ctx, func(tx *database.SQLStore) error {
		ns, err = tx.DataStore().Namespaces().GetByName(ctx, namespace)
		return err
	})
	if err != nil {
		return nil, err
	}

	calledAs := path

	span := trace.SpanFromContext(ctx)

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
			TraceID:       span.SpanContext().TraceID().String(),
			SpanID:        span.SpanContext().SpanID().String(),
			NamespaceName: ns.Name,
		},
	}

	im, err := engine.NewInstance(ctx, args)
	if err != nil {
		return nil, err
	}

	go engine.start(im) //nolint:contextcheck

	return im.instance.Instance, nil
}

func (flow *flow) StartWorkflow(ctx context.Context, req *grpc.StartWorkflowRequest) (*grpc.StartWorkflowResponse, error) {
	inst, err := flow.engine.StartWorkflow(ctx, req.GetNamespace(), req.GetPath(), req.GetInput())
	if err != nil {
		return nil, err
	}

	var resp grpc.StartWorkflowResponse

	resp.Namespace = req.GetNamespace()
	resp.Instance = inst.ID.String()

	return &resp, nil
}

func (engine *engine) CancelInstance(ctx context.Context, namespace, instanceID string) error {
	instance, err := engine.getInstance(ctx, namespace, instanceID)
	if err != nil {
		return err
	}

	engine.cancelInstance(instance.Instance.ID.String(), "direktiv.cancels.api", "cancelled by api request", false) //nolint:contextcheck

	return nil
}

func (flow *flow) CancelInstance(ctx context.Context, req *grpc.CancelInstanceRequest) (*emptypb.Empty, error) {
	err := flow.engine.CancelInstance(ctx, req.GetNamespace(), req.GetInstance())
	if err != nil {
		slog.Debug("Error returned to gRPC request", "this", this(), "error", err)
		return nil, err
	}

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
	slog.Debug("Handling gRPC request", "this", this())

	ctx := srv.Context()
	prop := otel.GetTextMapPropagator()
	requestMetadata, _ := metadata.FromIncomingContext(ctx)
	metadataCopy := requestMetadata.Copy()
	carrier := &grpcMetadataTMC{&metadataCopy}
	ctx = prop.Extract(ctx, carrier)

	var phash, nhash string

	var err error
	var ns *datastore.Namespace
	err = flow.runSQLTx(ctx, func(tx *database.SQLStore) error {
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
			TraceID:       span.SpanContext().TraceID().String(),
			SpanID:        span.SpanContext().SpanID().String(),
			NamespaceName: ns.Name,
		},
	}

	im, err := flow.engine.NewInstance(ctx, args)
	if err != nil {
		slog.Debug("Error returned to gRPC request", "this", this(), "error", err)

		return err
	}

	sub := flow.pubsub.SubscribeInstance(im.instance.Instance.ID)
	defer flow.cleanup(sub.Close)

	go flow.engine.start(im)

	var instance *enginerefactor.Instance

resend:

	instance, err = flow.getInstance(ctx, req.GetNamespace(), im.instance.Instance.ID.String())
	if err != nil {
		slog.Debug("Error returned to gRPC request", "this", this(), "error", err)

		return err
	}

	resp := new(grpc.AwaitWorkflowResponse)
	resp.Namespace = req.GetNamespace()
	resp.Instance = bytedata.ConvertInstanceToGrpcInstance(instance)
	resp.InvokedBy = instance.Instance.Invoker // TODO: is this accurate?
	resp.Flow = instance.RuntimeInfo.Flow
	resp.Data = instance.Instance.Output
	rwf := new(grpc.InstanceWorkflow)
	rwf.Path = instance.Instance.WorkflowPath
	resp.Workflow = rwf

	if instance.Instance.Status == instancestore.InstanceStatusComplete {
		tx, err := flow.beginSQLTx(ctx)
		if err != nil {
			slog.Debug("Error returned to gRPC request", "this", this(), "error", err)

			return err
		}
		defer tx.Rollback()

		idata, err := tx.InstanceStore().ForInstanceID(instance.Instance.ID).GetSummaryWithOutput(ctx)
		if err != nil {
			slog.Debug("Error returned to gRPC request", "this", this(), "error", err)

			return err
		}

		tx.Rollback()

		resp.Data = idata.Output
	}

	nhash = bytedata.Checksum(resp)
	if nhash != phash {
		err = srv.Send(resp)
		if err != nil {
			slog.Debug("Error returned to gRPC request", "this", this(), "error", err)

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
