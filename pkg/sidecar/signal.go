package sidecar

import (
	"log/slog"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sys/unix"
)

// SignalListener controls the signals for a sidecar.
type SignalListener struct {
	signals chan os.Signal
	stopper chan *time.Time
}

// Start starts listening for signals.
func (sl *SignalListener) Start() {
	sl.signals = make(chan os.Signal, 2)
	sl.stopper = make(chan *time.Time, 1)
	signal.Notify(sl.signals, os.Interrupt, unix.SIGTERM)

	slog.Debug("Listening for signals.")

	end := threads.Register(sl.stopper)

	go sl.listen(end)
}

func (sl *SignalListener) listen(end func()) {
	defer end()

	select {
	case <-sl.signals:
		slog.Info("Received shutdown signal.")
		Shutdown(SUCCESS)
	case <-sl.stopper:
		slog.Debug("Stopping signal listener.")
	}

	go func() {
		<-time.After(time.Second * 20)

		ForceQuit()
	}()
}
