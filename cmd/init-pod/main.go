package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/direktiv/direktiv/pkg/dlog"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

var (
	namespace, actionId, instanceId string
	step                            int32

	flow grpc.InternalClient
)

func main() {

	var err error

	log, err = dlog.ApplicationLogger("init-pod")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	lifecycle := os.Getenv("DIREKTIV_LIFECYCLE")

	err = initialize()
	if err != nil {
		log.Infof("Error: %v", err)
		os.Exit(1)
	}

	conf := new(util.Config)
	conf.OpenTelemetryBackend = os.Getenv(util.DirektivOpentelemetry)

	telend, err := util.InitTelemetry(conf, "direktiv/init-pod", "direktiv")
	if err != nil {
		log.Infof("Error: %v", err)
		os.Exit(1)
	}
	defer telend()

	if lifecycle == "init" {
		runAsInit()
	} else if lifecycle == "run" {
		runAsSidecar()
	} else {
		log.Infof("Invalid DIREKTIV_LIFECYCLE: \"%s\"", lifecycle)
	}

}

func initialize() error {

	actionId = os.Getenv("DIREKTIV_ACTIONID")
	instanceId = os.Getenv("DIREKTIV_INSTANCEID")
	namespace = os.Getenv("DIREKTIV_NAMESPACE")

	/* #nosec */
	x, err := strconv.Atoi(os.Getenv("DIREKTIV_STEP"))
	if err != nil {
		return err
	}

	/* #nosec */
	step = int32(x)

	log.Infof("DIREKTIV_ACTIONID: %s", actionId)
	log.Infof("DIREKTIV_INSTANCEID: %s", instanceId)
	log.Infof("DIREKTIV_NAMESPACE: %s", namespace)
	log.Infof("DIREKTIV_STEP: %v", step)

	// "Direktiv-Deadline"

	err = initFlow()
	if err != nil {
		return err
	}

	return nil

}

func initFlow() error {

	conn, err := util.GetEndpointTLS(fmt.Sprintf("%s:7777",
		os.Getenv(util.DirektivFlowEndpoint)))
	if err != nil {
		return err
	}

	flow = grpc.NewInternalClient(conn)

	return nil

}
