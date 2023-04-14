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
	ErrFileTypeIsDirectory = errors.New("ErrFileTypeIsDirectory")
	// TODO: fix this inconsistent error strings.
	ErrNotFound             = errors.New("not found")
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

type CalculateChecksumFunc func([]byte) []byte

var Sha256CalculateChecksum CalculateChecksumFunc = func(data []byte) []byte {
	res := fmt.Sprintf("%x", sha256.Sum256(data))

	return []byte(res)
}

var DefaultCalculateChecksum = Sha256CalculateChecksum
