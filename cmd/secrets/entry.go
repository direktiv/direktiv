package secrets

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/direktiv/direktiv/pkg/secrets"
)

func RunApplication() {
	backend := "db"

	srv, err := secrets.NewServer(backend)
	if err != nil {
		log.Fatalf("can not run secrets: %v", err)
	}

	srv.Run()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.Signal(0xA))
		<-sig
		srv.Stop()
		<-sig
		srv.Kill()
	}()

	<-srv.Lifeline()
}
