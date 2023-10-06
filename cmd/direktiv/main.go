package main

import (
	"fmt"
	"log"
	"os"

	"github.com/direktiv/direktiv/cmd/api"
	"github.com/direktiv/direktiv/cmd/flow"
	"github.com/direktiv/direktiv/cmd/sidecar"
)

func main() {
	appName := os.Getenv("DIREKTIV_APP")

	switch appName {
	case "api":
		api.RunApplication()
	case "sidecar":
		sidecar.RunApplication()
	case "flow":
		flow.RunApplication()
	case "":
		log.Fatalf("error: empty DIREKTIV_APP environment variable.\n")
	default:
		log.Fatalf(fmt.Sprintf("error: invalid DIREKTIV_APP environment variable value, got: '%s'.\n", appName))
	}
}
