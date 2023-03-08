package psql

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SQLFilestore struct {
	db *gorm.DB
}

var _ filestore.Root = &Root{}

func NewSQLFilestore(db *gorm.DB) *SQLFilestore {
	return &SQLFilestore{
		db: db,
	}
}

func NewMockFilestore() (*SQLFilestore, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&Root{}, &File{})
	if err != nil {
		return nil, err
	}
	fs := NewSQLFilestore(db)

	return fs, nil
}

//nolint:ireturn
func (s *SQLFilestore) CreateRoot(ctx context.Context, id uuid.UUID) (filestore.Root, error) {
	n := &Root{ID: id}
	res := s.db.WithContext(ctx).Create(n)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}
	n.db = s.db

	return n, nil
}

//nolint:ireturn
func (s *SQLFilestore) GetRoot(ctx context.Context, id uuid.UUID) (filestore.Root, error) {
	n := &Root{ID: id}
	res := s.db.WithContext(ctx).First(n)
	if res.Error != nil {
		return nil, res.Error
	}
	n.db = s.db

	return n, nil
}

//nolint:ireturn
func (s *SQLFilestore) GetAllRoots(ctx context.Context) ([]filestore.Root, error) {
	var list []Root
	res := s.db.WithContext(ctx).Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	var ns []filestore.Root
	for i := range list {
		list[i].db = s.db
		ns = append(ns, &list[i])
	}

	return ns, nil
}

var _ filestore.Filestore = &SQLFilestore{}
