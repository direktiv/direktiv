package flow

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/direktiv/direktiv/pkg/flow/nohome"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/refactor/cmd"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	eventsstore "github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	pubsub2 "github.com/direktiv/direktiv/pkg/refactor/pubsub"
	pubsubSQL "github.com/direktiv/direktiv/pkg/refactor/pubsub/sql"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"github.com/lib/pq"
	libgrpc "google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type server struct {
	ID uuid.UUID

	config *core.Config

	// db       *ent.Client
	pubsub *pubsub.Pubsub

	// the new pubsub bus
	pBus *pubsub2.Bus

	timers *timers
	engine *engine

	gormDB *gorm.DB
	rawDB  *sql.DB

	sqlStore *database.SQLStore

	mirrorManager *mirror.Manager

	flow     *flow
	internal *internal
	events   *events
}

func Run(circuit *core.Circuit) error {
	config := &core.Config{}
	if err := env.Parse(config); err != nil {
		return fmt.Errorf("parsing env variables: %w", err)
	}
	if err := config.Init(); err != nil {
		return fmt.Errorf("init config, err: %w", err)
	}

	slog.Info("initialize db connection")
	db, err := initDB(config)
	if err != nil {
		return fmt.Errorf("initialize db, err: %w", err)
	}
	// TODO: yassir, use the new db to refactor old code.
	dbManager := database.NewSQLStore(db, config.SecretKey)

	slog.Info("initialize legacy server")
	srv, err := initLegacyServer(circuit, config, db, dbManager)
	if err != nil {
		return fmt.Errorf("initialize legacy server, err: %w", err)
	}

	configureWorkflow := func(data string) error {
		event := pubsub2.FileChangeEvent{}
		err := json.Unmarshal([]byte(data), &event)
		if err != nil {
			slog.Error("critical! unmarshal file change event error", "error", err)
			panic("unmarshal file change event")
		}
		// If this is a delete workflow file
		if event.DeleteFileID.String() != (uuid.UUID{}).String() {
			return srv.flow.events.deleteWorkflowEventListeners(circuit.Context(), event.NamespaceID, event.DeleteFileID)
		}
		file, err := dbManager.FileStore().ForNamespace(event.Namespace).GetFile(circuit.Context(), event.FilePath)
		if err != nil {
			return err
		}
		err = srv.flow.configureWorkflowStarts(circuit.Context(), dbManager, event.NamespaceID, event.Namespace, file)
		if err != nil {
			return err
		}

		return srv.flow.placeholdSecrets(circuit.Context(), dbManager, event.Namespace, file)
	}

	instanceManager := &instancestore.InstanceManager{
		Start:  srv.engine.StartWorkflow,
		Cancel: srv.engine.CancelInstance,
	}

	err = cmd.NewMain(circuit, &cmd.NewMainArgs{
		Config:              srv.config,
		Database:            dbManager,
		PubSubBus:           srv.pBus,
		ConfigureWorkflow:   configureWorkflow,
		InstanceManager:     instanceManager,
		WakeInstanceByEvent: srv.engine.WakeEventsWaiter,
		WorkflowStart:       srv.engine.EventsInvoke,
		SyncNamespace: func(namespace any, mirrorConfig any) (any, error) {
			ns := namespace.(*datastore.Namespace)            //nolint:forcetypeassert
			mConfig := mirrorConfig.(*datastore.MirrorConfig) //nolint:forcetypeassert
			proc, err := srv.mirrorManager.NewProcess(context.Background(), ns, datastore.ProcessTypeSync)
			if err != nil {
				return nil, err
			}

			go func() {
				srv.mirrorManager.Execute(context.Background(), proc, mConfig, &mirror.DirektivApplyer{NamespaceID: ns.ID})
				err := srv.pBus.Publish(pubsub2.MirrorSync, ns.Name)
				if err != nil {
					slog.Error("pubsub publish", "error", err)
				}
			}()

			return proc, nil
		},
	})
	if err != nil {
		return fmt.Errorf("lunching new main, err: %w", err)
	}

	return nil
}

type mirrorProcessLogger struct{}

func (log *mirrorProcessLogger) Debug(pid uuid.UUID, msg string, kv ...interface{}) {
	slog.Debug(fmt.Sprintf(msg, kv...), "activity", pid, "track", "activity."+pid.String())
}

func (log *mirrorProcessLogger) Info(pid uuid.UUID, msg string, kv ...interface{}) {
	slog.Info(fmt.Sprintf(msg, kv...), "activity", pid, "track", "activity."+pid.String())
}

func (log *mirrorProcessLogger) Warn(pid uuid.UUID, msg string, kv ...interface{}) {
	slog.Warn(fmt.Sprintf(msg, kv...), "activity", pid, "track", "activity"+"."+pid.String())
}

func (log *mirrorProcessLogger) Error(pid uuid.UUID, msg string, kv ...interface{}) {
	slog.Error(fmt.Sprintf(msg, kv...), "activity", pid, "track", "activity"+"."+pid.String())
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

func initLegacyServer(circuit *core.Circuit, config *core.Config, db *gorm.DB, dbManager *database.SQLStore) (*server, error) {
	srv := new(server)
	srv.ID = uuid.New()
	srv.initJQ()
	srv.config = config

	var err error
	slog.Debug("Starting Flow server")
	slog.Debug("Initializing telemetry.")
	telend, err := util.InitTelemetry(srv.config.OpenTelemetry, "direktiv/flow", "direktiv")
	if err != nil {
		return nil, err
	}
	defer telend()
	slog.Info("Telemetry initialized successfully.")

	srv.gormDB = db
	srv.sqlStore = dbManager

	srv.rawDB, err = sql.Open("postgres", config.DB)
	if err == nil {
		err = srv.rawDB.Ping()
	}
	if err != nil {
		return nil, fmt.Errorf("creating raw db driver, err: %w", err)
	}
	slog.Debug("successfully connected to database with raw driver")

	slog.Debug("Initializing pub-sub.")

	srv.pubsub, err = pubsub.InitPubSub(srv, config.DB)
	if err != nil {
		return nil, err
	}
	defer srv.cleanup(srv.pubsub.Close)
	slog.Info("pub-sub was initialized successfully.")

	slog.Debug("Initializing timers.")

	srv.timers, err = initTimers(srv.pubsub)
	if err != nil {
		return nil, err
	}
	defer srv.cleanup(srv.timers.Close)
	slog.Info("timers where initialized successfully.")

	slog.Debug("Initializing pubsub routine.")
	coreBus, err := pubsubSQL.NewPostgresCoreBus(srv.rawDB, srv.config.DB)
	if err != nil {
		return nil, fmt.Errorf("creating pubsub core bus, err: %w", err)
	}
	slog.Info("pubsub routine was initialized.")

	srv.pBus = pubsub2.NewBus(coreBus)

	circuit.Start(func() error {
		// TODO: yassir, Implement bus crash handling.
		srv.pBus.Start(circuit)

		return nil
	})

	slog.Debug("Initializing engine.")

	srv.engine = initEngine(srv)
	defer srv.cleanup(srv.engine.Close)
	slog.Info("engine was started.")

	slog.Debug("Initializing internal grpc server.")

	srv.internal, err = initInternalServer(circuit.Context(), srv)
	if err != nil {
		return nil, err
	}
	slog.Info("Internal grpc server started.")

	srv.flow, err = initFlowServer(circuit.Context(), srv)
	if err != nil {
		return nil, err
	}

	slog.Debug("Initializing mirror manager.")
	slog.Debug("mirror manager was started.")

	slog.Debug("Initializing events.")
	srv.events = initEvents(srv, dbManager.DataStore().StagingEvents().Append)
	defer srv.cleanup(srv.events.Close)

	slog.Debug("Initializing EventWorkers.")

	interval := 1 * time.Second // TODO: Adjust the polling interval
	eventWorker := eventsstore.NewEventWorker(dbManager.DataStore().StagingEvents(), interval, srv.events.handleEvent)

	circuit.Start(func() error {
		eventWorker.Start(circuit.Context())

		return nil
	})
	slog.Info("Events-engine was started.")

	cc := func(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File) error {
		err = srv.flow.configureWorkflowStarts(ctx, dbManager, nsID, nsName, file)
		if err != nil {
			return err
		}

		err = srv.flow.placeholdSecrets(ctx, dbManager, nsName, file)
		if err != nil {
			slog.Debug("Error setting up placeholder secrets", "error", err, "track", "namespace."+nsName, "namespace", nsName, "file", file.Path)
		}

		return nil
	}

	srv.mirrorManager = mirror.NewManager(
		&mirrorCallbacks{
			logger: &mirrorProcessLogger{
				// logger: srv.logger,
			},
			store:    dbManager.DataStore().Mirror(),
			fstore:   dbManager.FileStore(),
			varstore: dbManager.DataStore().RuntimeVariables(),
			wfconf:   cc,
		},
	)

	// Todo: Yassir, delete this entirely in cycle6.
	// if srv.config.EnableEventing {
	// 	slog.Debug("Initializing knative eventing receiver.")
	// 	rcv, err := newEventReceiver(srv.events, srv.flow)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	// starting the event receiver
	// 	go rcv.Start()
	// }

	srv.registerFunctions()

	go srv.cronPoller()

	circuit.Start(func() error {
		e := srv.internal.Run()
		if e != nil {
			return fmt.Errorf("srv.internal.Run(), err: %w", err)
		}

		return nil
	})

	circuit.Start(func() error {
		e := srv.flow.Run()
		if e != nil {
			return fmt.Errorf("srv.flow.Run(), err: %w", err)
		}

		return nil
	})

	return srv, nil
}

func initDB(config *core.Config) (*gorm.DB, error) {
	gormConf := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
			},
		),
	}

	var err error
	var db *gorm.DB
	for i := 0; i < 10; i++ {
		slog.Info("connecting to database...")

		db, err = gorm.Open(postgres.New(postgres.Config{
			DSN:                  config.DB,
			PreferSimpleProtocol: false, // disables implicit prepared statement usage
			// Conn:                 edb.SQLStore(),
		}), gormConf)
		if err == nil {
			slog.Info("successfully connected to the database.")

			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		return nil, err
	}

	res := db.Exec(database.Schema)
	if res.Error != nil {
		return nil, fmt.Errorf("provisioning schema, err: %w", res.Error)
	}
	slog.Info("Schema provisioned successfully")

	gdb, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("modifying gorm driver, err: %w", err)
	}

	slog.Debug("Database connection pool limits set", "maxIdleConns", 32, "maxOpenConns", 16)
	gdb.SetMaxIdleConns(32)
	gdb.SetMaxOpenConns(16)

	return db, nil
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
	srv.pubsub.RegisterFunction(pubsub.PubsubCancelWorkflowFunction, srv.engine.finishCancelWorkflow)
	srv.pubsub.RegisterFunction(pubsub.PubsubCancelMirrorProcessFunction, srv.engine.finishCancelMirrorProcess)
	srv.pubsub.RegisterFunction(pubsub.PubsubConfigureRouterFunction, srv.flow.configureRouterHandler)

	srv.timers.registerFunction(timeoutFunction, srv.engine.timeoutHandler)
	srv.timers.registerFunction(wfCron, srv.flow.cronHandler)
	srv.timers.registerFunction(retryWakeupFunction, srv.flow.engine.retryWakeup)

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

func (srv *server) cronPollerWorkflow(ctx context.Context, tx *database.SQLStore, file *filestore.File) {
	ms, err := srv.validateRouter(ctx, tx, file)
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

func unaryInterceptor(ctx context.Context, req interface{}, info *libgrpc.UnaryServerInfo, handler libgrpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		return nil, translateError(err)
	}

	return resp, nil
}

func streamInterceptor(srv interface{}, ss libgrpc.ServerStream, info *libgrpc.StreamServerInfo, handler libgrpc.StreamHandler) error {
	err := handler(srv, ss)
	if err != nil {
		return translateError(err)
	}

	return nil
}

func this() string {
	pc, _, _, _ := runtime.Caller(1) //nolint:dogsled
	fn := runtime.FuncForPC(pc)
	elems := strings.Split(fn.Name(), ".")

	return elems[len(elems)-1]
}

func (srv *server) beginSQLTx(ctx context.Context, opts ...*sql.TxOptions) (*database.SQLStore, error) {
	return srv.sqlStore.BeginTx(ctx, opts...)
}

func (srv *server) runSQLTx(ctx context.Context, fun func(tx *database.SQLStore) error) error {
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
