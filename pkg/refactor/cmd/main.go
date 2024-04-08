package cmd

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/direktiv/direktiv/pkg/refactor/registry"
	"github.com/direktiv/direktiv/pkg/refactor/service"
	"github.com/direktiv/direktiv/pkg/util"
)

type NewMainArgs struct {
	Config            *core.Config
	Database          *database.DB
	PubSubBus         *pubsub.Bus
	ConfigureWorkflow func(data string) error
	InstanceManager   *instancestore.InstanceManager
	SyncNamespace     core.SyncNamespace
}

func NewMain(args *NewMainArgs) *sync.WaitGroup {
	initSLog()
	wg := &sync.WaitGroup{}

	go func() {
		err := api2.RunApplication(args.Config)
		if err != nil {
			slog.Error("booting v1 api server", "error", err)
			os.Exit(1)
		}
	}()

	done := make(chan struct{})

	// Create service manager
	//nolint
	slog.Info("initializing service manager")
	serviceManager, err := service.NewManager(args.Config, args.Config.EnableDocker)
	if err != nil {
		slog.Error("Failed to unmarshal file change event for pub/sub notification", "error", err)
		os.Exit(1)
	}
	slog.Debug("Service manager initialized successfully.")

	// Setup GetServiceURL function
	service.SetupGetServiceURLFunc(args.Config, args.Config.EnableDocker)

	// Start service manager
	wg.Add(1)
	serviceManager.Start(done, wg)

	slog.Debug("Initializing registry manager.")
	// Create registry manager
	registryManager, err := registry.NewManager(args.Config.EnableDocker)
	if err != nil {
		slog.Error("Failed creating service manager", "error", err)
		os.Exit(1)
	}
	slog.Debug("Registry manager initialized successfully.")

	slog.Debug("Initializing Gateway manager.")
	// Create endpoint manager
	gatewayManager := gateway.NewGatewayManager(args.Database)
	slog.Debug("Gateway manager initialized.")

	// Create App
	app := core.App{
		Version: &core.Version{
			UnixTime: time.Now().Unix(),
		},
		Config:          args.Config,
		ServiceManager:  serviceManager,
		RegistryManager: registryManager,
		GatewayManager:  gatewayManager,
		SyncNamespace:   args.SyncNamespace,
	}
	slog.Debug("Setting up pub/sub subscription for service manager events.")
	args.PubSubBus.Subscribe(func(_ string) {
		renderServiceManager(args.Database, serviceManager)
		slog.Debug("Service Manager was triggered via pub/sub event.")
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
	renderServiceManager(args.Database, serviceManager)
	slog.Debug("Setting up pub/sub subscription for workflow configuration changes.")
	args.PubSubBus.Subscribe(func(data string) {
		err := args.ConfigureWorkflow(data)
		if err != nil {
			slog.Error("configure workflow", "error", err)
		}
		slog.Debug("Configuring a workflow triggered via pub/sub event.")
	},
		pubsub.WorkflowCreate,
		pubsub.WorkflowUpdate,
		pubsub.WorkflowDelete,
		pubsub.WorkflowRename,
	)
	slog.Debug("Setting up pub/sub subscription for gateway manager namespace updates.")
	// endpoint manager
	args.PubSubBus.Subscribe(func(ns string) {
		gatewayManager.UpdateNamespace(ns)
		slog.Debug("Updating namespace configurations based on pub/sub event.", "namespace", ns)
	},
		pubsub.NamespaceCreate,
		pubsub.MirrorSync,
	)
	// endpoint manager deletes routes/consumers on namespace delete
	args.PubSubBus.Subscribe(func(ns string) {
		gatewayManager.DeleteNamespace(ns)
		slog.Debug("Deleting namespace based on pub/sub event.", "namespace", ns)
	},
		pubsub.NamespaceDelete,
	)

	// on sync redo all consumers and routes on sync or single file updates
	args.PubSubBus.Subscribe(func(data string) {
		event := pubsub.FileChangeEvent{}
		err := json.Unmarshal([]byte(data), &event)
		if err != nil {
			slog.Error("unmarshal file change event", "error", err)
			os.Exit(1)
		}
		slog.Debug("Updating namespace configurations based on pub/sub event.", "namespace", event.Namespace)
		gatewayManager.UpdateNamespace(event.Namespace)
	},
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
	slog.Debug("Setting up pub/sub subscription for service manager events.")
	args.PubSubBus.Subscribe(func(ns string) {
		err := registryManager.DeleteNamespace(ns)
		if err != nil {
			slog.Error("deleting registry namespace", "error", err)
		}
	},
		pubsub.NamespaceDelete,
	)
	slog.Debug("Starting API V2 server.", "addr", "0.0.0.0:6667")

	// Start api v2 server
	wg.Add(1)
	api.Start(app, args.Database, args.PubSubBus, args.InstanceManager, "0.0.0.0:6667", done, wg)
	slog.Debug("API V2 server started.", "addr", "0.0.0.0:6667")

	go func() {
		// Listen for syscall signals for process to interrupt/quit
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		slog.Info("Shutdown signal received, initiating graceful shutdown.")
		close(done)
	}()
	slog.Debug("New Main application setup complete. Listening for shutdown signals.")

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

func renderServiceManager(db *database.DB, serviceManager core.ServiceManager) {
	ctx := context.Background()
	slog := slog.With("subscriber", "services file watcher")

	fStore, dStore := db.FileStore(), db.DataStore()

	nsList, err := dStore.Namespaces().GetAll(ctx)
	if err != nil {
		slog.Error("listing namespaces", "error", err)

		return
	}

	funConfigList := []*core.ServiceFileData{}

	for _, ns := range nsList {
		slog = slog.With("namespace", ns.Name)
		files, err := fStore.ForNamespace(ns.Name).ListDirektivFilesWithData(ctx)
		if err != nil {
			slog.Error("listing direktiv files", "error", err)

			continue
		}
		for _, file := range files {
			if file.Typ == filestore.FileTypeService {
				serviceDef, err := core.ParseServiceFile(file.Data)
				if err != nil {
					slog.Error("parse service file", "error", err)

					continue
				}

				funConfigList = append(funConfigList, &core.ServiceFileData{
					Typ:         core.ServiceTypeNamespace,
					Name:        "",
					Namespace:   ns.Name,
					FilePath:    file.Path,
					ServiceFile: *serviceDef,
				})
			} else if file.Typ == filestore.FileTypeWorkflow {
				sub, err := getWorkflowFunctionDefinitionsFromWorkflow(ns, file)
				if err != nil {
					slog.Error("parse workflow def", "error", err)

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

			ServiceFile: core.ServiceFile{
				Image:   serviceDef.Image,
				Cmd:     serviceDef.Cmd,
				Size:    serviceDef.Size.String(),
				Envs:    serviceDef.Envs,
				Patches: serviceDef.Patches,
			},
		})
	}

	return list, nil
}
