package cmd

import (
	"github.com/direktiv/direktiv/pkg/refactor/function2"
	"github.com/direktiv/direktiv/pkg/refactor/webapi"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func NewMain() *sync.WaitGroup {
	funcManager, err := function2.NewManagerFromK8s()
	if err != nil {
		log.Fatalf("error creating functions client: %v\n", err)
	}

	wg := &sync.WaitGroup{}
	done := make(chan struct{})

	go func() {
		// Listen for syscall signals for process to interrupt/quit
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		close(done)
	}()

	// Start functions manager
	wg.Add(1)
	funcManager.Start(done, wg)

	// Start api v2 server
	wg.Add(1)
	webapi.Start(funcManager, "0.0.0.0:6667", done, wg)

	return wg
}
