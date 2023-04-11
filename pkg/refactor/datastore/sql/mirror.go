package sql

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlMirrorStore struct {
	db *gorm.DB
}

const (
	encryptionKey = "DIREKTIV_SECRETS_KEY"
)

func crypt(config *mirror.Config, encrypt bool) error {
	key := os.Getenv(encryptionKey)

	targets := []*string{
		&config.PrivateKeyPassphrase,
		&config.PrivateKey,
	}

	for i := range targets {
		t := targets[i]

		var (
			b   string
			err error
		)
		if encrypt {
			b, err = util.EncryptDataBase64([]byte(key), []byte(*t))
		} else {
			b, err = util.DecryptDataBase64([]byte(key), *t)
		}
		if err != nil {
			return err
		}
		*t = b
	}

	return nil
}

func (s sqlMirrorStore) CreateConfig(ctx context.Context, config *mirror.Config) (*mirror.Config, error) {
	err := crypt(config, true)
	if err != nil {
		return nil, err
	}
	newConfig := *config

	res := s.db.WithContext(ctx).Table("mirror_configs").Create(&newConfig)
	if res.Error != nil {
		return nil, res.Error
	}

	err = crypt(&newConfig, false)
	if err != nil {
		return nil, err
	}

	return &newConfig, nil
}

func (s sqlMirrorStore) UpdateConfig(ctx context.Context, config *mirror.Config) (*mirror.Config, error) {
	err := crypt(config, true)
	if err != nil {
		return nil, err
	}

	res := s.db.WithContext(ctx).
		Table("mirror_configs").
		Where("id", config.ID).
		Updates(config)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return s.GetConfig(ctx, config.ID)
}

func (s sqlMirrorStore) GetConfig(ctx context.Context, id uuid.UUID) (*mirror.Config, error) {
	config := &mirror.Config{ID: id}
	res := s.db.WithContext(ctx).Table("mirror_configs").First(config)

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, mirror.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	err := crypt(config, false)
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
	res := s.db.WithContext(ctx).
		Table("mirror_processes").
		Model(&mirror.Process{}).
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
	process := &mirror.Process{ID: id}
	res := s.db.WithContext(ctx).Table("mirror_processes").First(process)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, mirror.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return process, nil
}

func (s sqlMirrorStore) GetProcessesByConfig(ctx context.Context, configID uuid.UUID) ([]*mirror.Process, error) {
	var process []*mirror.Process

	res := s.db.WithContext(ctx).
		Table("mirror_processes").
		Where("config_id", configID).
		Find(&process)
	if res.Error != nil {
		return nil, res.Error
	}

	return process, nil
}

var _ mirror.Store = sqlMirrorStore{}
