package flow

import (
	"context"
	"errors"

	"github.com/direktiv/direktiv/pkg/model"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
)

func loadSource(rev []byte) (*model.Workflow, error) {
	workflow := new(model.Workflow)

	err := workflow.Load(rev)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}

func (flow *flow) placeholdSecrets(ctx context.Context, tx *database.SQLStore, ns string, file *filestore.File) error {
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
