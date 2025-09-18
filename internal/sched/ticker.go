package sched

import (
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/lifecycle"
)

type TickFunc func() error

func startTicking(lc *lifecycle.Manager, interval time.Duration, tick TickFunc) {
	lc.Go(func() error {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-lc.Done():
				return nil
			case <-t.C:
				if err := tick(); err != nil {
					// log and continue
					fmt.Printf("sched tick error: %v\n", err)
				}
			}
		}
	})
}
