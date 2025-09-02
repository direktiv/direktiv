package filesql

import (
	"context"
	"fmt"

	"github.com/direktiv/direktiv/pkg/filestore"
	"gorm.io/gorm"
)

type store struct {
	db *gorm.DB
}

func (s *store) ForRoot(rootID string) filestore.RootQuery {
	return &RootQuery{
		rootID:       rootID,
		db:           s.db,
		checksumFunc: filestore.DefaultCalculateChecksum,
	}
}

func (s *store) ForFile(file *filestore.File) filestore.FileQuery {
	return &FileQuery{
		file:         file,
		checksumFunc: filestore.DefaultCalculateChecksum,
		db:           s.db,
	}
}

var _ filestore.FileStore = &store{} // Ensures store struct conforms to filestore.FileStore interface.

func NewStore(db *gorm.DB) filestore.FileStore {
	return &store{
		db: db,
	}
}

func (s *store) CreateRoot(ctx context.Context, rootID string) (*filestore.Root, error) {
	n := &filestore.Root{
		ID: rootID,
	}
	res := s.db.WithContext(ctx).Table("filesystem_roots").Create(n)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpected gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return n, nil
}

func (s *store) GetRoot(ctx context.Context, id string) (*filestore.Root, error) {
	var list []filestore.Root
	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM filesystem_roots
					WHERE id = ?
					`, id).Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	if len(list) == 0 {
		return nil, filestore.ErrNotFound
	}

	return &list[0], nil
}

func (s *store) GetAllRoots(ctx context.Context) ([]*filestore.Root, error) {
	var list []filestore.Root
	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM filesystem_roots
					`).Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	ns := []*filestore.Root{}
	for i := range list {
		ns = append(ns, &list[i])
	}

	return ns, nil
}
