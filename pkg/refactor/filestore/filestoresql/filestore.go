package filestoresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sqlFileStore struct {
	db *gorm.DB
}

func (s *sqlFileStore) ForRootID(rootID uuid.UUID) filestore.RootQuery {
	return &RootQuery{
		rootID:       rootID,
		db:           s.db,
		checksumFunc: filestore.DefaultCalculateChecksum,
	}
}

func (s *sqlFileStore) ForRootNamespaceID(nsID uuid.UUID) filestore.RootQuery {
	return &RootQuery{
		nsID:         nsID,
		db:           s.db,
		checksumFunc: filestore.DefaultCalculateChecksum,
	}
}

func (s *sqlFileStore) ForFile(file *filestore.File) filestore.FileQuery {
	return &FileQuery{
		file:         file,
		checksumFunc: filestore.DefaultCalculateChecksum,
		db:           s.db,
	}
}

func (s *sqlFileStore) ForRevision(revision *filestore.Revision) filestore.RevisionQuery {
	return &RevisionQuery{
		rev: revision,
		db:  s.db,
	}
}

var _ filestore.FileStore = &sqlFileStore{} // Ensures sqlFileStore struct conforms to filestore.FileStore interface.

func NewSQLFileStore(db *gorm.DB) filestore.FileStore {
	return &sqlFileStore{
		db: db,
	}
}

func (s *sqlFileStore) CreateRoot(ctx context.Context, rootID, namespaceID uuid.UUID) (*filestore.Root, error) {
	n := &filestore.Root{
		ID:          rootID,
		NamespaceID: namespaceID,
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

func (s *sqlFileStore) CreateTempRoot(ctx context.Context, rootID uuid.UUID) (*filestore.Root, error) {
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

func (s *sqlFileStore) GetRoot(ctx context.Context, id uuid.UUID) (*filestore.Root, error) {
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

func (s *sqlFileStore) GetAllRoots(ctx context.Context) ([]*filestore.Root, error) {
	var list []filestore.Root
	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM filesystem_roots
					`).Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	var ns []*filestore.Root
	for i := range list {
		ns = append(ns, &list[i])
	}

	return ns, nil
}

func (s *sqlFileStore) GetRootByNamespaceID(ctx context.Context, namespaceID uuid.UUID) (*filestore.Root, error) {
	var list []filestore.Root
	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM filesystem_roots
					WHERE namespace_id = ?
					`, namespaceID).Find(&list)
	if res.Error != nil {
		return nil, res.Error
	}

	if len(list) == 0 {
		return nil, filestore.ErrNotFound
	}

	return &list[0], nil
}

func (s *sqlFileStore) GetFileByID(ctx context.Context, id uuid.UUID) (*filestore.File, error) {
	file := &filestore.File{}
	res := s.db.WithContext(ctx).Raw(`
					SELECT *
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

func (s *sqlFileStore) GetRevisionByID(ctx context.Context, id uuid.UUID) (*filestore.File, *filestore.Revision, error) {
	// TODO: yassir, reimplement this function using JOIN so that it becomes a single query.
	rev := &filestore.Revision{}
	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM filesystem_revisions
					WHERE id=?`, id).
		First(rev)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("revision '%s': %w", id, filestore.ErrNotFound)
		}

		return nil, nil, res.Error
	}

	file := &filestore.File{}
	res = s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM filesystem_files
					WHERE id=?`, rev.FileID).
		First(file)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("file '%s': %w", rev.FileID, filestore.ErrNotFound)
		}

		return nil, nil, res.Error
	}

	return file, rev, nil
}
