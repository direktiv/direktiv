package api

import (
	"log"

	"github.com/direktiv/direktiv/pkg/dlog"
)

func RunApplication() {
	logger, err := dlog.ApplicationLogger("api")
	if err != nil {
		log.Fatalf("can not get logger: %v", err)
	}

	s, err := NewServer(logger)
	if err != nil {
		logger.Errorf("can not create api server: %v", err)
	}

	err = s.Start()
	if err != nil {
		logger.Errorf("can not start api server: %v", err)
		log.Fatal(err.Error())
	}
}
