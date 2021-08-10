package main

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	_ "github.com/vorteil/direktiv/pkg/util"
)

func main() {

	if os.Getenv("DIREKTIV_DEBUG") == "true" {
		log.SetLevel(logrus.DebugLevel)
	}

	sl := new(SignalListener)
	sl.Start()

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
