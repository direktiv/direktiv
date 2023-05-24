package flow

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/dlog"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/flow/pubsub"
	"github.com/direktiv/direktiv/pkg/metrics"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/filestoresql"
	"github.com/direktiv/direktiv/pkg/refactor/logengine"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/direktiv/direktiv/pkg/version"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq" // postgres for ent
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	libgrpc "google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
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
	// fStore    filestore.FileStore
	// dataStore datastore.Store

	mirrorManager mirror.Manager

	flow     *flow
	internal *internal
	events   *events
	vars     *vars
	actions  *actions

	metrics  *metrics.Client
	logger   logengine.BetterLogger
	edb      *entwrapper.Database // TODO: remove
	database *database.CachedDatabase
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

//nolint:gocyclo
func (srv *server) start(ctx context.Context) error {
	var err error

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

	// srv.db, err = initDatabase(ctx, db)
	// if err != nil {
	// 	return err
	// }
	// defer srv.cleanup(srv.db.Close)

	edb, err := entwrapper.New(ctx, srv.sugar, db)
	if err != nil {
		return err
	}
	srv.edb = edb

	srv.database = database.NewCachedDatabase(srv.sugar, edb, srv)
	defer srv.cleanup(srv.database.Close)

	// fmt.Printf(">>>>>> dsn %s\n", db)

	srv.gormDB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  db,
		PreferSimpleProtocol: false, // disables implicit prepared statement usage
		// Conn:                 edb.DB(),
	}), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel: logger.Info,
			},
		),
	})
	if err != nil {
		return fmt.Errorf("creating filestore, err: %w", err)
	}

	gdb, err := srv.gormDB.DB()
	if err != nil {
		return fmt.Errorf("modifying gorm driver, err: %w", err)
	}
	gdb.SetMaxIdleConns(32)
	gdb.SetMaxOpenConns(16)
	// srv.fStore = psql.NewSQLFileStore(srv.gormDB)
	// srv.dataStore = sql.NewSQLStore(srv.gormDB)

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

	srv.metrics, err = metrics.NewClient()
	if err != nil {
		return err
	}

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

	srv.sugar.Debug("Initializing vars server.")

	srv.vars, err = initVarsServer(cctx, srv)
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.vars.Close)

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
	store := datastoresql.NewSQLStore(srv.gormDB, os.Getenv(direktivSecretKey))
	fStore := filestoresql.NewSQLFileStore(srv.gormDB)
	srv.logger = logengine.ChainedBetterLogger{
		logengine.SugarBetterLogger{
			Sugar: srv.sugar,
			AddTraceFrom: func(ctx context.Context, toTags map[string]interface{}) map[string]interface{} {
				span := trace.SpanFromContext(ctx)
				tid := span.SpanContext().TraceID()
				toTags["trace"] = tid
				return toTags
			},
		},
		logengine.DataStoreBetterLogger{Store: store.Logs(), LogError: srv.sugar.Errorf},
		logengine.NotifierBetterLogger{Callback: func(objectID uuid.UUID, objectType string) {
			srv.pubsub.NotifyLogs(objectID, recipient.RecipientType(objectType))
		}, LogError: srv.sugar.Errorf},
	}

	cc := func(ctx context.Context, file *filestore.File) error {
		_, router, err := getRouter(ctx, fStore, store.FileAnnotations(), file)
		if err != nil {
			return err
		}

		err = srv.flow.configureWorkflowStarts(ctx, fStore, store.FileAnnotations(), file.RootID, file, router, false)
		if err != nil {
			return err
		}

		return nil
	}

	srv.mirrorManager = mirror.NewDefaultManager(
		func(mirrorProcessID uuid.UUID, msg string, keysAndValues ...interface{}) {
			srv.sugar.Infow(msg, keysAndValues...)

			tags := map[string]interface{}{
				"recipientType": "mirror",
				"mirror-id":     mirrorProcessID.String(),
			}
			msg += strings.Repeat(", %s = %v", len(keysAndValues)/2)
			srv.logger.Infof(context.Background(), mirrorProcessID, tags, msg, keysAndValues...)
		},
		func(mirrorProcessID uuid.UUID, msg string, keysAndValues ...interface{}) {
			srv.sugar.Errorw(msg, keysAndValues...)

			tags := map[string]interface{}{
				"recipientType": "mirror",
				"mirror-id":     mirrorProcessID.String(),
			}
			msg += strings.Repeat(", %s = %v", len(keysAndValues)/2)
			srv.logger.Errorf(context.Background(), mirrorProcessID, tags, msg, keysAndValues...)
		},
		store.Mirror(),
		fStore,
		&mirror.GitSource{},
		cc,
	)

	srv.sugar.Debug("Initializing actions grpc server.")

	srv.actions, err = initActionsServer(cctx, srv)
	if err != nil {
		return err
	}

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
		e := srv.vars.Run()
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

	go func() {
		defer wg.Done()
		defer cancel()
		e := srv.actions.Run()
		if e != nil {
			srv.sugar.Error(err)
			lock.Lock()
			if err == nil {
				err = e
			}
			lock.Unlock()
		}
	}()

	srv.sugar.Info("Flow server started.")

	wg.Wait()

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

	conn, err := srv.edb.DB().Conn(ctx)
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

	conn, err := srv.edb.DB().Conn(ctx)
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

func (server *server) CacheNotify(req *pubsub.PubsubUpdate) {
	if server.ID.String() == req.Sender {
		return
	}

	server.database.HandleNotification(req.Key)
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

	fStore, store, _, rollback, err := srv.flow.beginSqlTx(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}
	defer rollback()

	roots, err := fStore.GetAllRoots(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	for _, root := range roots {
		files, err := fStore.ForRootID(root.ID).ListAllFiles(ctx)
		if err != nil {
			srv.sugar.Error(err)
			return
		}

		for _, file := range files {
			if file.Typ != filestore.FileTypeWorkflow {
				continue
			}

			srv.cronPollerWorkflow(ctx, fStore, store.FileAnnotations(), file)
		}
	}
}

func (srv *server) cronPollerWorkflow(ctx context.Context, fStore filestore.FileStore, store core.FileAnnotationsStore, file *filestore.File) {
	ms, muxErr, err := srv.validateRouter(ctx, fStore, store, file)
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

func (flow *flow) Build(ctx context.Context, in *emptypb.Empty) (*grpc.BuildResponse, error) {
	var resp grpc.BuildResponse
	resp.Build = version.Version
	return &resp, nil
}

func (engine *engine) UserLog(ctx context.Context, im *instanceMemory, msg string, a ...interface{}) {
	engine.logger.Infof(ctx, im.GetInstanceID(), im.GetAttributes(), msg, a...)

	if attr := im.runtime.LogToEvents; attr != "" {
		s := fmt.Sprintf(msg, a...)
		event := cloudevents.NewEvent()
		event.SetID(uuid.New().String())
		event.SetSource(im.cached.File.ID.String())
		event.SetType("direktiv.instanceLog")
		event.SetExtension("logger", attr)
		event.SetDataContentType("application/json")
		err := event.SetData("application/json", s)
		if err != nil {
			engine.sugar.Errorf("Failed to create CloudEvent: %v.", err)
		}

		err = engine.events.BroadcastCloudevent(ctx, im.cached.Namespace, &event, 0)
		if err != nil {
			engine.sugar.Errorf("Failed to broadcast CloudEvent: %v.", err)
			return
		}
	}
}

func (engine *engine) logRunState(ctx context.Context, im *instanceMemory, wakedata []byte, err error) {
	engine.sugar.Debugf("Running state logic -- %s:%v (%s) (%v)", im.ID().String(), im.Step(), im.logic.GetID(), time.Now())
	if im.GetMemory() == nil && len(wakedata) == 0 && err == nil {
		engine.logger.Infof(ctx, im.GetInstanceID(), im.GetAttributes(), "Running state logic (step:%v) -- %s", im.Step(), im.logic.GetID())
	}
}

func this() string {
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	elems := strings.Split(fn.Name(), ".")
	return elems[len(elems)-1]
}

func parent() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return ""
	}
	fn := runtime.FuncForPC(pc)
	elems := strings.Split(fn.Name(), ".")
	return elems[len(elems)-1]
}

func (flow *flow) beginSqlTx(ctx context.Context) (filestore.FileStore, datastore.Store, func(ctx context.Context) error, func(), error) {
	res := flow.gormDB.WithContext(ctx).Begin()
	if res.Error != nil {
		return nil, nil, nil, nil, res.Error
	}
	rollbackFunc := func() {
		err := res.Rollback().Error
		if err != nil {
			if !strings.Contains(err.Error(), "already") {
				fmt.Fprintf(os.Stderr, "failed to rollback transaction: %v\n", err)
			}
		}
	}
	commitFunc := func(ctx context.Context) error {
		return res.WithContext(ctx).Commit().Error
	}

	return filestoresql.NewSQLFileStore(res), datastoresql.NewSQLStore(res, os.Getenv(direktivSecretKey)), commitFunc, rollbackFunc, nil
}

func (flow *flow) runSqlTx(ctx context.Context, fun func(fStore filestore.FileStore, store datastore.Store) error) error {
	fStore, store, commit, rollback, err := flow.beginSqlTx(ctx)
	if err != nil {
		return err
	}
	defer rollback()

	if err := fun(fStore, store); err != nil {
		return err
	}

	return commit(ctx)
}
