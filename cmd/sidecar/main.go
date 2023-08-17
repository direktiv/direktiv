package sidecar

import (
	"fmt"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/dlog"
	"github.com/direktiv/direktiv/pkg/util"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func RunApplication() {
	var err error

	log, err = dlog.ApplicationLogger("sidecar")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		err := log.Sync()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to sync logs: %v\n", err)
		}
	}()

	sl := new(SignalListener)
	sl.Start()

	conf := new(util.Config)
	conf.OpenTelemetryBackend = os.Getenv(util.DirektivOpentelemetry)

	telend, err := util.InitTelemetry(conf, "direktiv/sidecar", "direktiv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize telemetry: %v\n", err)
		os.Exit(1)
	}
	defer telend()

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
	t := time.Now().UTC()
	threads.Stop(&t, code)
}

func ForceQuit() {
	log.Warn("Performing force-quit.")
	os.Exit(1)
}
