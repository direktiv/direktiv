package cmd

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	api2 "github.com/direktiv/direktiv/pkg/api"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/api"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/gateway"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/direktiv/direktiv/pkg/refactor/registry"
	"github.com/direktiv/direktiv/pkg/refactor/service"
	"github.com/direktiv/direktiv/pkg/util"
	"go.uber.org/zap"
)

func NewMain(config *core.Config, db *database.DB, pbus *pubsub.Bus, logger *zap.SugaredLogger, configureWorkflow func(data string) error) *sync.WaitGroup {
	initSLog()

	wg := &sync.WaitGroup{}

	go api2.RunApplication(config)

	done := make(chan struct{})

	// Create service manager
	serviceManager, err := service.NewManager(config, logger, config.EnableDocker)
	if err != nil {
		log.Fatalf("error creating service manager: %v\n", err)
	}

	// Setup GetServiceURL function
	service.SetupGetServiceURLFunc(config, config.EnableDocker)

	// Start service manager
	wg.Add(1)
	serviceManager.Start(done, wg)

	// Create registry manager
	registryManager, err := registry.NewManager(config.EnableDocker)
	if err != nil {
		log.Fatalf("error creating service manager: %v\n", err)
	}

	// Create endpoint manager
	gatewayManager := gateway.NewGatewayManager(db)

	// Create App
	app := core.App{
		Version: &core.Version{
			UnixTime: time.Now().Unix(),
		},
		Config:          config,
		ServiceManager:  serviceManager,
		RegistryManager: registryManager,
		GatewayManager:  gatewayManager,
		Bus:             pbus,
	}

	pbus.Subscribe(func(_ string) {
		renderServiceManager(db, serviceManager, logger)
	},
		pubsub.WorkflowCreate,
		pubsub.WorkflowUpdate,
		pubsub.WorkflowDelete,
		pubsub.WorkflowRename,
		pubsub.ServiceCreate,
		pubsub.ServiceUpdate,
		pubsub.ServiceDelete,
		pubsub.ServiceRename,
		pubsub.MirrorSync,
		pubsub.NamespaceDelete,
	)
	// Call at least once before booting
	renderServiceManager(db, serviceManager, logger)

	pbus.Subscribe(func(data string) {
		err := configureWorkflow(data)
		if err != nil {
			logger.Errorw("configure workflow", "error", err)
		}
	},
		pubsub.WorkflowCreate,
		pubsub.WorkflowUpdate,
		pubsub.WorkflowDelete,
	)

	// endpoint manager deletes routes/consumers on namespace delete
	pbus.Subscribe(func(ns string) {
		gatewayManager.DeleteNamespace(ns)
	},
		pubsub.NamespaceDelete,
	)

	// on sync redo all consumers and routes on sync or single file updates
	pbus.Subscribe(func(ns string) {
		gatewayManager.UpdateNamespace(ns)
	},
		pubsub.NamespaceCreate,
		pubsub.MirrorSync,
		pubsub.EndpointCreate,
		pubsub.EndpointUpdate,
		pubsub.EndpointDelete,
		pubsub.EndpointRename,
		pubsub.ConsumerCreate,
		pubsub.ConsumerDelete,
		pubsub.ConsumerUpdate,
		pubsub.ConsumerRename,
	)

	// initial loading of routes and consumers
	gatewayManager.UpdateAll()

	// TODO: yassir, this subscribe need to be removed when /api/v2/namespace delete endpoint is migrated.
	pbus.Subscribe(func(ns string) {
		err := registryManager.DeleteNamespace(ns)
		if err != nil {
			logger.Errorw("deleting registry namespace", "error", err)
		}
	},
		pubsub.NamespaceDelete,
	)

	// Start api v2 server
	wg.Add(1)
	api.Start(app, db, "0.0.0.0:6667", done, wg)

	go func() {
		// Listen for syscall signals for process to interrupt/quit
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		close(done)
	}()

	return wg
}

func initSLog() {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelInfo)

	logDebug := os.Getenv(util.DirektivDebug)
	if logDebug == "true" {
		lvl.Set(slog.LevelDebug)
	}

	slogger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	}))

	if os.Getenv(util.DirektivLogFormat) == "console" {
		slogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: lvl,
		}))
	}

	slog.SetDefault(slogger)
}

func renderServiceManager(db *database.DB, serviceManager core.ServiceManager, logger *zap.SugaredLogger) {
	logger = logger.With("subscriber", "services file watcher")

	fStore, dStore := db.FileStore(), db.DataStore()

	nsList, err := dStore.Namespaces().GetAll(context.Background())
	if err != nil {
		logger.Error("listing namespaces", "error", err)

		return
	}

	funConfigList := []*core.ServiceFileData{}

	for _, ns := range nsList {
		logger = logger.With("ns", ns.Name)
		files, err := fStore.ForNamespace(ns.Name).ListDirektivFilesWithData(context.Background())
		if err != nil {
			logger.Error("listing direktiv files", "error", err)

			continue
		}
		for _, file := range files {
			if file.Typ == filestore.FileTypeService {
				serviceDef, err := core.ParseServiceFile(file.Data)
				if err != nil {
					logger.Error("parse service file", "error", err)

					continue
				}

				funConfigList = append(funConfigList, &core.ServiceFileData{
					Typ:       core.ServiceTypeNamespace,
					Name:      "",
					Namespace: ns.Name,
					FilePath:  file.Path,
					Image:     serviceDef.Image,
					Cmd:       serviceDef.Cmd,
					Size:      serviceDef.Size,
					Scale:     serviceDef.Scale,
					Envs:      serviceDef.Envs,
					Patches:   serviceDef.Patches,
				})
			} else if file.Typ == filestore.FileTypeWorkflow {
				sub, err := getWorkflowFunctionDefinitionsFromWorkflow(ns, file)
				if err != nil {
					logger.Error("parse workflow def", "error", err)

					continue
				}

				funConfigList = append(funConfigList, sub...)
			}
		}
	}
	serviceManager.SetServices(funConfigList)
}

func getWorkflowFunctionDefinitionsFromWorkflow(ns *datastore.Namespace, f *filestore.File) ([]*core.ServiceFileData, error) {
	var wf model.Workflow

	err := wf.Load(f.Data)
	if err != nil {
		return nil, err
	}

	list := make([]*core.ServiceFileData, 0)

	for _, fn := range wf.Functions {
		if fn.GetType() != model.ReusableContainerFunctionType {
			// TODO: Alan, double check if continue here is valid.
			continue
		}

		serviceDef, ok := fn.(*model.ReusableFunctionDefinition)
		if !ok {
			return nil, errors.New("parse workflow def cast incorrectly")
		}

		list = append(list, &core.ServiceFileData{
			Typ:       core.ServiceTypeWorkflow,
			Name:      serviceDef.ID,
			Namespace: ns.Name,
			FilePath:  f.Path,
			Image:     serviceDef.Image,
			Cmd:       serviceDef.Cmd,
			Size:      serviceDef.Size.String(),
			Envs:      serviceDef.Envs,
			Patches:   serviceDef.Patches,
		})
	}

	return list, nil
}
