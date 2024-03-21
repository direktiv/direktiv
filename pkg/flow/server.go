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
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/metrics"
	"github.com/direktiv/direktiv/pkg/refactor/cmd"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	database2 "github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	eventsstore "github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/filestoresql"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore/instancestoresql"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	pubsub2 "github.com/direktiv/direktiv/pkg/refactor/pubsub"
	pubsubSQL "github.com/direktiv/direktiv/pkg/refactor/pubsub/sql"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"
	libgrpc "google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	parcelSize = 0x100000
)

type server struct {
	ID uuid.UUID

	conf *core.Config

	// db       *ent.Client
	pubsub *pubsub.Pubsub

	// the new pubsub bus
	pBus *pubsub2.Bus

	timers *timers
	engine *engine

	gormDB *gorm.DB

	rawDB *sql.DB

	mirrorManager *mirror.Manager

	flow     *flow
	internal *internal
	events   *events

	metrics *metrics.Client
}

func Run(serverCtx context.Context) error {
	srv, err := newServer()
	if err != nil {
		return err
	}

	config := &core.Config{}
	if err := env.Parse(config); err != nil {
		return fmt.Errorf("parsing env variables: %w", err)
	}
	if config.IsValid() != nil {
		return fmt.Errorf("parsing env variables: %w", config.IsValid())
	}

	srv.conf = config

	err = srv.start(serverCtx)
	if err != nil {
		return err
	}

	return nil
}

func newServer() (*server, error) {
	srv := new(server)
	srv.ID = uuid.New()

	srv.initJQ()

	return srv, nil
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
	logger    mirror.ProcessLogger
	syslogger *zap.SugaredLogger
	store     datastore.MirrorStore
	fstore    filestore.FileStore
	varstore  datastore.RuntimeVariablesStore
	wfconf    func(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File) error
}

func (c *mirrorCallbacks) ConfigureWorkflowFunc(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File) error {
	return c.wfconf(ctx, nsID, nsName, file)
}

func (c *mirrorCallbacks) ProcessLogger() mirror.ProcessLogger {
	return c.logger
}

func (c *mirrorCallbacks) SysLogCrit(msg string) {
	c.syslogger.Error(msg)
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

func (srv *server) start(serverCtx context.Context) error {
	var err error
	slog.Debug("Starting Flow server")
	slog.Debug("Initializing telemetry.")
	telend, err := util.InitTelemetry(srv.conf.OpenTelemetry, "direktiv/flow", "direktiv")
	if err != nil {
		return err
	}
	defer telend()
	slog.Info("Telemetry initialized successfully.")

	go func() {
		err := setupPrometheusEndpoint()
		if err != nil {
			slog.Error("Failed to set up Prometheus endpoint", "error", err)
		}
	}()

	db := srv.conf.DB

	slog.Debug("Initializing database.")
	gormConf := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
			},
		),
	}

	for i := 0; i < 10; i++ {
		slog.Debug("Connecting to database...")

		srv.gormDB, err = gorm.Open(postgres.New(postgres.Config{
			DSN:                  db,
			PreferSimpleProtocol: false, // disables implicit prepared statement usage
			// Conn:                 edb.DB(),
		}), gormConf)
		if err == nil {
			slog.Debug("Successfully connected to the database.")
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		return fmt.Errorf("creating gorm db driver, err: %w", err)
	}
	slog.Info("Database connection established.")

	res := srv.gormDB.Exec(database2.Schema)
	if res.Error != nil {
		return fmt.Errorf("provisioning schema, err: %w", res.Error)
	}
	slog.Info("Schema provisioned successfully")

	gdb, err := srv.gormDB.DB()
	if err != nil {
		return fmt.Errorf("modifying gorm driver, err: %w", err)
	}
	gdb.SetMaxIdleConns(32)
	gdb.SetMaxOpenConns(16)
	slog.Debug("Database connection pool limits set", "maxIdleConns", 32, "maxOpenConns", 16)

	srv.rawDB, err = sql.Open("postgres", db)
	if err == nil {
		err = srv.rawDB.Ping()
	}
	if err != nil {
		return fmt.Errorf("creating raw db driver, err: %w", err)
	}
	slog.Debug("Successfully connected to database with raw driver")

	// Repeat SecretKey length to 16 chars.
	srv.conf.SecretKey = srv.conf.SecretKey + "1234567890123456"
	srv.conf.SecretKey = srv.conf.SecretKey[0:16]

	slog.Debug("Initializing pub-sub.")

	srv.pubsub, err = pubsub.InitPubSub(srv, db)
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.pubsub.Close)
	slog.Info("pub-sub was initialized successfully.")

	slog.Debug("Initializing timers.")

	srv.timers, err = initTimers(srv.pubsub)
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.timers.Close)
	slog.Info("timers where initialized successfully.")

	slog.Debug("Initializing metrics.")

	srv.metrics = metrics.NewClient(srv.gormDB)
	slog.Info("Metrics Client was created.")

	var lock sync.Mutex
	var wg sync.WaitGroup

	wg.Add(5)

	cctx, cancel := context.WithCancel(serverCtx)
	defer cancel()

	slog.Debug("Initializing pubsub routine.")
	coreBus, err := pubsubSQL.NewPostgresCoreBus(srv.rawDB, srv.conf.DB)
	if err != nil {
		return fmt.Errorf("creating pubsub core bus, err: %w", err)
	}
	slog.Info("pubsub routine was initialized.")

	srv.pBus = pubsub2.NewBus(coreBus)
	go srv.pBus.Start(cctx.Done(), &wg)

	slog.Debug("Initializing engine.")

	srv.engine, err = initEngine(srv)
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.engine.Close)
	slog.Info("engine was started.")

	slog.Debug("Initializing internal grpc server.")

	srv.internal, err = initInternalServer(cctx, srv)
	if err != nil {
		return err
	}
	slog.Info("Internal grpc server started.")

	srv.flow, err = initFlowServer(cctx, srv)
	if err != nil {
		return err
	}

	slog.Debug("Initializing mirror manager.")
	noTx := &sqlTx{
		res:       srv.gormDB,
		secretKey: srv.conf.SecretKey,
	}
	slog.Debug("mirror manager was started.")

	slog.Debug("Initializing events.")
	srv.events, err = initEvents(srv, noTx.DataStore().StagingEvents().Append)
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.events.Close)

	slog.Debug("Initializing EventWorkers.")

	interval := 1 * time.Second // TODO: Adjust the polling interval
	eventWorker := eventsstore.NewEventWorker(noTx.DataStore().StagingEvents(), interval, srv.events.handleEvent)

	go eventWorker.Start(serverCtx)
	slog.Info("Events-engine was started.")

	cc := func(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File) error {
		err = srv.flow.configureWorkflowStarts(ctx, noTx, nsID, file)
		if err != nil {
			return err
		}

		err = srv.flow.placeholdSecrets(ctx, noTx, nsName, file)
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
			store:    noTx.DataStore().Mirror(),
			fstore:   noTx.FileStore(),
			varstore: noTx.DataStore().RuntimeVariables(),
			wfconf:   cc,
		},
	)

	if srv.conf.EnableEventing {
		slog.Debug("Initializing knative eventing receiver.")
		rcv, err := newEventReceiver(srv.events, srv.flow)
		if err != nil {
			return err
		}

		// starting the event receiver
		go rcv.Start()
	}

	srv.registerFunctions()

	go srv.cronPoller()

	go func() {
		defer wg.Done()
		defer cancel()
		e := srv.internal.Run()
		if e != nil {
			slog.Error("srv.internal.Run()", "error", err)
			lock.Lock()
			if err == nil {
				err = e
			}
			lock.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		defer cancel()
		e := srv.flow.Run()
		if e != nil {
			slog.Error("srv.flow.Run()", "error", err)
			lock.Lock()
			if err == nil {
				err = e
			}
			lock.Unlock()
		}
	}()

	// TODO: yassir, use the new db to refactor old code.
	dbManager := database2.NewDB(srv.gormDB, srv.conf.SecretKey)

	configureWorkflow := func(data string) error {
		event := pubsub2.FileChangeEvent{}
		err := json.Unmarshal([]byte(data), &event)
		if err != nil {
			slog.Error("critical! unmarshal file change event error", "error", err)
			panic("unmarshal file change event")
		}
		// If this is a delete workflow file
		if event.DeleteFileID.String() != (uuid.UUID{}).String() {
			return srv.flow.events.deleteWorkflowEventListeners(serverCtx, event.NamespaceID, event.DeleteFileID)
		}
		file, err := noTx.FileStore().ForNamespace(event.Namespace).GetFile(serverCtx, event.FilePath)
		if err != nil {
			return err
		}
		err = srv.flow.configureWorkflowStarts(serverCtx, noTx, event.NamespaceID, file)
		if err != nil {
			return err
		}

		return srv.flow.placeholdSecrets(serverCtx, noTx, event.Namespace, file)
	}

	instanceManager := &instancestore.InstanceManager{
		Start:  srv.engine.StartWorkflow,
		Cancel: srv.engine.CancelInstance,
	}

	newMainWG := cmd.NewMain(serverCtx, &cmd.NewMainArgs{
		Config:            srv.conf,
		Database:          dbManager,
		PubSubBus:         srv.pBus,
		ConfigureWorkflow: configureWorkflow,
		InstanceManager:   instanceManager,
	})

	slog.Info("Flow server started.")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		panic("TODO: Alan, remove this panic and handle signal gracefully")
	}()

	wg.Wait()

	newMainWG.Wait()

	if err != nil {
		return err
	}

	return nil
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
		Handler: database.PubsubNotifyFunction,
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
	srv.pubsub.RegisterFunction(database.PubsubNotifyFunction, srv.CacheNotify)

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
	tx, err := srv.flow.beginSqlTx(ctx)
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

func (srv *server) cronPollerWorkflow(ctx context.Context, tx *sqlTx, file *filestore.File) {
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

func unaryInterceptor(ctx context.Context, req interface{}, info *libgrpc.UnaryServerInfo, handler libgrpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
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
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	elems := strings.Split(fn.Name(), ".")
	return elems[len(elems)-1]
}

type sqlTx struct {
	res       *gorm.DB
	secretKey string
}

func (tx *sqlTx) FileStore() filestore.FileStore {
	return filestoresql.NewSQLFileStore(tx.res)
}

func (tx *sqlTx) DataStore() datastore.Store {
	return datastoresql.NewSQLStore(tx.res, tx.secretKey)
}

func (tx *sqlTx) InstanceStore() instancestore.Store {
	return instancestoresql.NewSQLInstanceStore(tx.res)
}

func (tx *sqlTx) Commit(ctx context.Context) error {
	return tx.res.WithContext(ctx).Commit().Error
}

func (tx *sqlTx) Rollback() {
	err := tx.res.Rollback().Error
	if err != nil {
		if !strings.Contains(err.Error(), "already") {
			fmt.Fprintf(os.Stderr, "failed to rollback transaction: %v\n", err)
		}
	}
}

func (srv *server) beginSqlTx(ctx context.Context, opts ...*sql.TxOptions) (*sqlTx, error) {
	res := srv.gormDB.WithContext(ctx).Begin(opts...)
	if res.Error != nil {
		return nil, res.Error
	}
	return &sqlTx{
		res:       res,
		secretKey: srv.conf.SecretKey,
	}, nil
}

func (srv *server) runSqlTx(ctx context.Context, fun func(tx *sqlTx) error) error {
	tx, err := srv.beginSqlTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fun(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
