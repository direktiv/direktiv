package direktiv

import (
	"context"
	"net"
	"strings"

	"github.com/vorteil/direktiv/pkg/secrets"
	"google.golang.org/grpc"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // postgres for ent
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/dlog"
)

const (
	runsWorkflows = "w"
	runsIsolates  = "i"
	runsSecrets   = "s"
)

const (
	lockID   = 2610
	lockWait = 10
)

type component interface {
	start(s *WorkflowServer) error
	stop()
	name() string
}

// WorkflowServer is a direktiv server
type WorkflowServer struct {
	id         uuid.UUID
	ctx        context.Context
	config     *Config
	serverType string

	dbManager     *dbManager
	tmManager     *timerManager
	engine        *workflowEngine
	isolateServer *isolateServer

	LifeLine       chan bool
	instanceLogger dlog.Log
	secrets        secrets.SecretsServiceClient

	components map[string]component
}

func (s *WorkflowServer) initWorkflowServer() error {

	var err error

	// prep timers
	s.tmManager, err = newTimerManager(s)
	if err != nil {
		return err
	}

	s.engine, err = newWorkflowEngine(s)
	if err != nil {
		return err
	}

	// register the timer functions
	var timerFunctions = map[string]func([]byte) error{
		timerCleanOneShot:         s.tmManager.cleanOneShot,
		timerCleanInstanceRecords: s.tmManager.cleanInstanceRecords,
	}

	for n, f := range timerFunctions {
		err := s.tmManager.registerFunction(n, f)
		if err != nil {
			return err
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

	ingressServer := newIngressServer(s)
	s.components[ingressComponent] = ingressServer

	flowServer := newFlowServer(s.config, s.engine)
	s.components[flowComponent] = flowServer

	return nil

}

// NewWorkflowServer creates a new workflow server
func NewWorkflowServer(config *Config, serverType string) (*WorkflowServer, error) {

	log.Debugf("server type: %s", serverType)
	ctx := context.Background()

	var (
		err error
	)

	s := &WorkflowServer{
		id:         uuid.New(),
		ctx:        ctx,
		LifeLine:   make(chan bool),
		serverType: serverType,
		config:     config,
		components: make(map[string]component),
	}

	s.dbManager, err = newDBManager(ctx, s.config.Database.DB)
	if err != nil {
		return nil, err
	}

	if s.runsComponent(runsWorkflows) {
		err = s.initWorkflowServer()
		if err != nil {
			return nil, err
		}
	}

	if s.runsComponent(runsIsolates) {
		is, err := newActionManager(s.config, s.dbManager, &s.instanceLogger)
		if err != nil {
			return nil, err
		}
		s.isolateServer = is
		s.components[isolateComponent] = is
	}

	if s.runsComponent(runsSecrets) {
		secretsServer, err := newSecretsServer(config)
		if err != nil {
			log.Errorf("can not create secret server: %v", err)
			return nil, err
		}
		s.components[secretsComponent] = secretsServer
	}

	healthServer := newHealthServer(config)
	s.components[healthComponent] = healthServer

	return s, nil

}

func (s *WorkflowServer) runsComponent(c string) bool {
	if strings.Contains(s.serverType, c) {
		return true
	}
	return false
}

// SetInstanceLogger set logger for direktiv for firecracker instances
func (s *WorkflowServer) SetInstanceLogger(l dlog.Log) {
	s.instanceLogger = l
}

// Lifeline interface impl
func (s *WorkflowServer) Lifeline() chan bool {
	return s.LifeLine
}

func (s *WorkflowServer) cleanup() {

	// closing db at the end
	if s.dbManager != nil {
		defer s.dbManager.dbEnt.Close()
	}

	if s.tmManager != nil {
		s.tmManager.stopTimers()
	}

	// stop components
	for _, comp := range s.components {
		log.Debugf("stopping %s", comp.name())
		comp.stop()
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
	if s.runsComponent(runsWorkflows) {

		log.Debugf("subscribing to sync queue")
		err := s.startDatabaseListener()
		if err != nil {
			s.Kill()
			return err
		}

		// start timers
		err = s.tmManager.startTimers()
		if err != nil {
			s.Kill()
			return err
		}

	}

	for _, comp := range s.components {
		log.Debugf("starting %s component", comp.name())
		err := comp.start(s)
		if err != nil {
			s.Kill()
			return err
		}
	}

	return nil

}

func (s *WorkflowServer) grpcStart(server **grpc.Server, name, bind string, register func(srv *grpc.Server)) error {

	log.Debugf("%s endpoint starting at %s", name, bind)

	options, err := optionsForGRPC(s.config.Certs.Directory, secretsComponent, (s.config.Certs.Secure != 1))
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}

	(*server) = grpc.NewServer(options...)

	register(*server)

	go (*server).Serve(listener)

	return nil

}
