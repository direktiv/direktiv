package secrets

import (
	"context"
	"errors"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/database"
	"github.com/direktiv/direktiv/internal/datastore"
)

type DBSecrets struct {
	namespace string
	db        *database.DB
}

func (dbs *DBSecrets) Get(ctx context.Context, name string) (*core.Secret, error) {
	db, err := dbs.db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Rollback()
	dStore := db.DataStore()

	// Fetch one
	s, err := dStore.Secrets().Get(ctx, dbs.namespace, name)
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &core.Secret{
		Name:      name,
		Data:      s.Data,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}, nil
}

func (dbs *DBSecrets) Set(ctx context.Context, secret *core.Secret) (*core.Secret, error) {
	db, err := dbs.db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Rollback()
	dStore := db.DataStore()

	s := &datastore.Secret{
		Name:      secret.Name,
		Namespace: dbs.namespace,
		Data:      secret.Data,
	}

	err = dStore.Secrets().Set(ctx, s)
	if err != nil {
		return nil, err
	}

	v, err := dStore.Secrets().Get(ctx, dbs.namespace, secret.Name)
	if err != nil {
		return nil, err
	}

	secret.CreatedAt = v.CreatedAt
	secret.UpdatedAt = v.UpdatedAt

	return secret, db.Commit(ctx)
}

func (dbs *DBSecrets) GetAll(ctx context.Context) ([]*core.Secret, error) {
	db, err := dbs.db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Rollback()
	dStore := db.DataStore()

	var list []*datastore.Secret
	list, err = dStore.Secrets().GetAll(ctx, dbs.namespace)
	if err != nil {
		return nil, err
	}

	res := make([]*core.Secret, len(list))
	for i := range list {
		res[i] = &core.Secret{
			Name:      list[i].Name,
			Data:      list[i].Data,
			CreatedAt: list[i].CreatedAt,
			UpdatedAt: list[i].UpdatedAt,
		}
	}

	return res, nil
}

func (dbs *DBSecrets) Update(ctx context.Context, secret *core.Secret) (*core.Secret, error) {
	db, err := dbs.db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Rollback()
	dStore := db.DataStore()

	_, err = dStore.Secrets().Get(ctx, dbs.namespace, secret.Name)
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	err = dStore.Secrets().Update(ctx, &datastore.Secret{
		Namespace: dbs.namespace,
		Name:      secret.Name,
		Data:      secret.Data,
	})
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	// Fetch the updated one
	s, err := dStore.Secrets().Get(ctx, dbs.namespace, secret.Name)
	if err != nil {
		return nil, err
	}

	err = db.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &core.Secret{
		Name:      s.Name,
		Data:      s.Data,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}, nil
}

func (dbs *DBSecrets) Delete(ctx context.Context, name string) error {
	db, err := dbs.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer db.Rollback()
	dStore := db.DataStore()

	// Fetch one
	err = dStore.Secrets().Delete(ctx, dbs.namespace, name)
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			return ErrNotFound
		}

		return err
	}

	return db.Commit(ctx)
}
