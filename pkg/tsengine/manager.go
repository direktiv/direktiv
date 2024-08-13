package tsengine

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/direktiv/direktiv/pkg/compiler"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/filestore/filestoresql"
	"github.com/direktiv/direktiv/pkg/instancestore"
	"github.com/direktiv/direktiv/pkg/instancestore/instancestoresql"
	"github.com/direktiv/direktiv/pkg/runtime"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// type InitData struct {
// 	Program            *goja.Program
// 	Definition         *compiler.Definition
// 	Secrets, Functions map[string]string
// }

// type Manager interface {
// 	Init() (InitData, error)
// 	CreateInstance() error
// }

type LocalManager struct {
	srcDir, flowPath string
}

func NewLocalManager(srcDir, flowPath string) *LocalManager {
	return &LocalManager{
		srcDir:   srcDir,
		flowPath: flowPath,
	}
}

func (i *LocalManager) CreateInstance() error {
	return nil
}

func (i *LocalManager) Init() error {
	slog.Info("reading flow")

	// b, err := os.ReadFile(i.flowPath)
	// if err != nil {
	// 	return err
	// }

	// c, err := compiler.New(i.flowPath, string(b))
	// if err != nil {
	// 	return err
	// }

	// fi, err := c.CompileFlow()
	// if err != nil {
	// 	return err
	// }

	// secrets := make(map[string]string)

	// // read secrets
	// for a := range fi.Secrets {
	// 	s := fi.Secrets[a]
	// 	secretFile := filepath.Join(i.engine.baseFS, "secrets", s.Name)
	// 	content, err := os.ReadFile(secretFile)
	// 	if err != nil {
	// 		slog.Error("can not read secret", slog.String("secret", s.Name), slog.Any("error", err))
	// 		continue
	// 	}
	// 	secrets[s.Name] = string(content)
	// }

	// // read files
	// for a := range fi.Files {
	// 	file := fi.Files[a]
	// 	if file.Scope == "shared" {
	// 		filePathSrc := filepath.Join(i.engine.baseFS, file.Name)
	// 		filePathTarget := filepath.Join(i.engine.baseFS, "shared", file.Name)
	// 		_, err := utils.CopyFile(filePathSrc, filePathTarget)
	// 		if err != nil {
	// 			slog.Error("can not read file", slog.String("file", file.Name), slog.Any("error", err))
	// 			continue
	// 		}
	// 	}
	// }

	// functions := make(map[string]string)

	// for i := range fi.Functions {
	// 	f := fi.Functions[i]
	// 	functions[f.GetID()] = os.Getenv(f.GetID())
	// 	slog.Info("adding function", slog.String("function", f.GetID()))
	// }

	// // files are already there
	// i.engine.Initialize(c.Program, fi.Definition.State, secrets, functions, fi.Definition.Json)

	return nil
}

func (i *LocalManager) fileWatcher(flow string) {

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

type DirektivManager struct {
	dataStore     datastore.Store
	fileStore     filestore.FileStore
	instanceStore instancestore.Store

	flowPath string
	baseDir  string

	namespace   string
	namespaceID uuid.UUID

	runtimeData *runtime.Data
}

func NewDirektivManager(baseDir, flowPath, namespace, secretKey string, db *gorm.DB) (*DirektivManager, error) {
	secretKey = secretKey[0:16]

	ds := datastoresql.NewSQLStore(db, secretKey)
	fs := filestoresql.NewSQLFileStore(db)
	is := instancestoresql.NewSQLInstanceStore(db)

	dm := &DirektivManager{
		dataStore:     ds,
		fileStore:     fs,
		instanceStore: is,
		flowPath:      flowPath,
		namespace:     namespace,
		baseDir:       baseDir,
	}

	err := dm.init()
	return dm, err
}

func (dm *DirektivManager) RuntimeData() *runtime.Data {
	return dm.runtimeData
}

func (dm *DirektivManager) CreateInstance(id uuid.UUID, invoker, definition string) error {
	fmt.Println("CREATE INSTANCE!!!!")

	instanceData := &instancestore.CreateInstanceDataArgs{
		ID:           id,
		NamespaceID:  dm.namespaceID,
		Namespace:    dm.namespace,
		WorkflowPath: dm.flowPath,
		// RootInstanceID: "",
		// Server: ,
		Invoker:       invoker,
		Definition:    []byte(base64.StdEncoding.EncodeToString([]byte(definition))),
		Input:         []byte{},
		LiveData:      []byte{},
		TelemetryInfo: []byte{},
		DescentInfo:   []byte{},
		RuntimeInfo:   []byte{},
		ChildrenInfo:  []byte{},
		// SyncHash: ,
	}

	fmt.Println(instanceData)

	_, err := dm.instanceStore.CreateInstanceData(context.Background(), instanceData)

	fmt.Println("ERRR")
	fmt.Println(err)

	return err
}

func (dm *DirektivManager) init() error {
	slog.Info("getting flow", slog.String("namespace", dm.namespace), slog.String("path", dm.flowPath))

	runtimeData := runtime.Data{}

	// get namespace
	ns, err := dm.dataStore.Namespaces().GetByName(context.Background(), dm.namespace)
	if err != nil {
		slog.Error("fetching namespace information", slog.Any("namespace", dm.namespace))
		return err
	}
	dm.namespaceID = ns.ID

	flow, err := dm.fileStore.ForNamespace(dm.namespace).GetFile(context.Background(), dm.flowPath)
	if err != nil {
		slog.Error("fetching flow failed", slog.String("flow", dm.flowPath))
		return err
	}

	data, err := dm.fileStore.ForFile(flow).GetData(context.Background())
	if err != nil {
		return err
	}

	compiler, err := compiler.New(dm.flowPath, string(data))
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
		s, err := dm.dataStore.Secrets().Get(context.Background(), dm.namespace, secret.Name)
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

		err = dm.writeFile(file)
		if err != nil {
			slog.Error("can not fetch file", slog.String("file", file.Name), slog.Any("error", err))
			return err
		}
	}

	runtimeData.Functions = functions
	runtimeData.Secrets = secrets
	runtimeData.Definition = fi.Definition
	runtimeData.Program = compiler.Program
	runtimeData.Script = string(data)

	dm.runtimeData = &runtimeData

	return nil
}

func (db *DirektivManager) writeFile(file compiler.File) error {

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

	newFile := filepath.Join(db.baseDir, engineFsShared, filepath.Base(fetchPath))
	tf, err := os.Create(newFile)
	if err != nil {
		return err
	}

	_, err = tf.Write(data)
	return err
}
