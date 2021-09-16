package main

import (
	"sync"
	"time"
)

type stoppable interface {
	Stop()
}

var threads Threads

func init() {
	threads.finished = make(chan bool)
}

type Threads struct {
	finished chan bool
	stoppers []chan *time.Time
	stopped  *time.Time

	lock    sync.Mutex
	counter int
	code    int
}

func (t *Threads) Wait() {
	<-t.finished
	log.Info("All threads returned.")
}

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

func (t *Threads) ExitStatus() int {
	return t.code
}
