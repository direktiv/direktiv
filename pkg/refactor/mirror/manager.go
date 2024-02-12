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

	"github.com/direktiv/direktiv/pkg/flow/database/recipient"
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
			slog.Error("Failed to query unfinished mirror processes", "error", err)

			continue
		}

		for _, proc := range procs {
			if time.Since(proc.CreatedAt) > maxRunTime {
				slog.Error("Detected an old unfinished mirror process for namespace", "process_id", proc.ID.String(), "namespace", proc.Namespace, "stream", recipient.Namespace.String()+"."+proc.Namespace)
				c, cancel := context.WithTimeout(ctx, 5*time.Second)
				err = d.Cancel(c, proc.ID)
				cancel()
				if err != nil {
					slog.Error("cancelling old unfinished mirror process", "process_id", proc.ID.String(), "namespace", proc.Namespace, "error", err, "stream", recipient.Namespace.String()+"."+proc.Namespace)
				}

				p, err := d.callbacks.Store().GetProcess(ctx, proc.ID)
				if err != nil {
					slog.Error("requerying old unfinished mirror process", "process_id", proc.ID.String(), "namespace", proc.Namespace, "error", err, "stream", recipient.Namespace.String()+"."+proc.Namespace)

					continue
				}

				if p.Status != ProcessStatusFailed && p.Status != ProcessStatusComplete {
					d.failProcess(p, errors.New("timed out"))
				}
			}
		}

		// this second loop deletes really old processes from the database so that it doesn't grow endlessly
		err = d.callbacks.Store().DeleteOldProcesses(ctx, time.Now().Add(-1*maxRecordTime))
		if err != nil {
			slog.Error("Failed to query old mirror processes", "error", err)

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

func (d *Manager) silentFailProcess(p *Process) {
	p.Status = ProcessStatusFailed
	p.EndedAt = time.Now().UTC()
	_, err := d.callbacks.Store().UpdateProcess(context.Background(), p)
	if err != nil {
		slog.Error("Error updating failed mirror process record in database", "error", err, "namespace", p.Namespace, "stream", recipient.Namespace.String()+"."+p.Namespace)

		return
	}
}

func (d *Manager) failProcess(p *Process, err error) {
	d.silentFailProcess(p)
	slog.Error("Mirroring process failed", "error", err, "namespace", p.Namespace, "process_id", p.ID, "stream", recipient.Namespace.String()+"."+p.Namespace)
}

func (d *Manager) setProcessStatus(ctx context.Context, process *Process, status string) error {
	process.Status = status
	if status == ProcessStatusComplete || status == ProcessStatusFailed {
		process.EndedAt = time.Now().UTC()
	}

	_, err := d.callbacks.Store().UpdateProcess(ctx, process)
	if err != nil {
		return err
	}

	return nil
}

// Execute ..
func (d *Manager) Execute(ctx context.Context, p *Process, get func(ctx context.Context) (Source, error), applyer Applyer) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		// TODO: find a way to store a separate status 'cancelled' instead of 'error'?
		d.local.Delete(p.ID.String())
	}()
	d.local.Store(p.ID.String(), cancel)

	err := d.setProcessStatus(ctx, p, ProcessStatusExecuting)
	if err != nil {
		//nolint:contextcheck
		d.failProcess(p, fmt.Errorf("updating process status: %w", err))

		return
	}

	src, err := get(ctx)
	if err != nil {
		//nolint:contextcheck
		d.failProcess(p, fmt.Errorf("initializing source: %w", err))

		return
	}
	defer func() {
		err := src.Free()
		if err != nil {
			slog.Error("Error freeing mirror source", "error", err, "namespace", p.Namespace, "process_id", p.ID, "stream", recipient.Namespace.String()+"."+p.Namespace)
		}
	}()

	parser, err := NewParser(p.Namespace, src)
	if err != nil {
		//nolint:contextcheck
		d.silentFailProcess(p)

		return
	}
	defer func() {
		err := parser.Close()
		if err != nil {
			slog.Error("Error freeing mirror temporary files", "error", err, "namespace", p.Namespace, "process_id", p.ID, "stream", recipient.Namespace.String()+"."+p.Namespace)
		}
	}()

	err = applyer.apply(ctx, d.callbacks, p, parser, src.Notes())
	if err != nil {
		//nolint:contextcheck
		d.failProcess(p, fmt.Errorf("applying changes: %w", err))

		return
	}

	err = d.setProcessStatus(ctx, p, ProcessStatusComplete)
	if err != nil {
		//nolint:contextcheck
		d.failProcess(p, fmt.Errorf("updating process status: %w", err))

		return
	}
}
