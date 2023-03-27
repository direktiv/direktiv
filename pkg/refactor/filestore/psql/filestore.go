package psql

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SQLFileStore struct {
	db *gorm.DB
}

func (s *SQLFileStore) Tx(ctx context.Context, fun func(ctx context.Context, fStore filestore.FileStore) error) error {
	db := s.db.WithContext(ctx).Begin()
	if db.Error != nil {
		return db.Error
	}
	defer db.WithContext(ctx).Rollback()
	newSqlStore := &SQLFileStore{
		db: db,
	}
	if err := fun(ctx, newSqlStore); err != nil {
		return err
	}

	return db.WithContext(ctx).Commit().Error
}

type TxSQLFileStore struct {
	*SQLFileStore
}

func (t *TxSQLFileStore) Commit(ctx context.Context) error {
	return t.db.WithContext(ctx).Commit().Error
}

func (t *TxSQLFileStore) Rollback(ctx context.Context) error {
	return t.db.WithContext(ctx).Rollback().Error
}

func (s *SQLFileStore) Begin(ctx context.Context) (filestore.TxFileStore, error) {
	db := s.db.WithContext(ctx).Begin()
	if db.Error != nil {
		return nil, db.Error
	}

	return &TxSQLFileStore{
		SQLFileStore: &SQLFileStore{
			db: db,
		},
	}, nil
}

func (s *SQLFileStore) ForRootID(rootID uuid.UUID) filestore.RootQuery {
	return &RootQuery{
		rootID:       rootID,
		db:           s.db,
		checksumFunc: filestore.DefaultCalculateChecksum,
	}
}

func (s *SQLFileStore) ForFile(file *filestore.File) filestore.FileQuery {
	return &FileQuery{
		file: file,
		db:   s.db,
	}
}

func (s *SQLFileStore) ForRevision(revision *filestore.Revision) filestore.RevisionQuery {
	return &RevisionQuery{
		rev:          revision,
		db:           s.db,
		checksumFunc: filestore.DefaultCalculateChecksum,
	}
}

var _ filestore.FileStore = &SQLFileStore{} // Ensures SQLFileStore struct conforms to filestore.FileStore interface.

func NewSQLFileStore(db *gorm.DB) (*SQLFileStore, error) {
	type File struct {
		filestore.File
		Revisions []filestore.Revision `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	}
	type Root struct {
		filestore.Root
		Files []File `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	}

	err := db.AutoMigrate(&Root{}, &File{}, &filestore.Revision{})
	if err != nil {
		return nil, err
	}

	return &SQLFileStore{
		db: db,
	}, nil
}

func NewMockFileStore() (*SQLFileStore, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	type File struct {
		filestore.File
		Revisions []filestore.Revision `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	}
	type Root struct {
		filestore.Root
		Files []File `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	}

	err = db.AutoMigrate(&Root{}, &File{}, &filestore.Revision{})
	if err != nil {
		return nil, err
	}

	return &SQLFileStore{
		db: db,
	}, nil
}

func (s *SQLFileStore) CreateRoot(ctx context.Context, id uuid.UUID) (*filestore.Root, error) {
	n := &filestore.Root{ID: id}
	res := s.db.WithContext(ctx).Create(n)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return n, nil
}

//nolint:ireturn
func (s *SQLFileStore) GetAllRoots(ctx context.Context) ([]*filestore.Root, error) {
	var list []filestore.Root
	res := s.db.WithContext(ctx).Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	var ns []*filestore.Root
	for i := range list {
		ns = append(ns, &list[i])
	}

	return ns, nil
}
