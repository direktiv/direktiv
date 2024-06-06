package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/api"
	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/events"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/gateway"
	"github.com/direktiv/direktiv/pkg/helpers"
	"github.com/direktiv/direktiv/pkg/instancestore"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/direktiv/direktiv/pkg/registry"
	"github.com/direktiv/direktiv/pkg/service"
)

type NewMainArgs struct {
	Config                       *core.Config
	Database                     *database.SQLStore
	PubSubBus                    *pubsub.Bus
	ConfigureWorkflow            func(data string) error
	InstanceManager              *instancestore.InstanceManager
	WakeInstanceByEvent          events.WakeEventsWaiter
	WorkflowStart                events.WorkflowStart
	SyncNamespace                core.SyncNamespace
	RenderAllStartEventListeners func(ctx context.Context, tx *database.SQLStore) error
}

func NewMain(circuit *core.Circuit, args *NewMainArgs) error {
	initSLog()

	// Create service manager
	var err error
	var serviceManager core.ServiceManager
	if !args.Config.DisableServices {
		serviceManager, err = service.NewManager(args.Config)
		if err != nil {
			slog.Error("initializing service manager", "err", err)
			panic(err)
		}
		slog.Info("service manager initialized successfully")

		// Setup GetServiceURL function
		service.SetupGetServiceURLFunc(args.Config)

		circuit.Start(func() error {
			// TODO: yassir, Implement service crash handling.
			serviceManager.Start(circuit)

			return nil
		})
	} else {
		slog.Info("service manager is disabled")
	}

	// Create registry manager
	registryManager, err := registry.NewManager(args.Config.DisableServices)
	if err != nil {
		slog.Error("registry manager", "err", err)
		panic(err)
	}
	slog.Info("registry manager initialized successfully")

	// Create endpoint manager
	gatewayManager2 := gateway.NewManager(args.Database)
	slog.Info("gateway manager2 initialized successfully")

	// Create App
	app := core.App{
		Version: &core.Version{
			UnixTime: time.Now().Unix(),
		},
		Config:          args.Config,
		ServiceManager:  serviceManager,
		RegistryManager: registryManager,
		GatewayManager:  gatewayManager2,
		SyncNamespace:   args.SyncNamespace,
	}

	if !args.Config.DisableServices {
		args.PubSubBus.Subscribe(func(_ string) {
			renderServiceManager(args.Database, serviceManager)
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
	}

	args.PubSubBus.Subscribe(func(data string) {
		err := args.ConfigureWorkflow(data)
		if err != nil {
			slog.Error("configure workflow", "err", err)
		}
	},
		pubsub.WorkflowCreate,
		pubsub.WorkflowUpdate,
		pubsub.WorkflowDelete,
		pubsub.WorkflowRename,
	)

	slog.Debug("Rendering event-listeners on server start")
	err = args.RenderAllStartEventListeners(circuit.Context(), args.Database)
	if err != nil {
		slog.Error("rendering event listener on server start", "error", err)
	}
	slog.Debug("Completed rendering event-listeners on server start")

	// endpoint manager
	args.PubSubBus.Subscribe(func(_ string) {
		helpers.RenderGatewayFiles(args.Database, gatewayManager2)
	},
		pubsub.EndpointCreate,
		pubsub.EndpointUpdate,
		pubsub.EndpointDelete,
		pubsub.EndpointRename,
		pubsub.ConsumerCreate,
		pubsub.ConsumerDelete,
		pubsub.ConsumerUpdate,
		pubsub.ConsumerRename,
		pubsub.NamespaceDelete,
		pubsub.NamespaceCreate,
		pubsub.MirrorSync,
	)
	// initial loading of routes and consumers
	helpers.RenderGatewayFiles(args.Database, gatewayManager2)

	// TODO: yassir, this subscribe need to be removed when /api/v2/namespace delete endpoint is migrated.
	args.PubSubBus.Subscribe(func(ns string) {
		err := registryManager.DeleteNamespace(ns)
		if err != nil {
			slog.Error("deleting registry namespace", "err", err)
		}
	},
		pubsub.NamespaceDelete,
	)

	// Start api v2 server
	err = api.Initialize(app, args.Database, args.PubSubBus, args.InstanceManager, args.WakeInstanceByEvent, args.WorkflowStart, circuit)
	if err != nil {
		return fmt.Errorf("initializing api v2, err: %w", err)
	}
	slog.Info("api server v2 started")

	return nil
}

func initSLog() {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelInfo)

	logDebug := os.Getenv("DIREKTIV_DEBUG")
	if logDebug == "true" {
		lvl.Set(slog.LevelDebug)
	}

	slogger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	}))

	slog.SetDefault(slogger)
}

func renderServiceManager(db *database.SQLStore, serviceManager core.ServiceManager) {
	ctx := context.Background()
	slog := slog.With("subscriber", "services file watcher")

	fStore, dStore := db.FileStore(), db.DataStore()

	nsList, err := dStore.Namespaces().GetAll(ctx)
	if err != nil {
		slog.Error("listing namespaces", "err", err)

		return
	}

	funConfigList := []*core.ServiceFileData{}

	for _, ns := range nsList {
		slog = slog.With("namespace", ns.Name)
		files, err := fStore.ForNamespace(ns.Name).ListDirektivFilesWithData(ctx)
		if err != nil {
			slog.Error("listing direktiv files", "err", err)

			continue
		}
		for _, file := range files {
			if file.Typ == filestore.FileTypeService {
				serviceDef, err := core.ParseServiceFile(file.Data)
				if err != nil {
					slog.Error("parse service file", "err", err)

					continue
				}
				typ := core.ServiceTypeNamespace
				if ns.Name == core.SystemNamespace {
					typ = core.ServiceTypeSystem
				}
				funConfigList = append(funConfigList, &core.ServiceFileData{
					Typ:         typ,
					Name:        "",
					Namespace:   ns.Name,
					FilePath:    file.Path,
					ServiceFile: *serviceDef,
				})
			} else if file.Typ == filestore.FileTypeWorkflow {
				sub, err := getWorkflowFunctionDefinitionsFromWorkflow(ns, file)
				if err != nil {
					slog.Error("parse workflow def", "err", err)

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
