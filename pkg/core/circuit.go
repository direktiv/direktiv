package core

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"
)

// nolint: containedctx
type Circuit struct {
	ctx  context.Context
	stop context.CancelFunc
	wg   sync.WaitGroup
}

func NewCircuit(parent context.Context, signals ...os.Signal) *Circuit {
	appCtx, appCancel := signal.NotifyContext(parent, signals...)

	return &Circuit{
		ctx:  appCtx,
		stop: appCancel,
		wg:   sync.WaitGroup{},
	}
}

func (c *Circuit) Context() context.Context {
	return c.ctx
}

func (c *Circuit) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Circuit) IsDone() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}

// Go lunches a goroutine and tracking it via a sync.WaitGroup. It enables simplified api to lunch graceful go
// routines.
func (c *Circuit) Go(job func() error) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		err := job()
		if err != nil {
			slog.Error("job crash", "err", err)
			c.stop()
		}
	}()
}

func (c *Circuit) Wait(timeout time.Duration) error {
	done := make(chan struct{})

	go func() {
		c.wg.Wait()
		close(done)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return errors.New("timeout exceeded")
	}
}
