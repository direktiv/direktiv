package main

import (
	"os"
	"os/signal"
	"syscall"

	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	log "github.com/sirupsen/logrus"

	"github.com/vorteil/direktiv/pkg/secrets"
	_ "github.com/vorteil/direktiv/pkg/util"
)

func main() {

	backend := "db"

	srv, err := secrets.NewServer(backend)
	if err != nil {
		log.Errorf("can not run secrets: %v", err)
		os.Exit(1)
	}

	if os.Getenv("DIREKTIV_DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
		formatter := runtime.Formatter{ChildFormatter: &log.TextFormatter{
			FullTimestamp: true,
		}}
		formatter.Line = true
		log.SetFormatter(&formatter)
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
