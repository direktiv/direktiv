package cmd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
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
	"github.com/direktiv/direktiv/pkg/metastore"
	"github.com/direktiv/direktiv/pkg/metastore/opensearchstore"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/direktiv/direktiv/pkg/registry"
	"github.com/direktiv/direktiv/pkg/service"
	"github.com/opensearch-project/opensearch-go"
)

type NewMainArgs struct {
	Config                       *core.Config
	Database                     *database.DB
	Metastore                    metastore.Store
	PubSubBus                    *pubsub.Bus
	ConfigureWorkflow            func(event *pubsub.FileSystemChangeEvent) error
	InstanceManager              *instancestore.InstanceManager
	WakeInstanceByEvent          events.WakeEventsWaiter
	WorkflowStart                events.WorkflowStart
	SyncNamespace                core.SyncNamespace
	RenderAllStartEventListeners func(ctx context.Context, tx *database.DB) error
}

func NewMain(circuit *core.Circuit, args *NewMainArgs) error {
	// Create service manager
	var err error
	var serviceManager core.ServiceManager
	if !args.Config.DisableServices {
		serviceManager, err = service.NewManager(args.Config)
		if err != nil {
			slog.Error("initializing service manager", "error", err)
			panic(err)
		}
		slog.Info("service manager initialized successfully")

		// Setup GetServiceURL function
		service.SetupGetServiceURLFunc(args.Config)

		circuit.Start(func() error {
			err := serviceManager.Run(circuit)
			if err != nil {
				return fmt.Errorf("service manager, err: %w", err)
			}

			return nil
		})
	} else {
		slog.Info("service manager is disabled")
	}

	// Create registry manager
	registryManager, err := registry.NewManager(args.Config.DisableServices)
	if err != nil {
		slog.Error("registry manager", "error", err)
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
		args.PubSubBus.Subscribe(&pubsub.FileSystemChangeEvent{}, func(_ string) {
			renderServiceManager(args.Database, serviceManager)
		})
		args.PubSubBus.Subscribe(&pubsub.NamespacesChangeEvent{}, func(_ string) {
			renderServiceManager(args.Database, serviceManager)
		})
		// Call at least once before booting
		renderServiceManager(args.Database, serviceManager)
	}

	args.PubSubBus.Subscribe(&pubsub.FileSystemChangeEvent{}, func(data string) {
		event := &pubsub.FileSystemChangeEvent{}
		err := json.Unmarshal([]byte(data), event)
		if err != nil {
			panic("Logic Error could not parse file system change event")
		}

		err = args.ConfigureWorkflow(event)
		if err != nil {
			slog.Error("configure workflow", "error", err)
		}
	})

	slog.Debug("Rendering event-listeners on server start")
	err = args.RenderAllStartEventListeners(circuit.Context(), args.Database)
	if err != nil {
		slog.Error("rendering event listener on server start", "error", err)
	}
	slog.Debug("Completed rendering event-listeners on server start")

	// endpoint manager
	args.PubSubBus.Subscribe(&pubsub.FileSystemChangeEvent{}, func(_ string) {
		helpers.RenderGatewayFiles(args.Database, gatewayManager2)
	})
	args.PubSubBus.Subscribe(&pubsub.NamespacesChangeEvent{}, func(_ string) {
		helpers.RenderGatewayFiles(args.Database, gatewayManager2)
	})
	// initial loading of routes and consumers
	helpers.RenderGatewayFiles(args.Database, gatewayManager2)
	if app.Config.OpenSearchInstalled {
		slog.Info("initialize OpenSearch", "config.OpenSearchHost", app.Config.OpenSearchHost, "config.OpenSearchPort", app.Config.OpenSearchPort, "config.OpenSearchProtocol", app.Config.OpenSearchProtocol)

		openSearchClient, err := initOpenSearch(app.Config)
		if err != nil {
			return fmt.Errorf("initialize OpenSearch client, err: %w", err)
		}
		slog.Info("connected to OpenSearch")
		meta, err := opensearchstore.NewMetaStore(circuit.Context(), openSearchClient, opensearchstore.Config{
			LogIndex:       "direktiv-logs",
			LogDeleteAfter: "7d",
			LogInit:        false,
		})
		if err != nil {
			return fmt.Errorf("initialize OpenSearch meta client, err: %w", err)
		}
		args.Metastore = meta
		// // Initialize log level based on config
		// lvl := new(slog.LevelVar)
		// lvl.Set(slog.LevelInfo)

		// if app.Config.LogDebug {
		// 	slog.Info("Logging is set to debug")
		// 	lvl.Set(slog.LevelDebug)
		// }

		// // Create a channel for logs and set up a worker to process it
		// logCh := make(chan metastore.LogEntry, 100)
		// worker := tracing.NewWorker(tracing.WorkerArgs{
		// 	LogCh:         logCh,
		// 	LogStore:      meta.LogStore(),
		// 	MaxBatchSize:  1,
		// 	FlushInterval: 1 * time.Millisecond,
		// 	CachedLevel:   int(lvl.Level()),
		// })

		// circuit.Start(func() error {
		// 	err := worker.Start(circuit)
		// 	if err != nil {
		// 		return fmt.Errorf("logs worker, err: %w", err)
		// 	}

		// 	return nil
		// })

		// // Create handlers
		// jsonHandler := tracing.NewContextHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		// 	Level: lvl,
		// }))
		// channelHandler := tracing.NewChannelHandler(logCh, nil, "default", lvl.Level())

		// // Combine handlers using a TeeHandler
		// compositeHandler := tracing.TeeHandler{
		// 	jsonHandler,
		// 	channelHandler,
		// }
		// // Set up the default logger
		// slogger := slog.New(compositeHandler)
		// slog.SetDefault(slogger)
		slog.Info("Metastore initialized")
	}
	// Start api v2 server
	err = api.Initialize(app, args.Database, args.Metastore, args.PubSubBus, args.InstanceManager, args.WakeInstanceByEvent, args.WorkflowStart, circuit)
	if err != nil {
		return fmt.Errorf("initializing api v2, err: %w", err)
	}
	slog.Info("api server v2 started")

	return nil
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

func initOpenSearch(cfg *core.Config) (*opensearch.Client, error) {
	retries := 40
	addr := cfg.OpenSearchProtocol + "://" + cfg.OpenSearchHost + ":" + strconv.Itoa(cfg.OpenSearchPort)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // nolint:gosec // Todo
		},
	}

	config := opensearch.Config{
		Addresses: []string{addr},
		Username:  cfg.OpenSearchUsername,
		Password:  cfg.OpenSearchPassword,
		Transport: transport,
		RetryBackoff: func(attempt int) time.Duration {
			return time.Second + time.Duration(attempt)
		},
		DisableRetry: false,
		MaxRetries:   retries,
	}

	client, err := opensearch.NewClient(config)
	if err != nil {
		slog.Info("connect to OpenSearch", "addr", addr, "error", err)
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}
	slog.Debug("connect to OpenSearch", "addr", addr)

	// Test the connection
	res, err := client.Info()
	if err != nil {
		slog.Info("OpenSearch connection test failed", "addr", addr, "error", err)
		return nil, fmt.Errorf("OpenSearch connection test failed: %w", err)
	}
	defer res.Body.Close()

	slog.Info("Connected to OpenSearch", "info", res.String())

	return client, nil
}
