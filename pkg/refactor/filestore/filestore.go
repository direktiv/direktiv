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
// Via 'filestore.FileStore' the caller manages the roots, and 'filestore.Root' the caller manages files and directories.

var (
	ErrFileTypeIsDirectory = errors.New("ErrFileTypeIsDirectory")
	// TODO: fix this inconsistent error strings.
	ErrNotFound             = errors.New("not found")
	ErrPathAlreadyExists    = errors.New("ErrPathAlreadyExists")
	ErrNoParentDirectory    = errors.New("ErrNoParentDirectory")
	ErrInvalidPathParameter = errors.New("ErrInvalidPathParameter")
)

// FileStore manages different operations on files and roots.
type FileStore interface {
	// CreateRoot creates a new root in the filestore. For each direktiv
	CreateRoot(ctx context.Context, id uuid.UUID) (*Root, error)

	// GetAllRoots list all roots.
	GetAllRoots(ctx context.Context) ([]*Root, error)

	// ForRootID returns a query object to do further queries on root.
	ForRootID(rootID uuid.UUID) RootQuery

	// ForFile returns a query object to do further queries on that file.
	ForFile(file *File) FileQuery

	// ForRevision returns a query object to do further queries on that revision.
	ForRevision(revision *Revision) RevisionQuery

	// GetFile queries a file by id.
	GetFile(ctx context.Context, id uuid.UUID) (*File, error)

	// GetRevision queries a revision by id.
	GetRevision(ctx context.Context, id uuid.UUID) (*File, *Revision, error)
}

type Root struct {
	ID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
}

type RootQuery interface {
	GetFile(ctx context.Context, path string) (*File, error)

	// CreateFile creates both files and directories,
	// param 'typ' indicates if file is of type directory or file.
	// Param 'path' should not already exist and the parent directory of 'path' should exist.
	// Param 'dataReader' should be nil when creating directories, and should be none nil when creating files.
	CreateFile(ctx context.Context, path string, typ FileType, dataReader io.Reader) (*File, *Revision, error)
	ReadDirectory(ctx context.Context, path string) ([]*File, error)
	Delete(ctx context.Context) error
	CalculateChecksumsMap(ctx context.Context) (map[string]string, error)

	// IsEmptyDirectory returns true if path exist and of type directory and empty,
	// and false if path exist and of type directory and none empty.
	// If directory doesn't exist, it returns ErrNotFound.
	IsEmptyDirectory(ctx context.Context, path string) (bool, error)
	ListAllFiles(ctx context.Context) ([]*File, error)

	// CropFilesAndDirectories removes all files and directories that don't appear in excludePaths.
	CropFilesAndDirectories(ctx context.Context, excludePaths []string) error
}

// CalculateChecksumFunc is a function type used to calculate files checksums.
type CalculateChecksumFunc func([]byte) []byte

var Sha256CalculateChecksum CalculateChecksumFunc = func(data []byte) []byte {
	res := fmt.Sprintf("%x", sha256.Sum256(data))

	return []byte(res)
}

var DefaultCalculateChecksum = Sha256CalculateChecksum
