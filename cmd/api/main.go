package main

import (
	"log"

	"github.com/vorteil/direktiv/pkg/api"
	"github.com/vorteil/direktiv/pkg/dlog"
)

func main() {

	logger, err := dlog.ApplicationLogger("api")
	if err != nil {
		log.Fatalf("can not get logger: %v", err)
	}

	s, err := api.NewServer(logger)
	if err != nil {
		logger.Errorf("can not create api server: %v", err)
	}

	err = s.Start()
	if err != nil {
		logger.Errorf("can not start api server: %v", err)
		log.Fatal(err.Error())
	}

}
