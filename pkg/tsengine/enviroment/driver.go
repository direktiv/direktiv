package enviroment

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/direktiv/direktiv/pkg/compiler"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/utils"
)

type Driver interface {
	SecretProvider
	FileWriter
	FileGetter
}

type SecretProvider interface {
	Get(ctx context.Context, namespace, name string) ([]byte, error)
}

type FileWriter interface {
	Write(namespace string, file compiler.File) error
}

type FileGetter interface {
	GetData(ctx context.Context, namespace, path string) ([]byte, error)
}

type DBBasedProvider struct {
	Secrets   datastore.SecretsStore
	Filestore filestore.FileStore
	FlowPath  string
	BaseFS    string
	FSShared  string
}

func (p *DBBasedProvider) Get(ctx context.Context, namespace, name string) ([]byte, error) {
	return p.Get(context.Background(), namespace, name)
}

func (p *DBBasedProvider) Write(namespace string, file compiler.File) error {
	fetchPath := file.Name
	if !filepath.IsAbs(file.Name) {
		fetchPath = filepath.Join(filepath.Dir(p.FlowPath), file.Name)
	}

	slog.Debug("fetching file", slog.String("file", fetchPath))

	f, err := p.Filestore.ForNamespace(namespace).GetFile(context.Background(), fetchPath)
	if err != nil {
		return err
	}
	data, err := p.Filestore.ForFile(f).GetData(context.Background())
	if err != nil {
		return err
	}

	newFile := filepath.Join(p.BaseFS, p.FSShared, filepath.Base(fetchPath))
	tf, err := os.Create(newFile)
	if err != nil {
		return err
	}

	_, err = tf.Write(data)
	return err
}

func (p *DBBasedProvider) GetData(ctx context.Context, namespace, path string) ([]byte, error) {
	file, err := p.Filestore.ForNamespace(namespace).GetFile(ctx, p.FlowPath)
	if err != nil {
		slog.Error("fetching file failed", slog.String("flow", p.FlowPath))
		return nil, err
	}
	data, err := p.Filestore.ForFile(file).GetData(ctx)
	if err != nil {
		return nil, err
	}

	return data, err

}

type FileBasedProvider struct {
	BaseFS string
}

func (p *FileBasedProvider) Get(ctx context.Context, namespace, name string) ([]byte, error) {
	secretFile := filepath.Join(p.BaseFS, "secrets", name)
	return os.ReadFile(secretFile)
}

func (p *FileBasedProvider) Write(namespace string, file compiler.File) error {
	filePathSrc := filepath.Join(p.BaseFS, file.Name)
	filePathTarget := filepath.Join(p.BaseFS, "shared", file.Name)
	_, err := utils.CopyFile(filePathSrc, filePathTarget)
	if err != nil {
		slog.Error("can not read file", slog.String("file", file.Name), slog.Any("error", err))
		return err
	}

	return nil
}

func (p *FileBasedProvider) GetData(ctx context.Context, namespace, path string) ([]byte, error) {
	return os.ReadFile(path)
}
