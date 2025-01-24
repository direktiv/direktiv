package filestoresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type store struct {
	db *gorm.DB
}

func (s *store) ForRootID(rootID uuid.UUID) filestore.RootQuery {
	return &RootQuery{
		rootID:       rootID,
		db:           s.db,
		checksumFunc: filestore.DefaultCalculateChecksum,
	}
}

func (s *store) ForNamespace(namespace string) filestore.RootQuery {
	return &RootQuery{
		namespace:    namespace,
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

func (s *store) CreateRoot(ctx context.Context, rootID uuid.UUID, namespace string) (*filestore.Root, error) {
	n := &filestore.Root{
		ID:        rootID,
		Namespace: namespace,
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

func (s *store) CreateTempRoot(ctx context.Context, rootID uuid.UUID) (*filestore.Root, error) {
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

func (s *store) GetRoot(ctx context.Context, id uuid.UUID) (*filestore.Root, error) {
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

func (s *store) GetRootByNamespace(ctx context.Context, namespace string) (*filestore.Root, error) {
	var list []filestore.Root
	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM filesystem_roots
					WHERE namespace = ?
					`, namespace).Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	if len(list) == 0 {
		return nil, filestore.ErrNotFound
	}

	return &list[0], nil
}

func (s *store) GetFileByID(ctx context.Context, id uuid.UUID) (*filestore.File, error) {
	file := &filestore.File{}
	res := s.db.WithContext(ctx).Raw(`
					SELECT id, root_id, path, depth, typ, created_at, updated_at, mime_type, length(data) AS size
					FROM filesystem_files
					WHERE id=?`, id).
		First(file)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("file '%s': %w", id, filestore.ErrNotFound)
		}

		return nil, res.Error
	}

	return file, nil
}
