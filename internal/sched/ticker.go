package sched

import (
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/lifecycle"
)

type TickFunc func() error

func startTicking(lc *lifecycle.Manager, interval time.Duration, tickFunc TickFunc) {
	lc.Go(func() error {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-lc.Done():
				return nil
			case <-t.C:
				err := tickFunc()
				if err != nil {
					return fmt.Errorf(" tick func, err: %w", err)
				}
			}
		}
	})
}
