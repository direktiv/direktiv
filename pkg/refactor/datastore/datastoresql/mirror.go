package datastoresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlMirrorStore struct {
	// database connection.
	db *gorm.DB
	// symmetric encryption key to encrypt and decrypt mirror data.
	configEncryptionKey string
}

func (s sqlMirrorStore) GetAllConfigs(ctx context.Context) ([]*mirror.Config, error) {
	list := []*mirror.Config{}
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

func cryptDecryptConfig(config *mirror.Config, key string, encrypt bool) (*mirror.Config, error) {
	resultConfig := &mirror.Config{}

	*resultConfig = *config

	targets := []*string{
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
			b, err = util.EncryptDataBase64([]byte(key), []byte(*t))
		} else if len(*t) > 0 {
			b, err = util.DecryptDataBase64([]byte(key), *t)
		}

		if err != nil {
			return nil, err
		}
		*t = b
	}

	return resultConfig, nil
}

func (s sqlMirrorStore) CreateConfig(ctx context.Context, config *mirror.Config) (*mirror.Config, error) {
	newConfig, err := cryptDecryptConfig(config, s.configEncryptionKey, true)
	if err != nil {
		return nil, err
	}

	res := s.db.WithContext(ctx).Table("mirror_configs").Create(&newConfig)
	if res.Error != nil {
		return nil, res.Error
	}

	return s.GetConfig(ctx, newConfig.Namespace)
}

func (s sqlMirrorStore) UpdateConfig(ctx context.Context, config *mirror.Config) (*mirror.Config, error) {
	config, err := cryptDecryptConfig(config, s.configEncryptionKey, true)
	if err != nil {
		return nil, err
	}

	res := s.db.WithContext(ctx).Table("mirror_configs").
		Where("namespace", config.Namespace).
		Updates(map[string]interface{}{
			"url":                    config.URL,
			"git_ref":                config.GitRef,
			"git_commit_hash":        config.GitCommitHash,
			"public_key":             config.PublicKey,
			"private_key":            config.PrivateKey,
			"private_key_passphrase": config.PrivateKeyPassphrase,
			"insecure":               config.Insecure,
		})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return s.GetConfig(ctx, config.Namespace)
}

func (s sqlMirrorStore) GetConfig(ctx context.Context, namespace string) (*mirror.Config, error) {
	config := &mirror.Config{}
	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM mirror_configs
					WHERE namespace=?`, namespace).
		First(config)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, mirror.ErrNotFound
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

func (s sqlMirrorStore) CreateProcess(ctx context.Context, process *mirror.Process) (*mirror.Process, error) {
	newProcess := *process
	res := s.db.WithContext(ctx).Table("mirror_processes").Create(&newProcess)
	if res.Error != nil {
		return nil, res.Error
	}

	return &newProcess, nil
}

func (s sqlMirrorStore) UpdateProcess(ctx context.Context, process *mirror.Process) (*mirror.Process, error) {
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

func (s sqlMirrorStore) GetProcess(ctx context.Context, id uuid.UUID) (*mirror.Process, error) {
	process := &mirror.Process{}
	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM mirror_processes
					WHERE id=?`, id).
		First(process)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, mirror.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return process, nil
}

func (s sqlMirrorStore) GetProcessesByNamespace(ctx context.Context, namespace string) ([]*mirror.Process, error) {
	var process []*mirror.Process

	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM mirror_processes
					WHERE namespace=?`, namespace).
		Find(&process)

	if res.Error != nil {
		return nil, res.Error
	}

	return process, nil
}

func (s sqlMirrorStore) GetUnfinishedProcesses(ctx context.Context) ([]*mirror.Process, error) {
	var process []*mirror.Process

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
	var process []*mirror.Process

	res := s.db.WithContext(ctx).Raw(`
					DELETE FROM mirror_processes
					WHERE ended_at < ?`, before).
		Find(&process)

	if res.Error != nil {
		return res.Error
	}

	return nil
}

var _ mirror.Store = sqlMirrorStore{}
