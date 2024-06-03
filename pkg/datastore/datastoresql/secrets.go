package datastoresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/utils"
	"gorm.io/gorm"
)

type sqlSecretsStore struct {
	db        *gorm.DB
	secretKey string
}

func (s sqlSecretsStore) Update(ctx context.Context, secret *datastore.Secret) error {
	if secret.Data != nil {
		var err error
		secret.Data, err = utils.EncryptData([]byte(s.secretKey), secret.Data)
		if err != nil {
			return err
		}
	}
	res := s.db.WithContext(ctx).Exec(`UPDATE secrets SET data=?, updated_at=CURRENT_TIMESTAMP WHERE namespace=? and name=?`,
		secret.Data, secret.Namespace, secret.Name)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (s sqlSecretsStore) Delete(ctx context.Context, namespace string, name string) error {
	res := s.db.WithContext(ctx).Exec(`DELETE FROM secrets WHERE namespace=? and name=?`,
		namespace, name)
	if res.Error != nil {
		return res.Error
	}
	// TODO: check if other delete queries check for row count == 0 and return not found error.
	if res.RowsAffected == 0 {
		return datastore.ErrNotFound
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (s sqlSecretsStore) Get(ctx context.Context, namespace string, name string) (*datastore.Secret, error) {
	secret := &datastore.Secret{}
	res := s.db.WithContext(ctx).Raw(`
			SELECT * FROM secrets WHERE namespace=? AND name=?`,
		namespace, name).
		First(secret)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}
	if secret.Data != nil {
		var err error
		secret.Data, err = utils.DecryptData([]byte(s.secretKey), secret.Data)
		if err != nil {
			return nil, err
		}
	}

	return secret, nil
}

func (s sqlSecretsStore) Set(ctx context.Context, secret *datastore.Secret) error {
	var res *gorm.DB
	x, err := s.Get(ctx, secret.Namespace, secret.Name)
	//nolint:nestif
	if errors.Is(err, datastore.ErrNotFound) {
		if secret.Data == nil {
			res = s.db.WithContext(ctx).Exec(`
				INSERT INTO secrets(namespace, name) VALUES(?, ?)
				`, secret.Namespace, secret.Name)
		} else {
			var err error
			secret.Data, err = utils.EncryptData([]byte(s.secretKey), secret.Data)
			if err != nil {
				return err
			}
			res = s.db.WithContext(ctx).Exec(`
				INSERT INTO secrets(namespace, name, data) VALUES(?, ?, ?)
				`, secret.Namespace, secret.Name, secret.Data)
		}
	} else if err != nil {
		return err
	} else {
		if secret.Data == nil {
			res = s.db.WithContext(ctx).Exec(`
				UPDATE secrets SET data=NULL WHERE namespace=? AND name=?, updated_at=CURRENT_TIMESTAMP 
				`, x.Namespace, x.Name)
		} else {
			var err error
			secret.Data, err = utils.EncryptData([]byte(s.secretKey), secret.Data)
			if err != nil {
				return err
			}
			res = s.db.WithContext(ctx).Exec(`
				UPDATE secrets SET data=? WHERE namespace=? AND name=?, updated_at=CURRENT_TIMESTAMP 
				`, secret.Data, x.Namespace, x.Name)
		}
	}

	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (s sqlSecretsStore) GetAll(ctx context.Context, namespace string) ([]*datastore.Secret, error) {
	var secrets []*datastore.Secret

	res := s.db.WithContext(ctx).Raw(`
							SELECT * FROM secrets WHERE namespace=? ORDER BY created_at`,
		namespace).
		Find(&secrets)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, secret := range secrets {
		if secret.Data != nil {
			var err error
			secret.Data, err = utils.DecryptData([]byte(s.secretKey), secret.Data)
			if err != nil {
				return nil, err
			}
		}
	}

	return secrets, nil
}

var _ datastore.SecretsStore = &sqlSecretsStore{}
