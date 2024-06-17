package tsengine

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/direktiv/direktiv/pkg/compiler"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/filestore/filestoresql"
	"github.com/direktiv/direktiv/pkg/utils"
	"github.com/fsnotify/fsnotify"
	"gorm.io/gorm"
)

type Environment interface {
	Init() error
}

type FileEnviroment struct {
	srcDir, flowPath string
	engine           *Engine
}

func NewFileEnviroment(srcDir, flowPath string, e *Engine) *FileEnviroment {
	return &FileEnviroment{
		srcDir:   srcDir,
		flowPath: flowPath,
		engine:   e,
	}
}

func (i *FileEnviroment) Init() error {
	slog.Info("reading flow")

	b, err := os.ReadFile(i.flowPath)
	if err != nil {
		return err
	}

	c, err := compiler.New(i.flowPath, string(b))
	if err != nil {
		return err
	}

	fi, err := c.CompileFlow()
	if err != nil {
		return err
	}

	secrets := make(map[string]string)

	// read secrets
	for a := range fi.Secrets {
		s := fi.Secrets[a]
		secretFile := filepath.Join(i.engine.baseFS, "secrets", s.Name)
		content, err := os.ReadFile(secretFile)
		if err != nil {
			slog.Error("can not read secret", slog.String("secret", s.Name), slog.Any("error", err))
			continue
		}
		secrets[s.Name] = string(content)
	}

	// read files
	for a := range fi.Files {
		file := fi.Files[a]
		if file.Scope == "shared" {
			filePathSrc := filepath.Join(i.engine.baseFS, file.Name)
			filePathTarget := filepath.Join(i.engine.baseFS, "shared", file.Name)
			_, err := utils.CopyFile(filePathSrc, filePathTarget)
			if err != nil {
				slog.Error("can not read file", slog.String("file", file.Name), slog.Any("error", err))
				continue
			}
		}
	}

	functions := make(map[string]string)

	for i := range fi.Functions {
		f := fi.Functions[i]
		functions[f.GetID()] = os.Getenv(f.GetID())
		slog.Info("adding function", slog.String("function", f.GetID()))
	}

	// files are already there
	i.engine.Initialize(c.Program, fi.Definition.State, secrets, functions, fi.Definition.Json)

	return nil
}

func (i *FileEnviroment) fileWatcher(flow string) {

	// dir to watch
	dir := filepath.Dir(flow)

	// file to watch
	file := filepath.Base(flow)

	slog.Info("watching flow", slog.String("flow", flow))

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	// listening for flow changes
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}
				if filepath.Base(event.Name) == file && event.Has(fsnotify.Write) {
					slog.Info("updating flow")
					i.Init()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				slog.Error("error occurred watching flow file:", slog.Any("error", err))
			}
		}
	}()

	err = watcher.Add(dir)
	if err != nil {
		panic(err)
	}

	<-make(chan struct{})

}

type DBEnviroment struct {
	dataStore           datastore.Store
	fileStore           filestore.FileStore
	flowPath, namespace string
	engine              *Engine
}

func NewDBEnviroment(srcDir, flowPath, namespace, secretKey string, db *gorm.DB, e *Engine) *DBEnviroment {
	secretKey = secretKey[0:16]
	ds := datastoresql.NewSQLStore(db, secretKey)
	fs := filestoresql.NewSQLFileStore(db)

	return &DBEnviroment{
		dataStore: ds,
		fileStore: fs,
		flowPath:  flowPath,
		namespace: namespace,
		engine:    e,
	}
}

func (db *DBEnviroment) Init() error {
	slog.Info("getting flow", slog.String("namespace", db.namespace), slog.String("path", db.flowPath))

	flow, err := db.fileStore.ForNamespace(db.namespace).GetFile(context.Background(), db.flowPath)
	if err != nil {
		slog.Error("fetching flow failed", slog.String("flow", db.flowPath))
		return err
	}

	data, err := db.fileStore.ForFile(flow).GetData(context.Background())
	if err != nil {
		return err
	}

	compiler, err := compiler.New(db.flowPath, string(data))
	if err != nil {
		return err
	}

	fi, err := compiler.CompileFlow()
	if err != nil {
		return err
	}

	secrets := make(map[string]string)
	for i := range fi.Secrets {
		secret := fi.Secrets[i]
		slog.Debug("fetching secret", slog.String("secret", secret.Name))
		s, err := db.dataStore.Secrets().Get(context.Background(), db.namespace, secret.Name)
		if err != nil {
			slog.Error("fetching secret failed", slog.String("secret", secret.Name))
			return err
		}
		secrets[secret.Name] = string(s.Data)
	}

	functions := make(map[string]string)
	for i := range fi.Functions {
		f := fi.Functions[i]
		// only do workflow functions
		if f.Image != "" {
			slog.Debug("adding function", slog.String("function", f.Image))
			functions[f.GetID()] = os.Getenv(f.GetID())
		}
	}

	for i := range fi.Files {
		file := fi.Files[i]

		if file.Scope != "shared" {
			continue
		}

		err = db.writeFile(file)
		if err != nil {
			slog.Error("can not fetch file", slog.String("file", file.Name), slog.Any("error", err))
			return err
		}
	}

	db.engine.Initialize(compiler.Program, fi.Definition.State, secrets, functions, fi.Definition.Json)

	return nil
}

func (db *DBEnviroment) writeFile(file compiler.File) error {

	fetchPath := file.Name
	if !filepath.IsAbs(file.Name) {
		fetchPath = filepath.Join(filepath.Dir(db.flowPath), file.Name)
	}

	slog.Debug("fetching file", slog.String("file", fetchPath))

	f, err := db.fileStore.ForNamespace(db.namespace).GetFile(context.Background(), fetchPath)
	if err != nil {
		return err
	}
	data, err := db.fileStore.ForFile(f).GetData(context.Background())
	if err != nil {
		return err
	}

	newFile := filepath.Join(db.engine.baseFS, engineFsShared, filepath.Base(fetchPath))
	tf, err := os.Create(newFile)
	if err != nil {
		return err
	}

	_, err = tf.Write(data)
	return err
}
