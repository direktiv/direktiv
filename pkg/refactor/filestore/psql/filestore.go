package psql

import (
	"context"
	"errors"
	"fmt"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SQLFileStore struct {
	db *gorm.DB
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
		file:         file,
		checksumFunc: filestore.DefaultCalculateChecksum,
		db:           s.db,
	}
}

func (s *SQLFileStore) ForRevision(revision *filestore.Revision) filestore.RevisionQuery {
	return &RevisionQuery{
		rev: revision,
		db:  s.db,
	}
}

var _ filestore.FileStore = &SQLFileStore{} // Ensures SQLFileStore struct conforms to filestore.FileStore interface.

func NewSQLFileStore(db *gorm.DB) filestore.FileStore {
	return &SQLFileStore{
		db: db,
	}
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

//nolint:ireturn
func (s *SQLFileStore) GetFile(ctx context.Context, id uuid.UUID) (*filestore.File, error) {
	file := &filestore.File{}
	res := s.db.WithContext(ctx).
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
func (s *SQLFileStore) GetRevision(ctx context.Context, id uuid.UUID) (*filestore.File, *filestore.Revision, error) {
	// TODO: yassir, reimplement this function using JOIN so that it becomes a single query.
	rev := &filestore.Revision{}
	res := s.db.WithContext(ctx).
		Where("id", id).
		First(rev)
	if res.Error != nil {
		return nil, nil, res.Error
	}

	file := &filestore.File{}
	res = s.db.WithContext(ctx).
		Where("id", rev.FileID).
		First(file)
	if res.Error != nil {
		return nil, nil, res.Error
	}

	return file, rev, nil
}
