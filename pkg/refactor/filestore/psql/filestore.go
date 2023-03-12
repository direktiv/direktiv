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
	//nolint:containedctx
	ctx context.Context
}

func (s *SQLFilestore) ForRoot(root filestore.Root) filestore.RootQuery {
	return &RootQuery{
		root: root,
		db:   s.db,
		ctx:  s.ctx,
	}
}

func (s *SQLFilestore) ForFile(file filestore.File) filestore.FileQuery {
	return &FileQuery{
		file: file,
		db:   s.db,
		ctx:  s.ctx,
	}
}

func (s *SQLFilestore) ForRevision(revision filestore.Revision) filestore.RevisionQuery {
	return &RevisionQuery{
		rev: revision,
		db:  s.db,
		ctx: s.ctx,
	}
}

func (s *SQLFilestore) WithContext(ctx context.Context) filestore.Filestore {
	s.ctx = ctx

	return s
}

var _ filestore.Filestore = &SQLFilestore{} // Ensures SQLFilestore struct conforms to filestore.Filestore interface.

func NewSQLFilestore(db *gorm.DB) *SQLFilestore {
	return &SQLFilestore{
		db:  db,
		ctx: context.Background(),
	}
}

func NewMockFilestore() (*SQLFilestore, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&Root{}, &File{}, &Revision{})
	if err != nil {
		return nil, err
	}
	fs := NewSQLFilestore(db)

	return fs, nil
}

func (s *SQLFilestore) CreateRoot(id uuid.UUID) (filestore.Root, error) {
	n := &Root{ID: id}
	res := s.db.WithContext(s.ctx).Create(n)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return n, nil
}

func (s *SQLFilestore) GetRoot(id uuid.UUID) (filestore.Root, error) {
	n := &Root{ID: id}
	res := s.db.WithContext(s.ctx).First(n)
	if res.Error != nil {
		return nil, res.Error
	}

	return n, nil
}

//nolint:ireturn
func (s *SQLFilestore) GetAllRoots() ([]filestore.Root, error) {
	var list []Root
	res := s.db.WithContext(s.ctx).Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	var ns []filestore.Root
	for i := range list {
		ns = append(ns, &list[i])
	}

	return ns, nil
}
