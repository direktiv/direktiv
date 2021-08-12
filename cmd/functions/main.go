package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/functions"
	_ "github.com/vorteil/direktiv/pkg/util"
)

func main() {

	formatter := runtime.Formatter{ChildFormatter: &log.TextFormatter{
		FullTimestamp: true,
	}}
	formatter.Line = true
	log.SetFormatter(&formatter)

	log.Infof("starting functions server")

	if os.Getenv("DIREKTIV_DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}

	// start health check
	go startHealthHandler()

	// start server
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM|syscall.SIGINT)
	defer signal.Stop(interrupt)

	errChan := make(chan error)
	go functions.StartServer(errChan)

	log.Infof("functions server started")

	select {
	case <-interrupt:
		break
	case e := <-errChan:
		log.Errorf("can not start functions server: %v", e)
		os.Exit(1)
	}

	log.Infof("stopping functions grpc server")
	functions.StopServer()
	log.Infof("functions grpc server stopped")

}

func startHealthHandler() {

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "")
	})

	log.Printf("health service started")
	http.ListenAndServe("0.0.0.0:1234", nil)

}
