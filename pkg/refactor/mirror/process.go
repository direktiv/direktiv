package mirror

import (
	"context"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/google/uuid"
)

func (d *Manager) NewProcess(ctx context.Context, ns *core.Namespace, processType string) (*Process, error) {
	// TODO: make this check 100% threadsafe in HA

	procs, err := d.callbacks.Store().GetProcessesByNamespace(ctx, ns.Name)
	if err != nil {
		return nil, fmt.Errorf("querying existing a mirroring processes, err: %w", err)
	}

	for _, proc := range procs {
		if status := proc.Status; status == ProcessStatusExecuting || status == ProcessStatusPending {
			return nil, errors.New("a mirroring process is already being executed on this namespace")
		}
	}

	process, err := d.callbacks.Store().CreateProcess(ctx, &Process{
		ID:        uuid.New(),
		Namespace: ns.Name,
		Typ:       processType,
		Status:    ProcessStatusPending,
	})
	if err != nil {
		return nil, fmt.Errorf("creating a new mirroring process, err: %w", err)
	}

	return process, nil
}
