package flow

import (
	"context"
	"errors"

	database2 "github.com/direktiv/direktiv/pkg/refactor/database"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
)

func (flow *flow) placeholdSecrets(ctx context.Context, tx *database2.DB, ns string, file *filestore.File) error {
	data, err := tx.FileStore().ForFile(file).GetData(ctx)
	if err != nil {
		return err
	}

	wf, err := loadSource(data)
	if err != nil {
		return err
	}

	secretRefs := wf.GetSecretReferences()

	for _, secretRef := range secretRefs {
		_, err = tx.DataStore().Secrets().Get(ctx, ns, secretRef)
		if errors.Is(err, datastore.ErrNotFound) {
			err = tx.DataStore().Secrets().Set(ctx, &datastore.Secret{
				Namespace: ns,
				Name:      secretRef,
				Data:      nil,
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
