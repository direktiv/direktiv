package core

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
)

// nolint: containedctx
type Circuit struct {
	Context context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// Start lunches a goroutine and tracking it via a sync.WaitGroup. It enables simplified api to lunch graceful go
// routines.
func (c *Circuit) Start(job func() error) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		err := job()
		if err != nil {
			slog.Error("job crash", "err", err)
			c.cancel()
		}
	}()
}

func (c *Circuit) IsDone() bool {
	select {
	case <-c.Context.Done():
		return true
	default:
		return false
	}
}

func (c *Circuit) Wait() {
	c.wg.Wait()
}

func NewCircuit(parent context.Context, signals ...os.Signal) *Circuit {
	appCtx, appCancel := signal.NotifyContext(parent, signals...)

	return &Circuit{
		Context: appCtx,
		cancel:  appCancel,
		wg:      sync.WaitGroup{},
	}
}
