package mirror

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TODO: validate credentials helper
// TODO: failed mirror garbage collector?

type Manager struct {
	callbacks Callbacks
}

func NewManager(callbacks Callbacks) *Manager {
	return &Manager{
		callbacks: callbacks,
	}
}

// Cancel stops a currently running mirroring process.
func (d *Manager) Cancel(ctx context.Context, processID uuid.UUID) error {
	// TODO

	return nil
}

func (d *Manager) silentFailProcess(p *Process) {
	p.Status = ProcessStatusFailed
	p.EndedAt = time.Now().UTC()
	_, e := d.callbacks.Store().UpdateProcess(context.Background(), p)
	if e != nil {
		d.callbacks.SysLogCrit(fmt.Sprintf("Error updating failed mirror process record in database: %v", e))
		return
	}
}

func (d *Manager) failProcess(p *Process, err error) {
	d.silentFailProcess(p)
	d.callbacks.ProcessLogger().Error(p.ID, fmt.Sprintf("Mirroring process failed %v", err), "process_id", p.ID)
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

// Execute
func (d *Manager) Execute(ctx context.Context, p *Process, get func(ctx context.Context) (Source, error), applyer Applyer) {
	err := d.setProcessStatus(ctx, p, ProcessStatusExecuting)
	if err != nil {
		d.failProcess(p, fmt.Errorf("updating process status: %w", err))
		return
	}

	src, err := get(ctx)
	if err != nil {
		d.failProcess(p, fmt.Errorf("initializing source: %v", err))
		return
	}
	defer func() {
		err := src.Free()
		if err != nil {
			d.callbacks.SysLogCrit(fmt.Sprintf("Error freeing mirror source: %v", err))
		}
	}()

	parser, err := NewParser(newPIDFormatLogger(d.callbacks.ProcessLogger(), p.ID), src)
	if err != nil {
		d.silentFailProcess(p)
		return
	}
	defer func() {
		err := parser.Close()
		if err != nil {
			d.callbacks.SysLogCrit(fmt.Sprintf("Error freeing mirror temporary files: %v", err))
		}
	}()

	err = applyer.apply(ctx, d.callbacks, p, parser, src.Notes())
	if err != nil {
		d.failProcess(p, fmt.Errorf("applying changes: %v", err))
		return
	}

	err = d.setProcessStatus(ctx, p, ProcessStatusComplete)
	if err != nil {
		d.failProcess(p, fmt.Errorf("updating process status: %w", err))
		return
	}
}
