package direktiv

import (
	"context"
	"log"
	"os"

	"github.com/vorteil/direktiv/pkg/jqer"
	"github.com/vorteil/direktiv/pkg/varstore"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // postgres for ent
	"github.com/vorteil/direktiv/pkg/dlog"
	"github.com/vorteil/direktiv/pkg/util"
	"go.uber.org/zap"
)

var (
	appLog, fnLog *zap.SugaredLogger
)

const (
	runsWorkflows = "w"
	runsSecrets   = "s"

	defaultLockWait = 10

	secretsEndpoint = "127.0.0.1:2610"
)

type component interface {
	start(s *WorkflowServer) error
	stop()
	name() string
}

// WorkflowServer is a direktiv server
type WorkflowServer struct {
	id     uuid.UUID
	ctx    context.Context
	config *Config

	dbManager *dbManager
	tmManager *timerManager
	engine    *workflowEngine

	LifeLine        chan bool
	instanceLogger  dlog.Log
	variableStorage varstore.VarStorage

	components map[string]component
	hostname   string
}

func (s *WorkflowServer) GetConfig() Config {
	return *s.config
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
		timerCleanInstanceRecords:  s.tmManager.cleanInstanceRecords,
		timerCleanNamespaceRecords: s.tmManager.cleanNamespaceRecords,
	}

	for n, f := range timerFunctions {
		err := s.tmManager.registerFunction(n, f)
		if err != nil {
			return err
		}
	}

	addCron := func(name, cron string) {
		s.tmManager.addCronNoBroadcast(name, name, cron, []byte(""))
	}

	addCron(timerCleanInstanceRecords, "0 * * * *")

	addCron(timerCleanNamespaceRecords, "0 */2 * * *")

	ingressServer, err := newIngressServer(s)
	if err != nil {
		return err
	}

	s.components[util.IngressComponent] = ingressServer

	flowServer := newFlowServer(s.config, s.engine)
	s.components[util.FlowComponent] = flowServer

	s.components[util.LogComponent] = newLogDBClient()

	return nil

}

func init() {

	// setup logging
	var err error

	appLog, err = dlog.ApplicationLogger("flow")
	if err != nil {
		log.Fatalf(err.Error())
	}

	fnLog, err = dlog.FunctionsLogger()
	if err != nil {
		log.Fatalf(err.Error())
	}

}

// NewWorkflowServer creates a new workflow server
func NewWorkflowServer(config *Config) (*WorkflowServer, error) {

	ctx := context.Background()

	jqer.StringQueryRequiresWrappings = true
	jqer.TrimWhitespaceOnQueryStrings = true
	jqer.SearchInStrings = true
	jqer.WrappingBegin = "jq"
	jqer.WrappingIncrement = "("
	jqer.WrappingDecrement = ")"

	var (
		err error
	)

	s := &WorkflowServer{
		id:         uuid.New(),
		ctx:        ctx,
		LifeLine:   make(chan bool),
		config:     config,
		components: make(map[string]component),
	}

	// not needed for secrets
	s.dbManager, err = newDBManager(ctx, s.config.Database.DB, config)
	if err != nil {
		return nil, err
	}
	s.dbManager.varStorage = &s.variableStorage

	err = s.initWorkflowServer()
	if err != nil {
		return nil, err
	}
	s.dbManager.tm = s.tmManager

	hn, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	s.hostname = hn

	return s, nil

}

// SetInstanceLogger set logger for direktiv for firecracker instances
func (s *WorkflowServer) SetInstanceLogger(l dlog.Log) {
	s.instanceLogger = l
}

func (s *WorkflowServer) SetVariableStorage(vs varstore.VarStorage) {
	s.variableStorage = vs
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
		appLog.Infof("stopping %s", comp.name())
		comp.stop()
	}

	if s.dbManager.grpcConn != nil {
		s.dbManager.grpcConn.Close()
	}

}

// Stop stops the server gracefully
func (s *WorkflowServer) Stop() {

	go func() {

		appLog.Info("stopping workflow server")
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

	appLog.Debug("subscribing to sync queue")
	err := s.startDatabaseListener()
	if err != nil {
		s.Kill()
		return err
	}

	for _, comp := range s.components {
		appLog.Infof("starting %s component", comp.name())
		err := comp.start(s)
		if err != nil {
			s.Kill()
			return err
		}
	}

	return nil

}
