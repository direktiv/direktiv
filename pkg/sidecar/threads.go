package sidecar

import (
	"log/slog"
	"sync"
	"time"
)

var threads Threads

func init() {
	threads.finished = make(chan bool)
}

// Threads definition of a thread being maintained on a sidecar.
type Threads struct {
	finished chan bool
	stoppers []chan *time.Time
	stopped  *time.Time

	lock    sync.Mutex
	counter int
	code    int
}

// Wait waits until all the threads have been returned.
func (t *Threads) Wait() {
	<-t.finished
	slog.Info("All threads returned.")
}

// Register adds a new channel to the threads.
func (t *Threads) Register(stopper chan *time.Time) func() {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.stopped != nil {
		stopper <- t.stopped
		close(stopper)
	} else {
		t.stoppers = append(t.stoppers, stopper)
	}

	t.counter++

	return func() {
		t.lock.Lock()
		defer t.lock.Unlock()

		t.counter--

		if t.counter == 0 {
			close(t.finished)
		}
	}
}

// Stop stops a thread.
func (t *Threads) Stop(st *time.Time, code int) {
	t.lock.Lock()

	if t.code != 0 {
		t.code = code
	}

	if t.stopped != nil {
		t.lock.Unlock()
		return
	}

	t.lock.Unlock()

	t.stopped = st

	for _, stopper := range t.stoppers {
		stopper <- t.stopped
		close(stopper)
	}
}

// ExitStatus returns the exit status of a code.
func (t *Threads) ExitStatus() int {
	return t.code
}
