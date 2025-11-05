package sched

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/direktiv/direktiv/pkg/lifecycle"
	"k8s.io/utils/clock"
)

type TickFunc func() error

func startTicking(lc *lifecycle.Manager, clk clock.WithTicker, interval time.Duration, tick TickFunc) {
	lc.Go(func() error {
		t := clk.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-lc.Done():
				return nil
			case <-t.C():
				if err := tick(); err != nil {
					// log and continue
					fmt.Printf("sched tick error: %v\n", err)
				}
				// give a random jitter on ticking intervals
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			}
		}
	})
}
