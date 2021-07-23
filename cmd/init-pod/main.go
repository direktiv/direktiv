package main

import (
	"log"
	"os"
	"strconv"

	"github.com/vorteil/direktiv/pkg/direktiv"
	"github.com/vorteil/direktiv/pkg/flow"
)

var (
	namespace, actionId, instanceId string
	step                            int32

	flowClient flow.DirektivFlowClient
)

func main() {

	lifecycle := os.Getenv("DIREKTIV_LIFECYCLE")

	err := initialize()
	if err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}

	if lifecycle == "init" {
		runAsInit()
	} else if lifecycle == "run" {
		runAsSidecar()
	} else {
		log.Printf("Invalid DIREKTIV_LIFECYCLE: \"%s\"", lifecycle)
	}

}

func initialize() error {

	actionId = os.Getenv("DIREKTIV_ACTIONID")
	instanceId = os.Getenv("DIREKTIV_INSTANCEID")
	namespace = os.Getenv("DIREKTIV_NAMESPACE")
	x, err := strconv.Atoi(os.Getenv("DIREKTIV_STEP"))
	if err != nil {
		return err
	}

	step = int32(x)

	log.Printf("DIREKTIV_ACTIONID: %s", actionId)
	log.Printf("DIREKTIV_INSTANCEID: %s", instanceId)
	log.Printf("DIREKTIV_NAMESPACE: %s", namespace)
	log.Printf("DIREKTIV_STEP: %v", step)

	// "Direktiv-Deadline"

	err = initFlow()
	if err != nil {
		return err
	}

	return nil

}

func initFlow() error {

	flowAddr := os.Getenv("DIREKTIV_FLOW_ENDPOINT")

	log.Printf("Connecting to flow: %s.", flowAddr)

	conn, err := direktiv.GetEndpointTLS(flowAddr, true)
	if err != nil {
		return err
	}

	flowClient = flow.NewDirektivFlowClient(conn)

	return nil

}
