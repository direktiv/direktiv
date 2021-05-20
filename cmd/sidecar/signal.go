package main

import (
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type SignalListener struct {
	signals chan os.Signal
	stopper chan *time.Time
}

func (sl *SignalListener) Start() {

	sl.signals = make(chan os.Signal, 2)
	sl.stopper = make(chan *time.Time, 1)
	signal.Notify(sl.signals, os.Interrupt, unix.SIGTERM)

	log.Debug("Listening for signals.")

	end := threads.Register(sl.stopper)

	go sl.listen(end)

}

func (sl *SignalListener) listen(end func()) {

	defer end()

	select {
	case <-sl.signals:
		log.Info("Received shutdown signal.")
		Shutdown(SUCCESS)
	case <-sl.stopper:
		log.Debug("Stopping signal listener.")
	}

	go func() {

		<-time.After(time.Second * 20)

		ForceQuit()

	}()

}
