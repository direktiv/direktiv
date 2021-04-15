package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/vorteil/direktiv/pkg/secrets"
)

func main() {

	if len(os.Args) != 2 {
		log.Fatal("secrets needs type, e.g. database, vault")
	}

	backend := os.Args[1]

	srv, err := secrets.NewServer(backend)
	if err != nil {
		log.Errorf("can not run secrets: %v", err)
		os.Exit(1)
	}

	srv.Run()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
		<-sig
		srv.Stop()
		<-sig
		srv.Kill()
	}()

	<-srv.Lifeline()

	log.Infof("secrets server stopped")
}
