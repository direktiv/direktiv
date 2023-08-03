package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/direktiv/direktiv/cmd/api"
	"github.com/direktiv/direktiv/cmd/flow"
	flow_dbinit "github.com/direktiv/direktiv/cmd/flow-dbinit"
	"github.com/direktiv/direktiv/cmd/functions"
	"github.com/direktiv/direktiv/cmd/sidecar"
)

func main() {
	appName := os.Getenv("DIREKTIV_APP")

	time.Local = time.UTC

	switch appName {
	case "api":
		api.RunApplication()
	case "sidecar":
		sidecar.RunApplication()
	case "flow":
		flow.RunApplication()
	case "flow_dbinit":
		flow_dbinit.RunApplication()
	case "functions":
		functions.RunApplication()
	case "":
		log.Fatalf("error: empty DIREKTIV_APP environment variable.\n")
	default:
		log.Fatalf(fmt.Sprintf("error: invalid DIREKTIV_APP environment variable value, got: '%s'.\n", appName))
	}
}
