package main

import (
	"fmt"
	"os"
	"time"

	"github.com/vorteil/direktiv/pkg/dlog"
	"go.uber.org/zap"
)

var logger *zap.Logger
var log *zap.SugaredLogger

func main() {

	var err error

	dlog.Init()

	logger, err = zap.NewDevelopment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	log = logger.Sugar()

	sl := new(SignalListener)
	sl.Start()

	/*
		telend, err := util.InitTelemetry(srv.conf, "direktiv", "direktiv/flow")
		if err != nil {
			return err
		}
		defer telend()
	*/

	local := new(LocalServer)
	local.Start()

	network := new(NetworkServer)
	network.local = local
	network.Start()

	threads.Wait()

	if code := threads.ExitStatus(); code != 0 {
		log.Errorf("Exiting with exit status: %d.", code)
		os.Exit(code)
	}

}

const (
	SUCCESS = 0
	ERROR   = 1
)

func Shutdown(code int) {
	t := time.Now()
	threads.Stop(&t, code)
}

func ForceQuit() {
	log.Warn("Performing force-quit.")
	os.Exit(1)
}
