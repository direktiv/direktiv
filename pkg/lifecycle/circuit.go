package lifecycle

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
type Manager struct {
	ctx  context.Context
	stop context.CancelFunc
	wg   sync.WaitGroup
}

func New(parent context.Context, signals ...os.Signal) *Manager {
	appCtx, appCancel := signal.NotifyContext(parent, signals...)

	return &Manager{
		ctx:  appCtx,
		stop: appCancel,
		wg:   sync.WaitGroup{},
	}
}

func (c *Manager) Context() context.Context {
	return c.ctx
}

func (c *Manager) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Manager) IsDone() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}

// Go lunches a goroutine and tracking it via a sync.WaitGroup. It enables simplified api to lunch graceful go
// routines.
func (c *Manager) Go(job func() error) {
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

func (c *Manager) OnShutdown(job func() error) {
	c.Go(func() error {
		<-c.Done()
		return job()
	})
}

func (c *Manager) Wait(timeout time.Duration) error {
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
