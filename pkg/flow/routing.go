package flow

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/model"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type muxStart struct {
	Type     string                       `json:"type"`
	Cron     string                       `json:"cron"`
	Events   []model.StartEventDefinition `json:"events"`
	Lifespan string                       `json:"lifespan"`
}

func newMuxStart(workflow *model.Workflow) *muxStart {
	ms := new(muxStart)

	def := workflow.GetStartDefinition()
	ms.Type = def.GetType().String()
	ms.Events = def.GetEvents()

	switch def.GetType() {
	case model.StartTypeDefault:
	case model.StartTypeEvent:
	case model.StartTypeEventsAnd:
		x := def.(*model.EventsAndStart)
		ms.Lifespan = x.LifeSpan
	case model.StartTypeEventsXor:
	case model.StartTypeScheduled:
		x := def.(*model.ScheduledStart)
		ms.Cron = x.Cron
	default:
		panic(fmt.Errorf("unexpected start type: %v", def.GetType()))
	}

	return ms
}

func (ms *muxStart) Hash() string {
	if ms == nil {
		ms = new(muxStart)
		ms.Type = model.StartTypeDefault.String()
	}

	return bytedata.Checksum(ms)
}

func (srv *server) validateRouter(ctx context.Context, tx *database.SQLStore, file *filestore.File) (*muxStart, error) {
	data, err := tx.FileStore().ForFile(file).GetData(ctx)
	if err != nil {
		return nil, err
	}

	workflow := new(model.Workflow)

	err = workflow.Load(data)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ms := newMuxStart(workflow)

	return ms, nil
}

func (engine *engine) mux(ctx context.Context, ns *datastore.Namespace, calledAs string) (*filestore.File, []byte, error) {
	// TODO: Alan, fix for the new filestore.(*Revision).GetRevision() api.
	uriElems := strings.SplitN(calledAs, ":", 2)
	path := uriElems[0]

	tx, err := engine.flow.beginSqlTx(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	file, err := tx.FileStore().ForNamespace(ns.Name).GetFile(ctx, path)
	if err != nil {
		return nil, nil, err
	}

	data, err := tx.FileStore().ForFile(file).GetData(ctx)
	if err != nil {
		return nil, nil, err
	}

	return file, data, nil
}

func (flow *flow) configureRouterHandler(req *pubsub.PubsubUpdate) {
	msg := new(pubsub.ConfigureRouterMessage)

	err := json.Unmarshal([]byte(req.Key), msg)
	if err != nil {
		slog.Error("Failed to unmarshal router configuration message.", "error", err)
		return
	}

	if msg.Cron == "" {
		flow.timers.deleteCronForWorkflow(msg.ID)
	}

	if msg.Cron != "" {
		err = flow.timers.addCron(msg.ID, wfCron, msg.Cron, []byte(msg.ID))
		if err != nil {
			slog.Error("Failed to add cron schedule for workflow.", "error", err, "cron_expression", msg.Cron)
			return
		}
	}
}

func (flow *flow) cronHandler(data []byte) {
	ctx := context.Background()

	id, err := uuid.Parse(string(data))
	if err != nil {
		slog.Error("Failed to parse UUID from cron data.", "error", err, "data", string(data))
		return
	}

	// tx is to be committed in the NewInstance call.
	tx, err := flow.beginSqlTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		slog.Error("Failed to begin SQL transaction in cron handler.", "error", err)
		return
	}
	defer tx.Rollback()

	file, err := tx.FileStore().GetFileByID(ctx, id)
	if err != nil {
		if errors.Is(err, filestore.ErrNotFound) {
			slog.Info("Workflow for cron not found, deleting associated cron entry.")
			flow.timers.deleteCronForWorkflow(id.String())
			return
		}
		slog.Error("Failed to retrieve file by ID in cron handler.", "error", err)
		return
	}

	root, err := tx.FileStore().GetRoot(ctx, file.RootID)
	if err != nil {
		slog.Error("cron getting files", "error", err)
		return
	}

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, root.Namespace)
	if err != nil {
		slog.Error("Failed to retrieve namespace in cron handler.", "error", err, "namespace", root.Namespace)
		return
	}

	err = tx.InstanceStore().AssertNoParallelCron(ctx, file.Path)
	if errors.Is(err, instancestore.ErrParallelCron) {
		// already triggered
		return
	} else if err != nil {
		slog.Error("Failed to assert no parallel cron executions.", "error", err, "workflow", file.Path)
		return
	}

	span := trace.SpanFromContext(ctx)

	args := &newInstanceArgs{
		tx:        tx,
		ID:        uuid.New(),
		Namespace: ns,
		CalledAs:  file.Path,
		Input:     make([]byte, 0),
		Invoker:   instancestore.InvokerCron,
		TelemetryInfo: &enginerefactor.InstanceTelemetryInfo{
			TraceID:       span.SpanContext().TraceID().String(),
			SpanID:        span.SpanContext().SpanID().String(),
			NamespaceName: ns.Name,
		},
	}

	im, err := flow.engine.NewInstance(ctx, args)
	if err != nil {
		if strings.Contains(err.Error(), "could not serialize access") {
			slog.Debug("Instance creation clash detected, likely due to parallel execution. Retrying may be required.", "workflow_path", file.Path)
			// this happens on a attempt to create an instance clashed with another server
			return
		}

		slog.Error("Failed to create new instance from cron job.", "workflow_path", file.Path, "error", err)

		return
	}

	go flow.engine.start(im)
}

func (flow *flow) configureWorkflowStarts(ctx context.Context, tx *database.SQLStore, nsID uuid.UUID, file *filestore.File) error {
	ms, err := flow.validateRouter(ctx, tx, file)
	if err != nil {
		return err
	}

	err = flow.events.processWorkflowEvents(ctx, nsID, file, ms)
	if err != nil {
		return err
	}

	flow.pubsub.ConfigureRouterCron(file.ID.String(), ms.Cron)

	return nil
}
