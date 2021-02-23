package direktiv

import (
	"context"
	"os"
	"time"

	"github.com/vorteil/direktiv/pkg/flow"
	"github.com/vorteil/direktiv/pkg/health"
	"github.com/vorteil/direktiv/pkg/secrets"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // postgres for ent
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/dlog"
	"github.com/vorteil/direktiv/pkg/ingress"
	"google.golang.org/grpc"
)

const (
	workflowOnly   = "wfo"
	workflowRunner = "wfr"
	workflowAll    = "wf"
)

const (
	lockID   = 2610
	lockWait = 10
)

type subServer interface {
	start() error
	stop()
	name() string
}

// WorkflowServer is a direktiv server
type WorkflowServer struct {
	flow.UnimplementedDirektivFlowServer
	health.UnimplementedHealthServer
	ingress.UnimplementedDirektivIngressServer

	id         uuid.UUID
	ctx        context.Context
	config     *Config
	serverType string

	dbManager      *dbManager
	tmManager      *timerManager
	engine         *workflowEngine
	actionManager  *actionManager
	LifeLine       chan bool
	instanceLogger dlog.Log

	grpcFlow    *grpc.Server
	grpcHealth  *grpc.Server
	grpcIngress *grpc.Server

	secrets secrets.SecretsServiceClient
}

// NewWorkflowServer creates a new workflow server
func NewWorkflowServer(config *Config, serverType string) (*WorkflowServer, error) {

	log.Debugf("server type: %s", serverType)

	var err error
	ctx := context.Background()

	s := &WorkflowServer{
		id:         uuid.New(),
		ctx:        ctx,
		LifeLine:   make(chan bool),
		serverType: serverType,
	}

	s.config = config

	s.dbManager, err = newDBManager(ctx, s.config.Database.DB)
	if err != nil {
		return nil, err
	}

	if s.isWorkflowServer() {

		// prep timers
		s.tmManager, err = newTimerManager(s)
		if err != nil {
			return nil, err
		}

		s.grpcFlowStart()

		s.engine, err = newWorkflowEngine(s)
		if err != nil {
			return nil, err
		}

		// register the timer functions
		var (
			timerFunctions = map[string]func([]byte) error{
				// timerCleanMessageID: s.dbManager.messageIDCleaner,
				timerCleanOneShot:         s.tmManager.cleanOneShot,
				timerCleanInstanceRecords: s.tmManager.cleanInstanceRecords,
			}
		)

		for n, f := range timerFunctions {
			err := s.tmManager.registerFunction(n, f)
			if err != nil {
				return nil, err
			}
		}

		addCron := func(name, cron string) {
			// add clean up timers
			_, err := s.dbManager.getTimer(name)

			// on error we assuming it is not in the database
			if err != nil {
				s.tmManager.addCron(name, name, cron, []byte(""))
			}

		}

		addCron(timerCleanOneShot, "*/10 * * * *")
		addCron(timerCleanInstanceRecords, "0 * * * *")

		hn, _ := os.Hostname()

		if hn == "server1" {
			log.Debugf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!HOSTNAME %v", hn)
			go func() {
				log.Debugf("START!!!")
				time.Sleep(3 * time.Second)
				s.tmManager.addOneShot("this1", timerCleanOneShot, time.Now().Add(30*time.Second), []byte("this1"))
				// _, _ = s.tmManager.addCron("this1", timerCleanOneShot, "*/1 * * * *", []byte("this1"))
				time.Sleep(5 * time.Second)
				s.tmManager.actionTimerByName("this1", disableTimerAction)
				// time.Sleep(5 * time.Second)
				// s.tmManager.actionTimerByName("this1", enableTimerAction)
				time.Sleep(1 * time.Minute)
				s.tmManager.actionTimerByName("this1", deleteTimerAction)
			}()

		}

		// s.tmManager.disableTimer(h, false)
		// s.tmManager.enableTimer(h)
		// s.tmManager.addOneShot("1234", "cleanOneShot", time.Now().UTC().Add(6*time.Second), []byte("JENS2"))
		// s.tmManager.addOneShot("12345", "cleanOneShot", time.Now().UTC().Add(7*time.Second), []byte("JENS3"))

		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())

		conn, err := grpc.Dial(s.config.SecretsAPI.Endpoint, opts...)
		if err != nil {
			return nil, err
		}
		s.secrets = secrets.NewSecretsServiceClient(conn)

	}

	if s.isRunnerServer() {

		am, err := newActionManager(s.config, s.dbManager, &s.instanceLogger)
		if err != nil {
			return nil, err
		}
		s.actionManager = am
		am.grpcIsolateStart()

	}

	s.grpcIngressStart()
	s.grpcHealthStart()

	var subs []subServer

	sec, err := newSecretsServer(config)
	if err != nil {
		log.Errorf("can not create secret server: %v", err)
		return nil, err
	}

	subs = append(subs, sec)

	for _, sub := range subs {
		log.Debugf("starting %s", sub.name())
		err := sub.start()
		if err != nil {
			log.Errorf("can not start")
			return nil, err
		}
	}

	return s, nil

}

func (s *WorkflowServer) isWorkflowServer() bool {
	if s.serverType == workflowOnly || s.serverType == workflowAll {
		return true
	}
	return false
}

func (s *WorkflowServer) isRunnerServer() bool {
	if s.serverType == workflowRunner || s.serverType == workflowAll {
		return true
	}
	return false
}

func (s *WorkflowServer) SetInstanceLogger(l dlog.Log) {
	s.instanceLogger = l
}

// Lifeline interface impl
func (s *WorkflowServer) Lifeline() chan bool {
	return s.LifeLine
}

func (s *WorkflowServer) cleanup() {

	// closeing db at the end
	if s.dbManager != nil {
		defer s.dbManager.dbEnt.Close()
	}

	if s.isWorkflowServer() {
		s.tmManager.stopTimers()
	}

	if s.actionManager != nil {
		s.actionManager.stop()
	}

}

// Stop stops the server gracefully
func (s *WorkflowServer) Stop() {

	go func() {

		log.Printf("stopping workflow server")
		s.cleanup()
		s.LifeLine <- true

	}()
}

// Kill kills the server
func (s *WorkflowServer) Kill() {

	go func() {

		defer func() {
			_ = recover()
		}()

		s.cleanup()
		s.LifeLine <- true

	}()

}

// Run starts all components of direktiv
func (s *WorkflowServer) Run() error {

	// subscribe to cmds
	if s.isWorkflowServer() {

		log.Debugf("subscribing to sync queue")
		err := s.startDatabaseListener()
		if err != nil {
			return err
		}

		// start timers
		err = s.tmManager.startTimers()
		if err != nil {
			return err
		}

	}

	if s.actionManager != nil {
		err := s.actionManager.start()
		if err != nil {
			return err
		}
	}

	return nil

}
