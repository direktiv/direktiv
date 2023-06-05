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

func (s *sqlFileStore) CreateRoot(ctx context.Context, id uuid.UUID) (*filestore.Root, error) {
	n := &filestore.Root{ID: id}
	res := s.db.WithContext(ctx).Table("filesystem_roots").Create(n)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("unexpedted gorm create count, got: %d, want: %d", res.RowsAffected, 1)
	}

	return n, nil
}

//nolint:ireturn
func (s *sqlFileStore) GetAllRoots(ctx context.Context) ([]*filestore.Root, error) {
	var list []filestore.Root
	res := s.db.WithContext(ctx).Raw(`
					SELECT *
					FROM filesystem_roots;
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

//nolint:ireturn
func (s *sqlFileStore) GetFile(ctx context.Context, id uuid.UUID) (*filestore.File, error) {
	file := &filestore.File{}
	res := s.db.WithContext(ctx).Table("filesystem_files").
		Where("id", id).
		First(file)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("file '%s': %w", id, filestore.ErrNotFound)
		}

		return nil, res.Error
	}

	return file, nil
}

//nolint:ireturn
func (s *sqlFileStore) GetRevision(ctx context.Context, id uuid.UUID) (*filestore.File, *filestore.Revision, error) {
	// TODO: yassir, reimplement this function using JOIN so that it becomes a single query.
	rev := &filestore.Revision{}
	res := s.db.WithContext(ctx).Table("filesystem_revisions").
		Where("id", id).
		First(rev)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("revision '%s': %w", id, filestore.ErrNotFound)
		}

		return nil, nil, res.Error
	}

	file := &filestore.File{}
	res = s.db.WithContext(ctx).Table("filesystem_files").
		Where("id", rev.FileID).
		First(file)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("file '%s': %w", rev.FileID, filestore.ErrNotFound)
		}

		return nil, nil, res.Error
	}

	return file, rev, nil
}
