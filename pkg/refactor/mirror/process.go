package mirror

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (d *Manager) NewProcess(ctx context.Context, nsID, rootID uuid.UUID, processType string) (*Process, error) {
	// TODO: prevent multiple sync operations on a single namespace

	process, err := d.callbacks.Store().CreateProcess(ctx, &Process{
		ID:          uuid.New(),
		NamespaceID: nsID,
		RootID:      rootID,
		Typ:         processType,
		Status:      ProcessStatusPending,
	})
	if err != nil {
		return nil, fmt.Errorf("creating a new mirroring process, err: %w", err)
	}

	return process, nil
}
