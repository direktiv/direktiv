package main

import (
	"log"

	"github.com/vorteil/direktiv/pkg/api"
)

func main() {

	cfg, err := api.Configure()
	if err != nil {
		log.Fatalf(err.Error())
	}

	s, err := api.NewServer(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = s.Start()
	if err != nil {
		log.Fatal(err.Error())
	}

}
