package datastoresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/util"
	"gorm.io/gorm"
)

type sqlSecretsStore struct {
	db        *gorm.DB
	secretKey string
}

func (s sqlSecretsStore) Update(ctx context.Context, secret *core.Secret) error {
	if secret.Data != nil {
		var err error
		secret.Data, err = util.EncryptData([]byte(s.secretKey), secret.Data)
		if err != nil {
			return err
		}
	}
	res := s.db.WithContext(ctx).Exec(`UPDATE secrets SET data=? WHERE namespace=? and name=?`,
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
		return core.ErrSecretNotFound
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (s sqlSecretsStore) Get(ctx context.Context, namespace string, name string) (*core.Secret, error) {
	secret := &core.Secret{}
	res := s.db.WithContext(ctx).Raw(`
			SELECT * FROM secrets WHERE namespace=? AND name=?`,
		namespace, name).
		First(secret)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, core.ErrSecretNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}
	if secret.Data != nil {
		var err error
		secret.Data, err = util.DecryptData([]byte(s.secretKey), secret.Data)
		if err != nil {
			return nil, err
		}
	}

	return secret, nil
}

func (s sqlSecretsStore) Set(ctx context.Context, secret *core.Secret) error {
	var res *gorm.DB
	x, err := s.Get(ctx, secret.Namespace, secret.Name)
	//nolint:nestif
	if errors.Is(err, core.ErrSecretNotFound) {
		if secret.Data == nil {
			res = s.db.WithContext(ctx).Exec(`
				INSERT INTO secrets(namespace, name) VALUES(?, ?)
				`, secret.Namespace, secret.Name)
		} else {
			var err error
			secret.Data, err = util.EncryptData([]byte(s.secretKey), secret.Data)
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
				UPDATE secrets SET data=NULL WHERE namespace=? AND name=?
				`, x.Namespace, x.Name)
		} else {
			var err error
			secret.Data, err = util.EncryptData([]byte(s.secretKey), secret.Data)
			if err != nil {
				return err
			}
			res = s.db.WithContext(ctx).Exec(`
				UPDATE secrets SET data=? WHERE namespace=? AND name=?
				`, secret.Data, x.Namespace, x.Name)
		}
	}

	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (s sqlSecretsStore) GetAll(ctx context.Context, namespace string) ([]*core.Secret, error) {
	var secrets []*core.Secret

	res := s.db.WithContext(ctx).Raw(`
							SELECT * FROM secrets WHERE namespace=?`,
		namespace).
		Find(&secrets)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, secret := range secrets {
		if secret.Data != nil {
			var err error
			secret.Data, err = util.DecryptData([]byte(s.secretKey), secret.Data)
			if err != nil {
				return nil, err
			}
		}
	}

	return secrets, nil
}

var _ core.SecretsStore = &sqlSecretsStore{}
