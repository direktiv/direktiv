package flow

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	enginerefactor "github.com/direktiv/direktiv/pkg/engine"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/telemetry"
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/google/uuid"
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
		x := def.(*model.EventsAndStart) //nolint:forcetypeassert
		ms.Lifespan = x.LifeSpan
	case model.StartTypeEventsXor:
	case model.StartTypeScheduled:
		x := def.(*model.ScheduledStart) //nolint:forcetypeassert
		ms.Cron = x.Cron
	default:
		panic(fmt.Errorf("unexpected start type: %v", def.GetType()))
	}

	return ms
}

func validateRouter(ctx context.Context, tx *database.DB, file *filestore.File) (*muxStart, error) {
	data, err := tx.FileStore().ForFile(file).GetData(ctx)
	if err != nil {
		return nil, err
	}

	workflow := new(model.Workflow)

	err = workflow.Load(data)
	if err != nil {
		return nil, err
	}

	ms := newMuxStart(workflow)

	return ms, nil
}

func (engine *engine) mux(ctx context.Context, ns *datastore.Namespace, calledAs string) (*filestore.File, []byte, error) {
	uriElems := strings.SplitN(calledAs, ":", 2)
	path := uriElems[0]

	tx, err := engine.flow.beginSQLTx(ctx)
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
		slog.Error("failed to unmarshal router configuration message", "error", err)
		return
	}

	if msg.Cron == "" {
		flow.timers.deleteCronForWorkflow(msg.ID)
	}

	if msg.Cron != "" {
		err = flow.timers.addCron(msg.ID, wfCron, msg.Cron, []byte(msg.ID))
		if err != nil {
			slog.Error("failed to add cron schedule for workflow", "error", err, "cron_expression", msg.Cron)
			return
		}
	}
}

func (flow *flow) cronHandler(data []byte) {
	ctx := context.Background()

	t := time.Now().Truncate(time.Minute).UTC()

	id, err := uuid.Parse(string(data))
	if err != nil {
		slog.Error("failed to parse UUID from cron data", "error", err, "data", string(data))
		return
	}

	// tx is to be committed in the NewInstance call.
	tx, err := flow.beginSQLTx(ctx, &sql.TxOptions{
		// Isolation: sql.LevelSerializable,
	})
	if err != nil {
		slog.Error("failed to begin SQL transaction in cron handler", "error", err)
		return
	}
	defer tx.Rollback()

	file, err := tx.FileStore().GetFileByID(ctx, id)
	if err != nil {
		if errors.Is(err, filestore.ErrNotFound) {
			slog.Info("workflow for cron not found, deleting associated cron entry")
			flow.timers.deleteCronForWorkflow(id.String())

			return
		}
		slog.Error("failed to retrieve file by ID in cron handler", "error", err)

		return
	}

	root, err := tx.FileStore().GetRoot(ctx, file.RootID)
	if err != nil {
		slog.Error("cron getting files", "error", err)

		return
	}

	ns, err := tx.DataStore().Namespaces().GetByName(ctx, root.Namespace)
	if err != nil {
		slog.Error("failed to retrieve namespace in cron handler", "error", err, "namespace", root.Namespace)

		return
	}

	// ctx = tracing.AddNamespace(ctx, ns.Name)
	// ctx, end, err := tracing.NewSpan(ctx, "starting cron handler")
	// if err != nil {
	// 	slog.Debug("cronhandler failed to start span", "error", err)
	// }
	// defer end()

	x, _ := json.Marshal([]string{ns.Name, file.Path, t.String()}) //nolint
	unique := string(x)
	md5sum := md5.Sum([]byte(unique))
	hash := base64.StdEncoding.EncodeToString(md5sum[:])
	traceParent, err := tracing.ExtractTraceParent(ctx)
	if err != nil {
		slog.Debug("cronhandler failed to init telemetry", "error", err)
	}
	args := &newInstanceArgs{
		tx:        tx,
		ID:        uuid.New(),
		Namespace: ns,
		CalledAs:  file.Path,
		Input:     make([]byte, 0),
		Invoker:   "cron",
		TelemetryInfo: &enginerefactor.InstanceTelemetryInfo{
			TraceParent:   traceParent,
			NamespaceName: ns.Name,
		},
		SyncHash: &hash,
	}

	telemetry.LogNamespaceInfo(ctx, fmt.Sprintf("running cron for %s", file.Path), ns.Name)

	im, err := flow.Engine.NewInstance(ctx, args)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			// this happens on a attempt to create an instance clashed with another server
			telemetry.LogNamespaceDebug(ctx, "instance creation clash detected, likely due to parallel execution", ns.Name)
			return
		}

		telemetry.LogNamespaceError(ctx, "failed to create new instance from cron job", ns.Name, fmt.Errorf("cron path error %s, %s", file.Path, err.Error()))

		return
	}

	go flow.Engine.start(im)
}

func (flow *flow) configureWorkflowStarts(ctx context.Context, tx *database.DB, nsID uuid.UUID, nsName string, file *filestore.File) error {
	ms, err := validateRouter(ctx, tx, file)
	if err != nil {
		return err
	}

	err = renderStartEventListener(ctx, nsID, nsName, file, ms, tx)
	if err != nil {
		return err
	}

	flow.pubsub.ConfigureRouterCron(file.ID.String(), ms.Cron)

	return nil
}
