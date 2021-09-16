package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vorteil/direktiv/pkg/dlog"
	"github.com/vorteil/direktiv/pkg/secrets"
	"github.com/vorteil/direktiv/pkg/util"
)

func main() {

	dlog.Init()
	util.Init()

	backend := "db"

	srv, err := secrets.NewServer(backend)
	if err != nil {
		log.Fatalf("can not run secrets: %v", err)
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

}
