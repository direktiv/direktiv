package flow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/direktiv/direktiv/pkg/cluster"
	"github.com/direktiv/direktiv/pkg/dlog"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	igrpc "github.com/direktiv/direktiv/pkg/functions/grpc"
	"github.com/direktiv/direktiv/pkg/metrics"
	database2 "github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/filestoresql"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore/instancestoresql"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq" // postgres for ent
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	libgrpc "google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	parcelSize        = 0x100000
	direktivSecretKey = "DIREKTIV_SECRETS_KEY"
)

type server struct {
	ID uuid.UUID

	sugar    *zap.SugaredLogger
	fnLogger *zap.SugaredLogger
	conf     *util.Config

	// db       *ent.Client
	pubsub *pubsub.Pubsub
	locks  *locks
	timers *timers
	engine *engine

	gormDB *gorm.DB

	rawDB *sql.DB

	mirrorManager mirror.Manager

	flow            *flow
	internal        *internal
	events          *events
	functionsClient igrpc.FunctionsClient

	metrics *metrics.Client
	logger  logengine.BetterLogger
}

func Run(ctx context.Context, logger *zap.SugaredLogger, conf *util.Config) error {
	srv, err := newServer(logger, conf)
	if err != nil {
		return err
	}

	err = srv.start(ctx)
	if err != nil {
		return err
	}

	return nil
}

func newServer(logger *zap.SugaredLogger, conf *util.Config) (*server, error) {
	var err error

	srv := new(server)
	srv.ID = uuid.New()

	srv.sugar = logger
	srv.conf = conf

	srv.fnLogger, err = dlog.FunctionsLogger()
	if err != nil {
		return nil, err
	}

	srv.initJQ()

	return srv, nil
}

type gormLogger struct {
	*zap.SugaredLogger
}

func (g gormLogger) Write(p []byte) (n int, err error) {
	g.Debugw(string(p), "component", "GORM")
	return len(p), nil
}

//nolint:gocyclo
func (srv *server) start(ctx context.Context) error {
	var err error
	// enableExperimentalFeatures := os.Getenv("ENABLE_EXPERIMENTAL_FEATURES") == "true"
	enableDeveloperMode := os.Getenv("ENABLE_DEVELOPER_MODE") == "true"

	srv.sugar.Debug("Initializing telemetry.")
	telend, err := util.InitTelemetry(srv.conf, "direktiv/flow", "direktiv")
	if err != nil {
		return err
	}
	defer telend()

	go func() {
		err := setupPrometheusEndpoint()
		if err != nil {
			srv.sugar.Errorf("Failed to set up Prometheus endpoint: %v.", err)
		}
	}()

	srv.sugar.Debug("Initializing locks.")

	db := os.Getenv(util.DBConn)

	srv.locks, err = initLocks(db)
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.locks.Close)

	srv.sugar.Debug("Initializing database.")
	gormConf := &gorm.Config{}
	jsonV := "json"
	if enableDeveloperMode && os.Getenv(util.DirektivLogJSON) == jsonV {
		gormConf = &gorm.Config{
			Logger: logger.New(
				log.New(gormLogger{SugaredLogger: srv.sugar}, "\r\n", log.LstdFlags),
				logger.Config{
					LogLevel:                  logger.Warn,
					IgnoreRecordNotFoundError: true,
				},
			),
		}
	}
	if enableDeveloperMode && os.Getenv(util.DirektivLogJSON) != jsonV {
		gormConf = &gorm.Config{
			Logger: logger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags),
				logger.Config{
					LogLevel: logger.Info,
				},
			),
		}
	}
	srv.gormDB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  db,
		PreferSimpleProtocol: false, // disables implicit prepared statement usage
		// Conn:                 edb.DB(),
	}), gormConf)

	if err != nil {
		return fmt.Errorf("creating gorm db driver, err: %w", err)
	}

	res := srv.gormDB.Exec(database2.Schema)
	if res.Error != nil {
		return fmt.Errorf("provisioning schema, err: %w", res.Error)
	}

	gdb, err := srv.gormDB.DB()
	if err != nil {
		return fmt.Errorf("modifying gorm driver, err: %w", err)
	}
	gdb.SetMaxIdleConns(32)
	gdb.SetMaxOpenConns(16)

	srv.rawDB, err = sql.Open("postgres", db)
	if err == nil {
		err = srv.rawDB.Ping()
	}
	if err != nil {
		return fmt.Errorf("creating raw db driver, err: %w", err)
	}
	if enableDeveloperMode {
		fmt.Printf(">>>>>> dsn %s\n", db)
	}

	if os.Getenv(direktivSecretKey) == "" {
		return fmt.Errorf("empty env variable '%s'", direktivSecretKey)
	}

	if len(os.Getenv(direktivSecretKey))%16 != 0 {
		return fmt.Errorf("invalid env variable '%s' length", direktivSecretKey)
	}

	srv.sugar.Debug("Initializing pub-sub.")

	srv.pubsub, err = pubsub.InitPubSub(srv.sugar, srv, db)
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.pubsub.Close)
	srv.sugar.Debug("Initializing timers.")

	srv.timers, err = initTimers(srv.pubsub)
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.timers.Close)

	srv.events, err = initEvents(srv)
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.events.Close)

	srv.sugar.Debug("Initializing metrics.")

	srv.metrics = metrics.NewClient(srv.gormDB)

	srv.sugar.Debug("Initializing engine.")

	srv.engine, err = initEngine(srv)
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.engine.Close)

	var lock sync.Mutex
	var wg sync.WaitGroup

	wg.Add(5)

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	srv.sugar.Debug("Initializing internal grpc server.")

	srv.internal, err = initInternalServer(cctx, srv)
	if err != nil {
		return err
	}

	srv.sugar.Debug("Initializing flow grpc server.")

	srv.flow, err = initFlowServer(cctx, srv)
	if err != nil {
		return err
	}

	srv.sugar.Debug("Initializing mirror manager.")
	noTx := &sqlTx{
		res: srv.gormDB,
	}
	dbLogger, logworker, closelogworker := logengine.NewCachedLogger(1024,
		noTx.DataStore().Logs().Append,
		func(objectID uuid.UUID, objectType string) {
			srv.pubsub.NotifyLogs(objectID, recipient.RecipientType(objectType))
		},
		srv.sugar.Errorf,
	)

	addTrace := func(ctx context.Context, toTags map[string]string) map[string]string {
		span := trace.SpanFromContext(ctx)
		tid := span.SpanContext().TraceID()
		toTags["trace"] = tid.String()
		return toTags
	}
	if enableDeveloperMode {
		addTrace = func(ctx context.Context, toTags map[string]string) map[string]string {
			_ = ctx
			return toTags
		}
	}
	var sugarBetterLogger logengine.BetterLogger
	sugarBetterLogger = logengine.SugarBetterJSONLogger{
		Sugar:        srv.sugar.Named("userLogger"),
		AddTraceFrom: addTrace,
	}
	if os.Getenv(util.DirektivLogJSON) != "json" {
		sugarBetterLogger = logengine.SugarBetterConsoleLogger{
			Sugar:        srv.sugar.Named("userLogger"),
			AddTraceFrom: addTrace,
			RetainTags:   []string{"caller", "Caller"}, // if set to nil all tags will be retained
		}
	}
	srv.logger = logengine.ChainedBetterLogger{
		sugarBetterLogger,
		dbLogger,
	}

	go func() {
		logworker()
	}()

	cc := func(ctx context.Context, file *filestore.File) error {
		_, router, err := getRouter(ctx, noTx, file)
		if err != nil {
			return err
		}

		err = srv.flow.configureWorkflowStarts(ctx, noTx, file.RootID, file, router, false)
		if err != nil {
			return err
		}

		return nil
	}

	srv.mirrorManager = mirror.NewDefaultManager(
		func(mirrorProcessID uuid.UUID, msg string, keysAndValues ...interface{}) {
			srv.sugar.Infow(msg, keysAndValues...)

			tags := map[string]string{
				"recipientType": "mirror",
				"mirror-id":     mirrorProcessID.String(),
			}
			msg += strings.Repeat(", %s = %v", len(keysAndValues)/2)
			srv.logger.Infof(context.Background(), mirrorProcessID, tags, msg, keysAndValues...)
		},
		func(mirrorProcessID uuid.UUID, msg string, keysAndValues ...interface{}) {
			srv.sugar.Errorw(msg, keysAndValues...)

			tags := map[string]string{
				"recipientType": "mirror",
				"mirror-id":     mirrorProcessID.String(),
			}
			msg += strings.Repeat(", %s = %v", len(keysAndValues)/2)
			srv.logger.Errorf(context.Background(), mirrorProcessID, tags, msg, keysAndValues...)
		},
		noTx.DataStore().Mirror(),
		noTx.FileStore(),
		noTx.DataStore().RuntimeVariables(),
		&mirror.GitSource{},
		cc,
	)

	srv.sugar.Debug("Initializing functions grpc client.")
	functionsClientConn, err := util.GetEndpointTLS(srv.conf.FunctionsService + ":5555")
	if err != nil {
		srv.sugar.Error("initializing functions grpc client", "error", err)
		return err
	}
	srv.functionsClient = igrpc.NewFunctionsClient(functionsClientConn)

	if srv.conf.Eventing {
		srv.sugar.Debug("Initializing knative eventing receiver.")
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
			srv.sugar.Error(err)
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
			srv.sugar.Error(err)
			lock.Lock()
			if err == nil {
				err = e
			}
			lock.Unlock()
		}
	}()

	var node *cluster.Node
	// start pub sub
	config := cluster.DefaultConfig()
	node, err = cluster.NewNode(ctx, config, cluster.NewNodeFinderKube(), 100*time.Millisecond, srv.sugar.Named("cluster"))
	if err != nil {
		return err
	}
	srv.sugar.Info("Flow server started.")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// stop inidiviual components here
	go func(n *cluster.Node) {
		err = node.Stop()
		if err != nil {
			srv.sugar.Error("could not stop cluster node")
		}
	}(node)
	wg.Wait()

	closelogworker()

	if err != nil {
		return err
	}

	return nil
}

func (srv *server) cleanup(closer func() error) {
	err := closer()
	if err != nil {
		srv.sugar.Error(err)
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
		srv.sugar.Errorf("db notification failed: %v", perr)
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
	srv.pubsub.RegisterFunction(pubsub.PubsubConfigureRouterFunction, srv.flow.configureRouterHandler)
	srv.pubsub.RegisterFunction(pubsub.PubsubUpdateEventDelays, srv.events.updateEventDelaysHandler)

	srv.timers.registerFunction(timeoutFunction, srv.engine.timeoutHandler)
	srv.timers.registerFunction(sleepWakeupFunction, srv.engine.sleepWakeup)
	srv.timers.registerFunction(wfCron, srv.flow.cronHandler)
	srv.timers.registerFunction(sendEventFunction, srv.events.sendEvent)
	srv.timers.registerFunction(retryWakeupFunction, srv.flow.engine.retryWakeup)

	srv.pubsub.RegisterFunction(pubsub.PubsubDeleteActivityTimersFunction, srv.timers.deleteActivityTimersHandler)

	srv.pubsub.RegisterFunction(deleteFilterCache, srv.flow.deleteCache)
	srv.pubsub.RegisterFunction(deleteFilterCacheNamespace, srv.flow.deleteCacheNamespace)
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
		srv.sugar.Error(err)
		return
	}
	defer tx.Rollback()

	roots, err := tx.FileStore().GetAllRoots(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	for _, root := range roots {
		files, err := tx.FileStore().ForRootID(root.ID).ListAllFiles(ctx)
		if err != nil {
			srv.sugar.Error(err)
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
	ms, muxErr, err := srv.validateRouter(ctx, tx, file)
	if err != nil || muxErr != nil {
		return
	}

	if !ms.Enabled || ms.Cron != "" {
		srv.timers.deleteCronForWorkflow(file.ID.String())
	}

	if ms.Cron != "" && ms.Enabled {
		err := srv.timers.addCron(file.ID.String(), wfCron, ms.Cron, []byte(file.ID.String()))
		if err != nil {
			srv.sugar.Error(err)
			return
		}

		srv.sugar.Debugf("Loaded cron: %s", file.ID.String())
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
	res *gorm.DB
}

func (tx *sqlTx) FileStore() filestore.FileStore {
	return filestoresql.NewSQLFileStore(tx.res)
}

func (tx *sqlTx) DataStore() datastore.Store {
	return datastoresql.NewSQLStore(tx.res, os.Getenv(direktivSecretKey))
}

func (tx *sqlTx) InstanceStore() instancestore.Store {
	logger := zap.NewNop()
	return instancestoresql.NewSQLInstanceStore(tx.res, logger.Sugar())
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

func (srv *server) beginSqlTx(ctx context.Context) (*sqlTx, error) {
	res := srv.gormDB.WithContext(ctx).Begin()
	if res.Error != nil {
		return nil, res.Error
	}
	return &sqlTx{
		res: res,
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
