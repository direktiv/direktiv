package datasql

import (
	"context"
	"errors"

	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/utils"
	"gorm.io/gorm"
)

type sqlSecretsConfigsStore struct {
	db *gorm.DB
}

func (s sqlSecretsConfigsStore) Set(ctx context.Context, secretsConfigs *datastore.SecretsConfigs) error {
	var res *gorm.DB
	x, err := s.Get(ctx, secretsConfigs.Namespace)
	//nolint:nestif
	if errors.Is(err, datastore.ErrNotFound) {
		if secretsConfigs.Configuration == nil {
			res = s.db.WithContext(ctx).Exec(`
				INSERT INTO secrets_configs(namespace) VALUES(?)
				`, secretsConfigs.Namespace)
		} else {
			var err error
			secretsConfigs.Configuration, err = utils.EncryptData([]byte(datastore.SymmetricEncryptionKey), secretsConfigs.Configuration)
			if err != nil {
				return err
			}
			res = s.db.WithContext(ctx).Exec(`
				INSERT INTO secrets_configs(namespace, configuration) VALUES(?, ?)
				`, secretsConfigs.Namespace, secretsConfigs.Configuration)
		}
	} else if err != nil {
		return err
	} else {
		var err error
		secretsConfigs.Configuration, err = utils.EncryptData([]byte(datastore.SymmetricEncryptionKey), secretsConfigs.Configuration)
		if err != nil {
			return err
		}
		res = s.db.WithContext(ctx).Exec(`
			UPDATE secrets_configs SET configuration=? WHERE namespace=?, updated_at=CURRENT_TIMESTAMP 
			`, secretsConfigs.Configuration, x.Namespace)
	}

	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (s sqlSecretsConfigsStore) Get(ctx context.Context, namespace string) (*datastore.SecretsConfigs, error) {
	secretsConfigs := &datastore.SecretsConfigs{}
	res := s.db.WithContext(ctx).Raw(`
			SELECT * FROM secrets_configs WHERE namespace=?`,
		namespace).
		First(secretsConfigs)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}
	if secretsConfigs.Configuration != nil {
		var err error
		secretsConfigs.Configuration, err = utils.DecryptData([]byte(datastore.SymmetricEncryptionKey), secretsConfigs.Configuration)
		if err != nil {
			return nil, err
		}
	}

	return secretsConfigs, nil
}
