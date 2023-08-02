package flow

import (
	"context"
	"errors"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
)

func (flow *flow) placeholdSecrets(ctx context.Context, tx *sqlTx, nsID uuid.UUID, file *filestore.File) error {
	rev, err := tx.FileStore().ForFile(file).GetCurrentRevision(ctx)
	if err != nil {
		return err
	}

	wf, err := loadSource(rev)
	if err != nil {
		return err
	}

	secretRefs := wf.GetSecretReferences()

	for _, secretRef := range secretRefs {
		_, err = tx.DataStore().Secrets().Get(ctx, nsID, secretRef)
		if errors.Is(err, core.ErrSecretNotFound) {
			err = tx.DataStore().Secrets().Set(ctx, &core.Secret{
				ID:          uuid.New(),
				NamespaceID: nsID,
				Name:        secretRef,
				Data:        nil,
			})
			if err != nil {
				continue
			}
		} else if err != nil {
			continue
		}
	}

	return nil
}
