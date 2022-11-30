package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/direktiv/direktiv/pkg/functions"
)

func main() {

	// start health check
	go startHealthHandler()

	// start server
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM|syscall.SIGINT)
	defer signal.Stop(interrupt)

	errChan := make(chan error)
	go functions.StartServer(errChan)

	select {
	case <-interrupt:
		break
	case e := <-errChan:
		log.Fatalf("can not start functions server: %v", e)
	}

	functions.StopServer()

}

func startHealthHandler() {

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "")
	})

	err := http.ListenAndServe("0.0.0.0:1234", nil)
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v.", err)
		}
	}

}
