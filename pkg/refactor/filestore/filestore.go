package filestore

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

// Package 'filestore' implements a filesystem that is responsible to store user's projects and files. For each
// direktiv namespace, a 'filestore.Root' should be created to host namespace files and directories.
// 'Root' interface provide all the methods needed to create direktiv namespace files and directories.
// Via 'filestore.Manager' the caller manages the roots, and 'filestore.Root' the caller manages files and directories.

var (
	ErrFileTypeIsDirectory  = errors.New("ErrFileTypeIsDirectory")
	ErrNotFound             = errors.New("ErrNotFound")
	ErrPathAlreadyExists    = errors.New("ErrPathAlreadyExists")
	ErrNoParentDirectory    = errors.New("ErrNoParentDirectory")
	ErrInvalidPathParameter = errors.New("ErrInvalidPathParameter")
)

type FileStore interface {
	CreateRoot(ctx context.Context, id uuid.UUID) (*Root, error)
	GetAllRoots(ctx context.Context) ([]*Root, error)

	ForRootID(rootID uuid.UUID) RootQuery
	ForFile(file *File) FileQuery
	ForRevision(revision *Revision) RevisionQuery

	GetFile(ctx context.Context, id uuid.UUID) (*File, error)
	GetRevision(ctx context.Context, id uuid.UUID) (*File, *Revision, error)
}

type Root struct {
	ID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
}

type RootQuery interface {
	GetFile(ctx context.Context, path string) (*File, error)
	CreateFile(ctx context.Context, path string, typ FileType, dataReader io.Reader) (*File, *Revision, error)
	ReadDirectory(ctx context.Context, path string) ([]*File, error)
	Delete(ctx context.Context) error
	CalculateChecksumsMap(ctx context.Context, path string) (map[string]string, error)
	IsEmptyDirectory(ctx context.Context, path string) (bool, error)
	ListAllFiles(ctx context.Context) ([]*File, error)

	// BulkRemoveFilesWithExclude removes all files and directories that don't appear in excludePaths.
	BulkRemoveFilesWithExclude(ctx context.Context, excludePaths []string) error
}

type CalculateChecksumFunc func([]byte) []byte

var Sha256CalculateChecksum CalculateChecksumFunc = func(data []byte) []byte {
	res := fmt.Sprintf("%x", sha256.Sum256(data))

	return []byte(res)
}

var DefaultCalculateChecksum = Sha256CalculateChecksum
