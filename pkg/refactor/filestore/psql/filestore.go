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

func (s *SQLFileStore) ForRoot(root *filestore.Root) filestore.RootQuery {
	return &RootQuery{
		root: root,
		db:   s.db,
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
		rev: revision,
		db:  s.db,
	}
}

var _ filestore.FileStore = &SQLFileStore{} // Ensures SQLFileStore struct conforms to filestore.FileStore interface.

func NewSQLFileStore(db *gorm.DB) *SQLFileStore {
	return &SQLFileStore{
		db: db,
	}
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
	fs := NewSQLFileStore(db)

	return fs, nil
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

func (s *SQLFileStore) GetRoot(ctx context.Context, id uuid.UUID) (*filestore.Root, error) {
	n := &filestore.Root{ID: id}
	res := s.db.WithContext(ctx).First(n)
	if res.Error != nil {
		return nil, res.Error
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
