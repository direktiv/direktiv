package datastoresql

import (
	"context"
	"errors"
	"fmt"

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

// SetNamespaceVariable sets namespace variable into the database.
// nolint
func (s sqlMirrorStore) SetNamespaceVariable(ctx context.Context, namespaceID uuid.UUID, key string, data []byte, hash string, mType string) error {
	// try to update a variable if exists.
	res := s.db.WithContext(ctx).Exec(`
							UPDATE var_data SET size = ?, hash = ?, data = ?, mime_type = ?  WHERE oid = (
								SELECT var_data_varrefs FROM var_refs WHERE name = ? AND namespace_vars = ?
							)`, len(data), hash, data, mType, key, namespaceID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return nil
	}

	newID := uuid.New()

	// create var_data entry.
	res = s.db.WithContext(ctx).Exec(`
							INSERT INTO var_data(oid, size, hash, data, mime_type, created_at, updated_at) VALUES(?, ?, ?, ?, ?, NOW(), NOW());
							`,
		newID, len(data), hash, data, mType,
	)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected var_data inserted rows count, got: %d, want: %d", res.RowsAffected, 1)
	}

	// create var_refs entry.
	res = s.db.WithContext(ctx).Exec(`
							INSERT INTO var_refs(oid, name, behaviour, namespace_vars, var_data_varrefs) VALUES(?, ?, ?, ?, ?);
							`,
		uuid.New(), key, "", namespaceID, newID,
	)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected var_refs inserted rows count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
}

// SetWorkflowVariable sets workflow variable into the database.
// nolint
func (s sqlMirrorStore) SetWorkflowVariable(ctx context.Context, workflowID uuid.UUID, key string, data []byte, hash string, mType string) error {
	// try to update a variable if exists.
	res := s.db.WithContext(ctx).Exec(`
							UPDATE var_data SET size = ?, hash = ?, data = ?, mime_type = ?  WHERE oid = (
								SELECT var_data_varrefs FROM var_refs WHERE name = ? AND workflow_id = ?
							)`, len(data), hash, data, mType, key, workflowID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return nil
	}

	newID := uuid.New()

	// create var_data entry.
	res = s.db.WithContext(ctx).Exec(`
							INSERT INTO var_data(oid, size, hash, data, mime_type, created_at, updated_at) VALUES(?, ?, ?, ?, ?, NOW(), NOW());
							`,
		newID, len(data), hash, data, mType,
	)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected var_data inserted rows count, got: %d, want: %d", res.RowsAffected, 1)
	}

	// create var_refs entry.
	res = s.db.WithContext(ctx).Exec(`
							INSERT INTO var_refs(oid, name, behaviour, workflow_id, var_data_varrefs) VALUES(?, ?, ?, ?, ?);
							`,
		uuid.New(), key, "", workflowID, newID,
	)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("unexpected var_refs inserted rows count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return nil
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
		} else {
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

	return s.GetConfig(ctx, newConfig.NamespaceID)
}

func (s sqlMirrorStore) UpdateConfig(ctx context.Context, config *mirror.Config) (*mirror.Config, error) {
	config, err := cryptDecryptConfig(config, s.configEncryptionKey, true)
	if err != nil {
		return nil, err
	}

	res := s.db.WithContext(ctx).Table("mirror_configs").
		Where("namespace_id", config.NamespaceID).
		Updates(config)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm update count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return s.GetConfig(ctx, config.NamespaceID)
}

func (s sqlMirrorStore) GetConfig(ctx context.Context, namespaceID uuid.UUID) (*mirror.Config, error) {
	config := &mirror.Config{}
	res := s.db.WithContext(ctx).Table("mirror_configs").
		Where("namespace_id", namespaceID).
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
	res := s.db.WithContext(ctx).Table("mirror_processes").
		Where("id", id).
		First(process)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, mirror.ErrNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return process, nil
}

func (s sqlMirrorStore) GetProcessesByNamespaceID(ctx context.Context, namespaceID uuid.UUID) ([]*mirror.Process, error) {
	var process []*mirror.Process

	res := s.db.WithContext(ctx).Table("mirror_processes").
		Where("namespace_id", namespaceID).
		Find(&process)
	if res.Error != nil {
		return nil, res.Error
	}

	return process, nil
}

var _ mirror.Store = sqlMirrorStore{}
