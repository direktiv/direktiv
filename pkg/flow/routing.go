package flow

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/bytedata"
	"github.com/direktiv/direktiv/pkg/flow/database"
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

func (srv *server) validateRouter(ctx context.Context, tx *sqlTx, file *filestore.File) (*muxStart, error) {
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

func (engine *engine) mux(ctx context.Context, ns *database.Namespace, calledAs string) (*filestore.File, []byte, error) {
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
		flow.sugar.Error(err)
		return
	}

	if msg.Cron == "" {
		flow.timers.deleteCronForWorkflow(msg.ID)
	}

	if msg.Cron != "" {
		err = flow.timers.addCron(msg.ID, wfCron, msg.Cron, []byte(msg.ID))
		if err != nil {
			flow.sugar.Error(err)
			return
		}
	}
}

func (flow *flow) cronHandler(data []byte) {
	ctx := context.Background()

	id, err := uuid.Parse(string(data))
	if err != nil {
		flow.sugar.Error(err)
		return
	}

	// tx is to be committed in the NewInstance call.
	tx, err := flow.beginSqlTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		flow.sugar.Error(err)
		return
	}
	defer tx.Rollback()

	file, err := tx.FileStore().GetFileByID(ctx, id)
	if err != nil {
		if errors.Is(err, filestore.ErrNotFound) {
			flow.sugar.Infof("Cron failed to find workflow. Deleting cron.")
			flow.timers.deleteCronForWorkflow(id.String())
			return
		}
		flow.sugar.Error(err)
		return
	}

	root, err := tx.FileStore().GetRoot(ctx, file.RootID)
	if err != nil {
		flow.sugar.Error(err)
		return
	}

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, root.Namespace)
	if err != nil {
		flow.sugar.Error(err)
		return
	}

	err = tx.InstanceStore().AssertNoParallelCron(ctx, file.Path)
	if errors.Is(err, instancestore.ErrParallelCron) {
		// already triggered
		return
	} else if err != nil {
		flow.sugar.Error(err)
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
			// return without logging because this attempt to create an instance clashed with another server
			return
		}

		flow.sugar.Errorf("Error returned to gRPC request %s: %v", this(), err)

		return
	}

	go flow.engine.start(im)
}

func (flow *flow) configureWorkflowStarts(ctx context.Context, tx *sqlTx, nsID uuid.UUID, file *filestore.File) error {
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
