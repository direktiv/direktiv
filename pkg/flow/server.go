package flow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	eventsstore "github.com/direktiv/direktiv/pkg/events"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/flow/nohome"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/mirror"
	pubsub2 "github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

type server struct {
	ID uuid.UUID

	config *core.Config

	// db       *ent.Client
	pubsub *pubsub.Pubsub

	// the new pubsub bus
	Bus *pubsub2.Bus

	timers *timers
	Engine *engine

	rawDB *sql.DB

	db *database.DB

	MirrorManager *mirror.Manager

	flow   *flow
	events *events
	nats   *nats.Conn

	ConfigureWorkflow func(event *pubsub2.FileSystemChangeEvent) error
}

type mirrorProcessLogger struct{}

func (log *mirrorProcessLogger) Debug(pid uuid.UUID, msg string, kv ...interface{}) {
	slog.Debug(fmt.Sprintf(msg, kv...), "activity", pid, string(core.LogTrackKey), "activity."+pid.String())
}

func (log *mirrorProcessLogger) Info(pid uuid.UUID, msg string, kv ...interface{}) {
	slog.Info(fmt.Sprintf(msg, kv...), "activity", pid, string(core.LogTrackKey), "activity."+pid.String())
}

func (log *mirrorProcessLogger) Warn(pid uuid.UUID, msg string, kv ...interface{}) {
	slog.Warn(fmt.Sprintf(msg, kv...), "activity", pid, string(core.LogTrackKey), "activity"+"."+pid.String())
}

func (log *mirrorProcessLogger) Error(pid uuid.UUID, msg string, kv ...interface{}) {
	slog.Error(fmt.Sprintf(msg, kv...), "activity", pid, string(core.LogTrackKey), "activity"+"."+pid.String())
}

var _ mirror.ProcessLogger = &mirrorProcessLogger{}

type mirrorCallbacks struct {
	logger   mirror.ProcessLogger
	store    datastore.MirrorStore
	fstore   filestore.FileStore
	varstore datastore.RuntimeVariablesStore
	wfconf   func(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File) error
}

func (c *mirrorCallbacks) ConfigureWorkflowFunc(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File) error {
	return c.wfconf(ctx, nsID, nsName, file)
}

func (c *mirrorCallbacks) ProcessLogger() mirror.ProcessLogger {
	return c.logger
}

func (c *mirrorCallbacks) Store() datastore.MirrorStore {
	return c.store
}

func (c *mirrorCallbacks) FileStore() filestore.FileStore {
	return c.fstore
}

func (c *mirrorCallbacks) VarStore() datastore.RuntimeVariablesStore {
	return c.varstore
}

var _ mirror.Callbacks = &mirrorCallbacks{}

//nolint:revive
func InitLegacyServer(circuit *core.Circuit, config *core.Config, bus *pubsub2.Bus, db *database.DB, rawDB *sql.DB) (*server, error) {
	srv := new(server)
	srv.ID = uuid.New()
	srv.initJQ()
	srv.config = config
	srv.db = db
	srv.rawDB = rawDB
	srv.Bus = bus

	var err error
	slog.Debug("starting Flow server")
	slog.Debug("initializing telemetry.")
	telEnd, err := tracing.InitTelemetry(circuit.Context(), srv.config.OpenTelemetry, "direktiv/flow", "direktiv")
	if err != nil {
		return nil, fmt.Errorf("telemetry init failed: %w", err)
	}
	slog.Info("telemetry initialized successfully.")

	slog.Debug("initializing pub-sub.")

	srv.pubsub, err = pubsub.InitPubSub(srv, config.DB)
	if err != nil {
		return nil, err
	}

	slog.Info("pub-sub was initialized successfully.")

	slog.Debug("initializing timers.")

	srv.timers, err = initTimers(srv.pubsub)
	if err != nil {
		return nil, err
	}
	slog.Info("timers where initialized successfully.")

	slog.Debug("initializing engine.")

	srv.Engine = initEngine(srv)
	slog.Info("engine was started.")

	slog.Debug("initializing flow server.")

	srv.flow, err = initFlowServer(circuit.Context(), srv)
	if err != nil {
		return nil, err
	}

	slog.Debug("initializing mirror manager.")
	slog.Debug("mirror manager was started.")

	slog.Debug("initializing events.")
	srv.events = initEvents(srv, db.DataStore().StagingEvents().Append)

	slog.Debug("initializing EventWorkers.")

	interval := 1 * time.Second // TODO: Adjust the polling interval
	eventWorker := eventsstore.NewEventWorker(db.DataStore().StagingEvents(), interval, srv.events.handleEvent)

	circuit.Start(func() error {
		eventWorker.Start(circuit.Context())

		return nil
	})
	slog.Info("events-engine was started.")

	cc := func(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File) error {
		err = srv.flow.configureWorkflowStarts(ctx, db, nsID, nsName, file)
		if err != nil {
			return err
		}

		err = srv.flow.placeholdSecrets(ctx, db, nsName, file)
		if err != nil {
			slog.Debug("failed setting up placeholder secrets", "error", err, string(core.LogTrackKey), "namespace."+nsName, "namespace", nsName, "file", file.Path)
		}

		return nil
	}

	srv.MirrorManager = mirror.NewManager(
		&mirrorCallbacks{
			logger: &mirrorProcessLogger{
				// logger: srv.logger,
			},
			store:    db.DataStore().Mirror(),
			fstore:   db.FileStore(),
			varstore: db.DataStore().RuntimeVariables(),
			wfconf:   cc,
		},
	)

	srv.registerFunctions()

	go srv.cronPoller()

	circuit.Start(func() error {
		<-circuit.Done()
		telEnd()
		srv.cleanup(srv.pubsub.Close)
		srv.cleanup(srv.timers.Close)
		if srv.nats != nil {
			srv.nats.Close()
			slog.Info("NATS connection closed")
		}

		return nil
	})

	if config.NatsInstalled {
		var err error
		slog.Info("conecting to NATS", "config.NatsHost", config.NatsHost, "config.NatsPort", config.NatsPort)
		for i := range 12 {
			err = srv.initNATS(config)
			if err == nil {
				slog.Info("NATS connection established successfully")
				break
			}
			slog.Error("failed to connect to NATS, retrying...", "attempt", i+1, "error", err)
			time.Sleep(time.Duration(i+2) * time.Second)
		}

		if err != nil {
			return nil, fmt.Errorf("initialize NATS connection, err: %w", err)
		}

		if err := srv.publishDemoMessage("test", "test-connection"); err != nil {
			return nil, fmt.Errorf("testing NATS connection, err: %w", err)
		}
		slog.Info("connected to NATS")
	}

	srv.ConfigureWorkflow = func(event *pubsub2.FileSystemChangeEvent) error {
		// If this is a delete workflow file
		if event.DeleteFileID.String() != (uuid.UUID{}).String() {
			return srv.flow.events.deleteWorkflowEventListeners(circuit.Context(), event.NamespaceID, event.DeleteFileID)
		}
		file, err := db.FileStore().ForNamespace(event.Namespace).GetFile(circuit.Context(), event.FilePath)
		if err != nil {
			return err
		}
		err = srv.flow.configureWorkflowStarts(circuit.Context(), db, event.NamespaceID, event.Namespace, file)
		if err != nil {
			return err
		}

		return srv.flow.placeholdSecrets(circuit.Context(), db, event.Namespace, file)
	}

	return srv, nil
}

func (srv *server) cleanup(closer func() error) {
	err := closer()
	if err != nil {
		slog.Error("Server cleanup failed.", "error", err)
	}
}

func (srv *server) NotifyCluster(msg string) error {
	ctx := context.Background()

	conn, err := srv.rawDB.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "SELECT pg_notify($1, $2)", pubsub.FlowSync, msg)

	perr := new(pq.Error)

	if errors.As(err, &perr) {
		slog.Error("Database notification to cluster failed.", "error", perr)
		if perr.Code == "57014" {
			return fmt.Errorf("canceled query")
		}

		return err
	}

	return nil
}

func (srv *server) NotifyHostname(hostname, msg string) error {
	ctx := context.Background()

	conn, err := srv.rawDB.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	channel := fmt.Sprintf("hostname:%s", hostname)

	_, err = conn.ExecContext(ctx, "SELECT pg_notify($1, $2)", channel, msg)

	perr := new(pq.Error)

	if errors.As(err, &perr) {
		fmt.Fprintf(os.Stderr, "db notification failed: %v", perr)
		if perr.Code == "57014" {
			return fmt.Errorf("canceled query")
		}

		return err
	}

	return nil
}

func (srv *server) PublishToCluster(payload string) {
	srv.pubsub.Publish(&pubsub.PubsubUpdate{
		Handler: nohome.PubsubNotifyFunction,
		Key:     payload,
	})
}

func (srv *server) CacheNotify(req *pubsub.PubsubUpdate) {
	if srv.ID.String() == req.Sender {
		return
	}

	// TODO: Alan, needfix.
	// srv.database.HandleNotification(req.Key)
}

func (srv *server) registerFunctions() {
	srv.pubsub.RegisterFunction(nohome.PubsubNotifyFunction, srv.CacheNotify)

	srv.pubsub.RegisterFunction(pubsub.PubsubNotifyFunction, srv.pubsub.Notify)
	srv.pubsub.RegisterFunction(pubsub.PubsubDisconnectFunction, srv.pubsub.Disconnect)
	srv.pubsub.RegisterFunction(pubsub.PubsubDeleteTimerFunction, srv.timers.deleteTimerHandler)
	srv.pubsub.RegisterFunction(pubsub.PubsubDeleteInstanceTimersFunction, srv.timers.deleteInstanceTimersHandler)
	srv.pubsub.RegisterFunction(pubsub.PubsubCancelWorkflowFunction, srv.Engine.finishCancelWorkflow)
	srv.pubsub.RegisterFunction(pubsub.PubsubCancelMirrorProcessFunction, srv.Engine.finishCancelMirrorProcess)
	srv.pubsub.RegisterFunction(pubsub.PubsubConfigureRouterFunction, srv.flow.configureRouterHandler)

	srv.timers.registerFunction(timeoutFunction, srv.Engine.timeoutHandler)
	srv.timers.registerFunction(wfCron, srv.flow.cronHandler)
	srv.timers.registerFunction(retryWakeupFunction, srv.flow.Engine.retryWakeup)

	srv.pubsub.RegisterFunction(pubsub.PubsubDeleteActivityTimersFunction, srv.timers.deleteActivityTimersHandler)
}

func (srv *server) cronPoller() {
	for {
		srv.cronPoll()
		time.Sleep(time.Minute * 15)
	}
}

func (srv *server) cronPoll() {
	ctx := context.Background()
	tx, err := srv.flow.beginSQLTx(ctx)
	if err != nil {
		slog.Error("cronPoll executing transaction", "error", err)
		return
	}
	defer tx.Rollback()

	roots, err := tx.FileStore().GetAllRoots(ctx)
	if err != nil {
		slog.Error("cronPoll fetching all Roots from db", "error", err)
		return
	}

	for _, root := range roots {
		files, err := tx.FileStore().ForRootID(root.ID).ListAllFiles(ctx)
		if err != nil {
			slog.Error("cronPoll fetching RootID from db", "error", err)
			return
		}

		for _, file := range files {
			if file.Typ != filestore.FileTypeWorkflow {
				continue
			}

			srv.cronPollerWorkflow(ctx, tx, file)
		}
	}
}

func (srv *server) cronPollerWorkflow(ctx context.Context, tx *database.DB, file *filestore.File) {
	ms, err := validateRouter(ctx, tx, file)
	if err != nil {
		slog.Error("Failed to validate Routing for a cron schedule.", "error", err)
		return
	}

	if ms.Cron != "" {
		srv.timers.deleteCronForWorkflow(file.ID.String())
	}

	if ms.Cron != "" {
		err := srv.timers.addCron(file.ID.String(), wfCron, ms.Cron, []byte(file.ID.String()))
		if err != nil {
			slog.Error("Failed to add cron schedule for workflow", "error", err, "cron_expression", ms.Cron)
			return
		}

		slog.Debug("Successfully loaded cron schedule for workflow", "workflow", file.Path, "cron_expression", ms.Cron)
	}
}

func this() string {
	pc, _, _, _ := runtime.Caller(1) //nolint:dogsled
	fn := runtime.FuncForPC(pc)
	elems := strings.Split(fn.Name(), ".")

	return elems[len(elems)-1]
}

func (srv *server) beginSQLTx(ctx context.Context, opts ...*sql.TxOptions) (*database.DB, error) {
	return srv.db.BeginTx(ctx, opts...)
}

func (srv *server) runSQLTx(ctx context.Context, fun func(tx *database.DB) error) error {
	tx, err := srv.beginSQLTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fun(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Initialize NATS connection using environment variables.
func (srv *server) initNATS(config *core.Config) error {
	natsURL := config.NatsHost
	if natsURL == "" {
		return fmt.Errorf("NATS URL not set in environment")
	}

	// Connect to NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS, err: %w", err)
	}

	// Store the NATS connection in the server struct
	srv.nats = nc
	slog.Info("successfully connected to NATS")

	return nil
}

// Example of publishing a message to NATS.
func (srv *server) publishDemoMessage(subject, msg string) error {
	err := srv.nats.Publish(subject, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to publish message, err: %w", err)
	}
	slog.Info("message published", "subject", subject, "msg", msg)

	return nil
}
