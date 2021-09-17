package flow

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"gopkg.in/yaml.v2"

	"github.com/lib/pq"
	_ "github.com/lib/pq" // postgres for ent
	"github.com/vorteil/direktiv/pkg/dlog"
	"github.com/vorteil/direktiv/pkg/flow/ent"
	"github.com/vorteil/direktiv/pkg/metrics"
	"github.com/vorteil/direktiv/pkg/util"
)

const parcelSize = 0x100000

type Config struct {
	Database          string `yaml:"database"`
	BindFlow          string `yaml:"bind_flow"`
	BindInternal      string `yaml:"bind_internal"`
	FunctionsProtocol string `yaml:"functions-protocol"`
}

func ReadConfig(file string) (*Config, error) {

	c := new(Config)

	/* #nosec */
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}

	s := os.Getenv("DATABASE")
	if s != "" {
		c.Database = s
	}

	s = os.Getenv("BIND_FLOW")
	if s != "" {
		c.BindFlow = s
	}

	s = os.Getenv("BIND_INTERNAL")
	if s != "" {
		c.BindInternal = s
	}

	s = os.Getenv("FUNCTIONS_PROTOCOL")
	if s != "" {
		c.FunctionsProtocol = s
	}

	return c, nil

}

type server struct {
	ID uuid.UUID

	logger *zap.Logger
	sugar  *zap.SugaredLogger
	conf   *Config

	db       *ent.Client
	pubsub   *pubsub
	locks    *locks
	timers   *timers
	engine   *engine
	secrets  *secrets
	flow     *flow
	internal *internal
	events   *events

	metrics *metrics.Client
}

func Run(ctx context.Context, logger *zap.Logger, conf *Config) error {

	dlog.Init()
	util.Init()

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

func newServer(logger *zap.Logger, conf *Config) (*server, error) {

	srv := new(server)
	srv.ID = uuid.New()

	srv.logger = logger
	srv.sugar = logger.Sugar()
	srv.conf = conf

	srv.initJQ()

	return srv, nil

}

func (srv *server) start(ctx context.Context) error {

	var err error

	// go setupPrometheusEndpoint()

	srv.sugar.Debug("Initializing secrets.")
	srv.secrets, err = initSecrets()
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.secrets.Close)

	srv.sugar.Debug("Initializing locks.")

	srv.locks, err = initLocks(srv.conf.Database)
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.locks.Close)

	srv.sugar.Debug("Initializing pub-sub.")

	srv.pubsub, err = initPubSub(srv.sugar, srv, srv.conf.Database)
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.pubsub.Close)

	srv.sugar.Debug("Initializing database.")

	srv.db, err = initDatabase(ctx, srv.conf.Database)
	if err != nil {
		return err
	}
	defer srv.cleanup(srv.db.Close)

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

	wg.Add(2)

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

func (srv *server) notifyCluster(msg string) error {

	ctx := context.Background()

	conn, err := srv.db.DB().Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(ctx, "SELECT pg_notify($1, $2)", flowSync, msg)
	if err, ok := err.(*pq.Error); ok {

		fmt.Fprintf(os.Stderr, "db notification failed: %v", err)
		if err.Code == "57014" {
			return fmt.Errorf("canceled query")
		}

		return err

	}

	return err

}

func (srv *server) notifyHostname(hostname, msg string) error {

	ctx := context.Background()

	conn, err := srv.db.DB().Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	channel := fmt.Sprintf("hostname:%s", hostname)

	_, err = conn.ExecContext(ctx, "SELECT pg_notify($1, $2)", channel, msg)
	if err, ok := err.(*pq.Error); ok {

		fmt.Fprintf(os.Stderr, "db notification failed: %v", err)
		if err.Code == "57014" {
			return fmt.Errorf("canceled query")
		}

		return err

	}

	return err

}

func (srv *server) registerFunctions() {

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

}

func (srv *server) cronPoller() {

	for {
		srv.cronPoll()
		time.Sleep(time.Minute * 15)
	}

}

func (srv *server) cronPoll() {

	ctx := context.Background()

	wfs, err := srv.db.Workflow.Query().All(ctx)
	if err != nil {
		srv.sugar.Error(err)
		return
	}

	for _, wf := range wfs {
		srv.cronPollerWorkflow(wf)
	}

}

func (srv *server) cronPollerWorkflow(wf *ent.Workflow) {

	ctx := context.Background()

	ms, muxErr, err := validateRouter(ctx, wf)
	if err != nil || muxErr != nil {
		return
	}

	if !ms.Enabled || ms.Cron != "" {
		srv.timers.deleteCronForWorkflow(wf.ID.String())
	}

	if ms.Cron != "" && ms.Enabled {

		err = srv.timers.addCron(fmt.Sprintf("cron:%s", wf.ID.String()), wfCron, ms.Cron, []byte(wf.ID.String()))
		if err != nil {
			srv.sugar.Error(err)
			return
		}

		srv.sugar.Debugf("Loaded cron: %s", wf.ID.String())

	}

}
