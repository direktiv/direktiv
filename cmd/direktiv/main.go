package main

import (
	"fmt"
	"github.com/direktiv/direktiv/cmd/api"
	"github.com/direktiv/direktiv/cmd/secrets"
	"github.com/direktiv/direktiv/cmd/sidecar"
	"log"
	"os"
)

func main() {
	appName := os.Getenv("DIREKTIV_APP")

	switch appName {
	case "api":
		api.RunApplication()
	case "secrets":
		secrets.RunApplication()
	case "sidecar":
		sidecar.RunApplication()
	case "":
		log.Fatalf("error: empty DIREKTIV_APP environment variable.\n")
	default:
		log.Fatalf(fmt.Sprintf("error: invalid DIREKTIV_APP environment variable value, got: '%s'.\n", appName))
	}
}
