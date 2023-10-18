package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	api2 "github.com/direktiv/direktiv/pkg/api"
	"github.com/direktiv/direktiv/pkg/refactor/api"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/gateway"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/direktiv/direktiv/pkg/refactor/registry"
	"github.com/direktiv/direktiv/pkg/refactor/service"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
	"go.uber.org/zap"
	"golang.org/x/exp/slog"
)

func NewMain(config *core.Config, db *database.DB, pbus pubsub.Bus, logger *zap.SugaredLogger) *sync.WaitGroup {
	wg := &sync.WaitGroup{}

	go api2.RunApplication(config)

	done := make(chan struct{})

	// Create service manager
	serviceManager, err := service.NewManager(os.Getenv("DIREKITV_ENABLE_DOCKER") == "true")
	if err != nil {
		log.Fatalf("error creating service manager: %v\n", err)
	}
	// Start service manager
	wg.Add(1)
	serviceManager.Start(done, wg)

	// Create registry manager
	registryManager, err := registry.NewManager()
	if err != nil {
		log.Fatalf("error creating service manager: %v\n", err)
	}
	gatewayManager := gateway.NewHandler()
	// Create App
	app := &core.App{
		Version: &core.Version{
			UnixTime: time.Now().Unix(),
		},
		ServiceManager:  serviceManager,
		RegistryManager: registryManager,
		GatewayHandler:  gatewayManager,
	}

	pbus.Subscribe(func(_ string) {
		renderServiceManager(db, serviceManager, logger)
	},
		pubsub.WorkflowCreate,
		pubsub.WorkflowUpdate,
		pubsub.WorkflowDelete,
		pubsub.ServiceCreate,
		pubsub.ServiceUpdate,
		pubsub.ServiceDelete,
		pubsub.MirrorSync,
	)

	// Call at least once before booting
	renderServiceManager(db, serviceManager, logger)

	pbus.Subscribe(func(_ string) {
		renderGatewayManager(context.Background(), db, gatewayManager)
	},
		pubsub.WorkflowCreate,
		pubsub.WorkflowUpdate,
		pubsub.WorkflowDelete,
		pubsub.MirrorSync,
	)

	// Call at least once before booting
	renderGatewayManager(context.Background(), db, gatewayManager)

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

func renderGatewayManager(ctx context.Context, db *database.DB, gatewayManager *gateway.Handler) {
	fStore, dStore := db.FileStore(), db.DataStore()
	ns, err := dStore.Namespaces().GetByName(ctx, "gateway")
	if err != nil {
		slog.Error("listing namespaces", "error", err)

		return
	}

	files, err := fStore.ForNamespace(ns.Name).ListGatewayFiles(ctx)
	if err != nil {
		slog.Error("listing direktiv files", "error", err)
	}
	pluginroutes := make([]*gateway.RouteConfiguration, 0)
	for _, file := range files {
		data, err := fStore.ForFile(file).GetData(ctx)
		if err != nil {
			slog.Error("read file data", "error", err)

			continue
		}
		pluginroute, err := spec.ParsePluginRouteFile(data)
		if err != nil {
			slog.Error("parse service file", "error", err)

			continue
		}

		r := gateway.RouteConfiguration{
			Path:           pluginroute.Path,
			Method:         pluginroute.Method,
			TimeoutSeconds: pluginroute.TimeoutSeconds,
			Targets:        plugins.Targets{},
			PluginsConfig:  make([]plugins.Configuration, 0),
		}
		for _, v := range pluginroute.PluginsConfig {
			r.PluginsConfig = append(r.PluginsConfig, plugins.Configuration{
				Name:          v.Name,
				Version:       v.Version,
				RuntimeConfig: v.RuntimeConfig,
			})
		}
		pluginroutes = append(pluginroutes, &r)
	}
	gatewayManager.SetRoutes(pluginroutes)
}

func renderServiceManager(db *database.DB, serviceManager *service.Manager, logger *zap.SugaredLogger) {
	logger = logger.With("subscriber", "services file watcher")

	fStore, dStore := db.FileStore(), db.DataStore()

	nsList, err := dStore.Namespaces().GetAll(context.Background())
	if err != nil {
		logger.Error("listing namespaces", "error", err)

		return
	}

	funConfigList := []*service.Config{}

	for _, ns := range nsList {
		logger = logger.With("ns", ns.Name)
		files, err := fStore.ForNamespace(ns.Name).ListDirektivFiles(context.Background())
		if err != nil {
			logger.Error("listing direktiv files", "error", err)

			continue
		}
		for _, file := range files {
			data, err := fStore.ForFile(file).GetData(context.Background())
			if err != nil {
				logger.Error("read file data", "error", err)

				continue
			}
			if file.Typ == filestore.FileTypeService {
				serviceDef, err := spec.ParseServicesFile(data)
				if err != nil {
					logger.Error("parse service file", "error", err)

					continue
				}
				funConfigList = append(funConfigList, &service.Config{
					Typ:       "namespace-service",
					Name:      serviceDef.Name,
					Namespace: ns.Name,
					FilePath:  file.Path,
					Image:     serviceDef.Image,
					CMD:       serviceDef.Cmd,
					Size:      serviceDef.Size,
					Scale:     serviceDef.Scale,
				})
			} else if file.Typ == filestore.FileTypeWorkflow {
				serviceDef, err := spec.ParseWorkflowServiceDefinition(data)
				if err != nil {
					logger.Error("parse workflow service def", "error", err)

					continue
				}
				if serviceDef.Typ == "knative-workflow" {
					funConfigList = append(funConfigList, &service.Config{
						Typ:       "workflow-service",
						Name:      serviceDef.Name,
						Namespace: ns.Name,
						FilePath:  file.Path,
						Image:     serviceDef.Image,
						CMD:       serviceDef.Cmd,
						Size:      serviceDef.Size,
						Scale:     serviceDef.Scale,
					})
				}
			}
		}
	}
	serviceManager.SetServices(funConfigList)
}
