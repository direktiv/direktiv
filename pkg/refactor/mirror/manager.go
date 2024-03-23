package mirror

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/google/uuid"
)

// TODO: validate credentials helper

type Manager struct {
	callbacks Callbacks
	local     sync.Map
}

func NewManager(callbacks Callbacks) *Manager {
	mgr := &Manager{
		callbacks: callbacks,
	}

	go mgr.gc()

	return mgr
}

// Garbage collector.
func (d *Manager) gc() {
	ctx := context.Background()

	jitter := 1000
	interval := time.Second * 10
	maxRunTime := 5 * time.Minute
	maxRecordTime := time.Hour * 48

	// TODO: gracefully close the loop
	for {
		a, _ := rand.Int(rand.Reader, big.NewInt(int64(jitter)*2))
		delta := int(a.Int64()) - jitter // this gets a value between +/- jitter
		time.Sleep(interval + time.Duration(delta*int(time.Millisecond)))

		// this first loop looks for operations that seem to have timed out
		procs, err := d.callbacks.Store().GetUnfinishedProcesses(ctx)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to query unfinished mirror processes: %v", err))

			continue
		}

		for _, proc := range procs {
			if time.Since(proc.CreatedAt) > maxRunTime {
				slog.Error(fmt.Sprintf("Detected an old unfinished mirror process '%s' for namespace '%s'. Terminating...", proc.ID.String(), proc.Namespace))
				c, cancel := context.WithTimeout(ctx, 5*time.Second)
				err = d.Cancel(c, proc.ID)
				cancel()
				if err != nil {
					slog.Error(fmt.Sprintf("Error cancelling old unfinished mirror process '%s' for namespace '%s': %v", proc.ID.String(), proc.Namespace, err))
				}

				p, err := d.callbacks.Store().GetProcess(ctx, proc.ID)
				if err != nil {
					slog.Error(fmt.Sprintf("Error requerying old unfinished mirror process '%s' for namespace '%s': %v", proc.ID.String(), proc.Namespace, err))

					continue
				}

				if p.Status != datastore.ProcessStatusFailed && p.Status != datastore.ProcessStatusComplete {
					d.failProcess(p, errors.New("timed out"))
				}
			}
		}

		// this second loop deletes really old processes from the database so that it doesn't grow endlessly
		err = d.callbacks.Store().DeleteOldProcesses(ctx, time.Now().Add(-1*maxRecordTime))
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to query old mirror processes: %v", err))

			continue
		}
	}
}

// Cancel stops a currently running mirroring process.
func (d *Manager) Cancel(_ context.Context, id uuid.UUID) error {
	v, ok := d.local.Load(id.String())
	if !ok {
		return nil // Not running on this machine. It's the caller's responsibility to ensure the whole cluster gets this call.
	}

	cancel, ok := v.(func())
	if !ok {
		panic(v)
	}

	cancel()

	return nil
}

func (d *Manager) silentFailProcess(p *datastore.MirrorProcess) {
	p.Status = datastore.ProcessStatusFailed
	p.EndedAt = time.Now().UTC()
	_, e := d.callbacks.Store().UpdateProcess(context.Background(), p)
	if e != nil {
		slog.Error(fmt.Sprintf("Error updating failed mirror process record in database: %v", e))

		return
	}
}

func (d *Manager) failProcess(p *datastore.MirrorProcess, err error) {
	d.silentFailProcess(p)
	d.callbacks.ProcessLogger().Error(p.ID, fmt.Sprintf("Mirroring process failed %v", err), "process_id", p.ID)
}

func (d *Manager) setProcessStatus(ctx context.Context, process *datastore.MirrorProcess, status string) error {
	process.Status = status
	if status == datastore.ProcessStatusComplete || status == datastore.ProcessStatusFailed {
		process.EndedAt = time.Now().UTC()
	}

	_, err := d.callbacks.Store().UpdateProcess(ctx, process)
	if err != nil {
		return err
	}

	return nil
}

// Execute ..
func (d *Manager) Execute(ctx context.Context, p *datastore.MirrorProcess, m *datastore.MirrorConfig, applyer Applyer) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		// TODO: find a way to store a separate status 'cancelled' instead of 'error'?
		d.local.Delete(p.ID.String())
	}()
	d.local.Store(p.ID.String(), cancel)

	err := d.setProcessStatus(ctx, p, datastore.ProcessStatusExecuting)
	if err != nil {
		//nolint:contextcheck
		d.failProcess(p, fmt.Errorf("updating process status: %w", err))

		return
	}

	src, err := GetSource(ctx, m)
	if err != nil {
		//nolint:contextcheck
		d.failProcess(p, fmt.Errorf("initializing source: %w", err))

		return
	}
	defer func() {
		err := src.Free()
		if err != nil {
			slog.Error(fmt.Sprintf("Error freeing mirror source: %v", err))
		}
	}()

	parser, err := NewParser(newPIDFormatLogger(d.callbacks.ProcessLogger(), p.ID), src)
	if err != nil {
		//nolint:contextcheck
		d.silentFailProcess(p)

		return
	}
	defer func() {
		err := parser.Close()
		if err != nil {
			slog.Error(fmt.Sprintf("Error freeing mirror temporary files: %v", err))
		}
	}()

	err = applyer.apply(ctx, d.callbacks, p, parser, src.Notes())
	if err != nil {
		//nolint:contextcheck
		d.failProcess(p, fmt.Errorf("applying changes: %w", err))

		return
	}

	err = d.setProcessStatus(ctx, p, datastore.ProcessStatusComplete)
	if err != nil {
		//nolint:contextcheck
		d.failProcess(p, fmt.Errorf("updating process status: %w", err))

		return
	}
}
