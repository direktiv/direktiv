package environment

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/direktiv/direktiv/pkg/compiler"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/utils"
)

// Driver is an interface that encompasses SecretProvider, FileWriter, and FileGetter.
type Driver interface {
	SecretProvider
	FileWriter
	FileGetter
	FunctionProvider
}

// SecretProvider defines the method to get secrets.
type SecretProvider interface {
	GetSecret(ctx context.Context, namespace, name string) ([]byte, error)
}

// FileWriter defines the method to write files.
type FileWriter interface {
	WriteFile(ctx context.Context, namespace string, file compiler.File) error
}

// FileGetter defines the method to get file data.
type FileGetter interface {
	GetFileData(ctx context.Context, namespace, path string) ([]byte, error)
}

// FunctionProvider is an interface for retrieving functions.
type FunctionProvider interface {
	GetFunction(ctx context.Context, functionID string) (string, error)
}

// DBBasedProvider implements the Driver interface using a database-based backend.
type DBBasedProvider struct {
	SecretsStore datastore.SecretsStore
	FileStore    filestore.FileStore
	FlowFilePath string
	BaseFilePath string
	SharedFSPath string
}

// GetSecret retrieves a secret from the datastore.
func (p *DBBasedProvider) GetSecret(ctx context.Context, namespace, name string) ([]byte, error) {
	secret, err := p.SecretsStore.Get(ctx, namespace, name)
	if err != nil {
		slog.Error("failed to get secret", slog.String("namespace", namespace), slog.String("name", name), slog.Any("error", err))
		return nil, fmt.Errorf("failed to get secret %s/%s: %w", namespace, name, err)
	}
	return secret.Data, nil
}

// WriteFile writes a file to the filestore.
func (p *DBBasedProvider) WriteFile(ctx context.Context, namespace string, file compiler.File) error {
	fetchPath := file.Name
	if !filepath.IsAbs(file.Name) {
		fetchPath = filepath.Join(filepath.Dir(p.FlowFilePath), file.Name)
	}

	slog.Debug("fetching file", slog.String("file", fetchPath))

	fileHandle, err := p.FileStore.ForNamespace(namespace).GetFile(ctx, fetchPath)
	if err != nil {
		slog.Error("failed to fetch file", slog.String("namespace", namespace), slog.String("file", fetchPath), slog.Any("error", err))
		return fmt.Errorf("failed to fetch file %s: %w", fetchPath, err)
	}

	data, err := p.FileStore.ForFile(fileHandle).GetData(ctx)
	if err != nil {
		slog.Error("failed to get data for file", slog.String("namespace", namespace), slog.String("file", fetchPath), slog.Any("error", err))
		return fmt.Errorf("failed to get data for file %s: %w", fetchPath, err)
	}

	newFilePath := filepath.Join(p.BaseFilePath, p.SharedFSPath, filepath.Base(fetchPath))
	targetFile, err := os.Create(newFilePath)
	if err != nil {
		slog.Error("failed to create new file", slog.String("path", newFilePath), slog.Any("error", err))
		return fmt.Errorf("failed to create new file %s: %w", newFilePath, err)
	}
	defer targetFile.Close()

	_, err = targetFile.Write(data)
	if err != nil {
		slog.Error("failed to write data to file", slog.String("file", newFilePath), slog.Any("error", err))
		return fmt.Errorf("failed to write data to file %s: %w", newFilePath, err)
	}

	return nil
}

// GetFileData retrieves file data from the filestore.
func (p *DBBasedProvider) GetFileData(ctx context.Context, namespace, path string) ([]byte, error) {
	fileHandle, err := p.FileStore.ForNamespace(namespace).GetFile(ctx, path)
	if err != nil {
		slog.Error("failed to fetch file", slog.String("namespace", namespace), slog.String("path", path), slog.Any("error", err))
		return nil, fmt.Errorf("failed to fetch file %s: %w", path, err)
	}

	data, err := p.FileStore.ForFile(fileHandle).GetData(ctx)
	if err != nil {
		slog.Error("failed to get data for file", slog.String("namespace", namespace), slog.String("path", path), slog.Any("error", err))
		return nil, fmt.Errorf("failed to get data for file %s: %w", path, err)
	}

	return data, nil
}

// GetFunction retrieves the function value from the database.
func (p *DBBasedProvider) GetFunction(ctx context.Context, functionID string) (string, error) {
	return os.Getenv(functionID), nil
}

// FileBasedProvider implements the Driver interface using a file-based backend.
type FileBasedProvider struct {
	BaseFilePath string
}

// GetSecret retrieves a secret from the file system.
func (p *FileBasedProvider) GetSecret(ctx context.Context, namespace, name string) ([]byte, error) {
	secretFilePath := filepath.Join(p.BaseFilePath, "secrets", name)
	data, err := os.ReadFile(secretFilePath)
	if err != nil {
		slog.Error("failed to read secret file", slog.String("file", secretFilePath), slog.Any("error", err))
		return nil, fmt.Errorf("failed to read secret file %s: %w", secretFilePath, err)
	}
	return data, nil
}

// WriteFile writes a file to the shared filesystem.
func (p *FileBasedProvider) WriteFile(ctx context.Context, namespace string, file compiler.File) error {
	sourceFilePath := filepath.Join(p.BaseFilePath, file.Name)
	targetFilePath := filepath.Join(p.BaseFilePath, "shared", file.Name)
	_, err := utils.CopyFile(sourceFilePath, targetFilePath)
	if err != nil {
		slog.Error("failed to copy file", slog.String("source", sourceFilePath), slog.String("target", targetFilePath), slog.Any("error", err))
		return fmt.Errorf("failed to copy file from %s to %s: %w", sourceFilePath, targetFilePath, err)
	}

	return nil
}

// GetFileData retrieves file data from the file system.
func (p *FileBasedProvider) GetFileData(ctx context.Context, namespace, path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("failed to read file", slog.String("path", path), slog.Any("error", err))
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return data, nil
}

// GetFunction retrieves the function value from an environment variable.
func (p *FileBasedProvider) GetFunction(ctx context.Context, functionID string) (string, error) {
	return os.Getenv(functionID), nil
}
