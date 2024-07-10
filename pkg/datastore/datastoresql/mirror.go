package datastoresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlMirrorStore struct {
	// database connection.
	db *gorm.DB
	// symmetric encryption key to encrypt and decrypt mirror data.
	configEncryptionKey string
}

func (s sqlMirrorStore) GetAllConfigs(ctx context.Context) ([]*datastore.MirrorConfig, error) {
	list := []*datastore.MirrorConfig{}
	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM mirror_configs`).
		Find(&list)

	if res.Error != nil {
		return nil, res.Error
	}
	for i := range list {
		config, err := cryptDecryptConfig(list[i], s.configEncryptionKey, false)
		if err != nil {
			return nil, err
		}
		list[i] = config
	}

	return list, nil
}

func cryptDecryptConfig(config *datastore.MirrorConfig, key string, encrypt bool) (*datastore.MirrorConfig, error) {
	resultConfig := &datastore.MirrorConfig{}

	*resultConfig = *config

	targets := []*string{
		&resultConfig.AuthToken,
		&resultConfig.PrivateKeyPassphrase,
		&resultConfig.PrivateKey,
	}

	for i := range targets {
		t := targets[i]

		var (
			b   string
			err error
		)
		if encrypt {
			b, err = utils.EncryptDataBase64([]byte(key), []byte(*t))
		} else if len(*t) > 0 {
			b, err = utils.DecryptDataBase64([]byte(key), *t)
		}

		if err != nil {
			return nil, err
		}
		*t = b
	}

	return resultConfig, nil
}

func (s sqlMirrorStore) CreateConfig(ctx context.Context, config *datastore.MirrorConfig) (*datastore.MirrorConfig, error) {
	newConfig, err := cryptDecryptConfig(config, s.configEncryptionKey, true)
	if err != nil {
		return nil, err
	}

	config.Normalize()
	if errs := newConfig.Validate(); len(errs) > 0 {
		return nil, errs
	}

	res := s.db.WithContext(ctx).Table("mirror_configs").Create(&newConfig)
	if res.Error != nil {
		return nil, res.Error
	}

	return s.GetConfig(ctx, newConfig.Namespace)
}

func (s sqlMirrorStore) UpdateConfig(ctx context.Context, config *datastore.MirrorConfig) (*datastore.MirrorConfig, error) {
	config, err := cryptDecryptConfig(config, s.configEncryptionKey, true)
	if err != nil {
		return nil, err
	}

	config.Normalize()
	if errs := config.Validate(); len(errs) > 0 {
		return nil, errs
	}

	res := s.db.WithContext(ctx).Table("mirror_configs").
		Where("namespace", config.Namespace).
		Updates(map[string]interface{}{
			"url":                    config.URL,
			"git_ref":                config.GitRef,
			"public_key":             config.PublicKey,
			"private_key":            config.PrivateKey,
			"auth_token":             config.AuthToken,
			"private_key_passphrase": config.PrivateKeyPassphrase,
			"insecure":               config.Insecure,
		})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, datastore.ErrNotFound
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return s.GetConfig(ctx, config.Namespace)
}

func (s sqlMirrorStore) DeleteConfig(ctx context.Context, namespace string) error {
	res := s.db.WithContext(ctx).Exec(`DELETE FROM mirror_configs WHERE namespace=?`, namespace)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return datastore.ErrNotFound
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected gorm delete count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

func (s sqlMirrorStore) GetConfig(ctx context.Context, namespace string) (*datastore.MirrorConfig, error) {
	config := &datastore.MirrorConfig{}
	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM mirror_configs
					WHERE namespace=?`, namespace).
		First(config)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	config, err := cryptDecryptConfig(config, s.configEncryptionKey, false)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s sqlMirrorStore) CreateProcess(ctx context.Context, process *datastore.MirrorProcess) (*datastore.MirrorProcess, error) {
	newProcess := *process
	res := s.db.WithContext(ctx).Table("mirror_processes").Create(&newProcess)
	if res.Error != nil {
		return nil, res.Error
	}

	return &newProcess, nil
}

func (s sqlMirrorStore) UpdateProcess(ctx context.Context, process *datastore.MirrorProcess) (*datastore.MirrorProcess, error) {
	res := s.db.WithContext(ctx).Table("mirror_processes").
		Where("id", process.ID).
		Updates(process)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return s.GetProcess(ctx, process.ID)
}

func (s sqlMirrorStore) GetProcess(ctx context.Context, id uuid.UUID) (*datastore.MirrorProcess, error) {
	process := &datastore.MirrorProcess{}
	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM mirror_processes
					WHERE id=?`, id).
		First(process)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, datastore.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return process, nil
}

func (s sqlMirrorStore) GetProcessesByNamespace(ctx context.Context, namespace string) ([]*datastore.MirrorProcess, error) {
	var process []*datastore.MirrorProcess

	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM mirror_processes
					WHERE namespace=? ORDER BY created_at DESC `, namespace).
		Find(&process)

	if res.Error != nil {
		return nil, res.Error
	}

	return process, nil
}

func (s sqlMirrorStore) GetUnfinishedProcesses(ctx context.Context) ([]*datastore.MirrorProcess, error) {
	var process []*datastore.MirrorProcess

	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM mirror_processes
					WHERE ended_at IS NULL`).
		Find(&process)

	if res.Error != nil {
		return nil, res.Error
	}

	return process, nil
}

func (s sqlMirrorStore) DeleteOldProcesses(ctx context.Context, before time.Time) error {
	var process []*datastore.MirrorProcess

	res := s.db.WithContext(ctx).Raw(`
					DELETE FROM mirror_processes
					WHERE ended_at <> '0001-01-01T00:00:00Z' AND ended_at < ?`, before).
		Find(&process)

	if res.Error != nil {
		return res.Error
	}

	return nil
}

var _ datastore.MirrorStore = sqlMirrorStore{}
