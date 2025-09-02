package filestore

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"
)

// Package 'filestore' implements a filesystem that is responsible to store user's projects and files. For each
// direktiv namespace, a 'filestore.Root' should be created to host namespace files and directories.
// 'Root' interface provide all the methods needed to create direktiv namespace files and directories.
// Via 'filestore.FileStore' the caller manages the roots, and 'filestore.Root' the caller manages files and directories.

var (
	ErrFileTypeIsDirectory  = errors.New("ErrFileTypeIsDirectory")
	ErrNotFound             = errors.New("ErrNotFound")
	ErrPathAlreadyExists    = errors.New("ErrPathAlreadyExists")
	ErrNoParentDirectory    = errors.New("ErrNoParentDirectory")
	ErrInvalidPathParameter = errors.New("ErrInvalidPathParameter")
	ErrInvalidTypeParameter = errors.New("ErrInvalidTypeParameter")
)

// FileStore manages different operations on files and roots.

type FileStore interface {
	// CreateRoot creates a new root in the filestore.
	CreateRoot(ctx context.Context, id string) (*Root, error)

	GetRoot(ctx context.Context, id string) (*Root, error)

	// GetAllRoots list all roots.
	GetAllRoots(ctx context.Context) ([]*Root, error)

	ForRoot(id string) RootQuery

	// ForFile returns a query object to do further queries on that file.
	ForFile(file *File) FileQuery
}

// Root represents an isolated filesystems. Users of filestore can create and deletes multiple roots. In Direktiv,
// we create a dedicated root for every namespace.
type Root struct {
	ID string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// RootQuery performs different queries associated to a root.
type RootQuery interface {
	// GetFile grabs a file from the root.
	GetFile(ctx context.Context, path string) (*File, error)

	// CreateFile creates both files and directories,
	// param 'typ' indicates if file is of type directory or file.
	// Param 'path' should not already exist and the parent directory of 'path' should exist.
	// Param 'dataReader' should be nil when creating directories, and should be none nil when creating files.
	CreateFile(ctx context.Context, path string, typ FileType, mimeType string, data []byte) (*File, error)

	// ReadDirectory lists all files and directories in a path.
	ReadDirectory(ctx context.Context, path string) ([]*File, error)

	// Delete the root itself.
	Delete(ctx context.Context) error

	// ListAllFiles lists all files and directories in the filestore, this method used to help testing filestore logic.
	ListAllFiles(ctx context.Context) ([]*File, error)

	// ListDirektivFilesWithData lists all direktiv files(workflows and services, etc...) in the root.
	// It returns list of files with data fields already loaded, so the caller don't have to call GetData().
	ListDirektivFilesWithData(ctx context.Context) ([]*File, error)

	SetID(ctx context.Context, id string) error
}

// CalculateChecksumFunc is a function type used to calculate files checksums.
type CalculateChecksumFunc func([]byte) []byte

var Sha256CalculateChecksum CalculateChecksumFunc = func(data []byte) []byte {
	res := fmt.Sprintf("%x", sha256.Sum256(data))

	return []byte(res)
}

var DefaultCalculateChecksum = Sha256CalculateChecksum
