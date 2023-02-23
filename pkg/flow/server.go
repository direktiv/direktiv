package flow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	libgrpc "google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/direktiv/direktiv/pkg/dlog"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/database/entwrapper"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/metrics"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/direktiv/direktiv/pkg/version"
	"github.com/lib/pq"
	_ "github.com/lib/pq" // postgres for ent
)

const parcelSize = 0x100000

type server struct {
	ID uuid.UUID

	sugar    *zap.SugaredLogger
	fnLogger *zap.SugaredLogger
	conf     *util.Config

	// db       *ent.Client
	pubsub   *pubsub
	locks    *locks
	timers   *timers
	engine   *engine
	syncer   *syncer
	secrets  *secrets
	flow     *flow
	internal *internal
	events   *events
	vars     *vars
	actions  *actions

	metrics *metrics.Client

	logQueue     chan *logMessage
	logWorkersWG sync.WaitGroup

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

	srv.logQueue = make(chan *logMessage, 1000)

	srv.initJQ()

	return srv, nil

}

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

	srv.sugar.Debug("Initializing secrets.")
	srv.secrets, err = initSecrets()
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.secrets.Close)

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

	srv.startLogWorkers(1)

	srv.sugar.Debug("Initializing pub-sub.")

	srv.pubsub, err = initPubSub(srv.sugar, srv, db)
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

	srv.sugar.Debug("Initializing syncer.")

	srv.syncer, err = initSyncer(srv)
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.syncer.Close)

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
	go srv.syncerCronPoller()

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

	srv.closeLogWorkers()

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

func (srv *server) notifyCluster(msg string) error {

	ctx := context.Background()

	conn, err := srv.edb.DB().Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "SELECT pg_notify($1, $2)", flowSync, msg)

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

func (srv *server) notifyHostname(hostname, msg string) error {

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

func (server *server) CacheNotify(req *PubsubUpdate) {

	if server.ID.String() == req.Sender {
		return
	}

	server.database.HandleNotification(req.Key)

}

func (srv *server) registerFunctions() {

	srv.pubsub.registerFunction(database.PubsubNotifyFunction, srv.CacheNotify)

	srv.pubsub.registerFunction(pubsubNotifyFunction, srv.pubsub.Notify)
	srv.pubsub.registerFunction(pubsubDisconnectFunction, srv.pubsub.Disconnect)
	srv.pubsub.registerFunction(pubsubDeleteTimerFunction, srv.timers.deleteTimerHandler)
	srv.pubsub.registerFunction(pubsubDeleteInstanceTimersFunction, srv.timers.deleteInstanceTimersHandler)
	srv.pubsub.registerFunction(pubsubCancelWorkflowFunction, srv.engine.finishCancelWorkflow)
	srv.pubsub.registerFunction(pubsubConfigureRouterFunction, srv.flow.configureRouterHandler)
	srv.pubsub.registerFunction(pubsubUpdateEventDelays, srv.events.updateEventDelaysHandler)

	srv.timers.registerFunction(timeoutFunction, srv.engine.timeoutHandler)
	srv.timers.registerFunction(sleepWakeupFunction, srv.engine.sleepWakeup)
	srv.timers.registerFunction(wfCron, srv.flow.cronHandler)
	srv.timers.registerFunction(sendEventFunction, srv.events.sendEvent)
	srv.timers.registerFunction(retryWakeupFunction, srv.flow.engine.retryWakeup)

	srv.pubsub.registerFunction(pubsubDeleteActivityTimersFunction, srv.timers.deleteActivityTimersHandler)
	srv.timers.registerFunction(syncerTimeoutFunction, srv.syncer.timeoutHandler)
	srv.timers.registerFunction(syncerCron, srv.syncer.cronHandler)

	srv.pubsub.registerFunction(deleteFilterCache, srv.flow.deleteCache)
	srv.pubsub.registerFunction(deleteFilterCacheNamespace, srv.flow.deleteCacheNamespace)

}

func (srv *server) cronPoller() {

	for {
		srv.cronPoll()
		time.Sleep(time.Minute * 15)
	}

}

func (srv *server) cronPoll() {

	ctx := context.Background()

	clients := srv.edb.Clients(nil)

	wfs, err := clients.Workflow.Query().All(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	for _, wf := range wfs {
		cached, err := srv.reverseTraverseToWorkflow(ctx, nil, wf.ID.String())
		if err != nil {
			srv.sugar.Error(err)
			continue
		}

		srv.cronPollerWorkflow(ctx, nil, cached)
	}

}

func (srv *server) cronPollerWorkflow(ctx context.Context, tx database.Transaction, cached *database.CacheData) {

	ms, muxErr, err := srv.validateRouter(ctx, tx, cached)
	if err != nil || muxErr != nil {
		return
	}

	if !ms.Enabled || ms.Cron != "" {
		srv.timers.deleteCronForWorkflow(cached.Workflow.ID.String())
	}

	if ms.Cron != "" && ms.Enabled {
		err = srv.timers.addCron(cached.Workflow.ID.String(), wfCron, ms.Cron, []byte(cached.Workflow.ID.String()))
		if err != nil {
			srv.sugar.Error(err)
			return
		}

		srv.sugar.Debugf("Loaded cron: %s", cached.Workflow.ID.String())

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
