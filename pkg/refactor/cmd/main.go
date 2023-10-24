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
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/direktiv/direktiv/pkg/refactor/registry"
	"github.com/direktiv/direktiv/pkg/refactor/service"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
	"go.uber.org/zap"
)

func NewMain(config *core.Config, db *database.DB, pbus pubsub.Bus, logger *zap.SugaredLogger) *sync.WaitGroup {
	wg := &sync.WaitGroup{}

	go api2.RunApplication(config)

	done := make(chan struct{})

	isDockerEnabled := os.Getenv("DIREKITV_ENABLE_DOCKER") == "true"

	// Create service manager
	serviceManager, err := service.NewManager(config, logger, isDockerEnabled)
	if err != nil {
		log.Fatalf("error creating service manager: %v\n", err)
	}

	// Setup GetServiceURL function
	service.SetupGetServiceURLFunc(config, isDockerEnabled)

	// Start service manager
	wg.Add(1)
	serviceManager.Start(done, wg)

	// Create registry manager
	registryManager, err := registry.NewManager(isDockerEnabled)
	if err != nil {
		log.Fatalf("error creating service manager: %v\n", err)
	}

	// Create gateway manager
	gw := gateway.NewHandler()
	// Start service manager
	wg.Add(1)
	gw.Start(done, wg)

	// Create App
	app := core.App{
		Version: &core.Version{
			UnixTime: time.Now().Unix(),
		},
		Config:          config,
		ServiceManager:  serviceManager,
		RegistryManager: registryManager,
		GatewayManager:  gw,
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
		renderGatewayManager(db, gw, logger)
	},
		pubsub.EndpointCreate,
		pubsub.EndpointUpdate,
		pubsub.EndpointDelete,
		pubsub.MirrorSync,
	)
	renderGatewayManager(db, gw, logger)

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

func renderGatewayManager(db *database.DB, gwManager core.GatewayManager, logger *zap.SugaredLogger) {
	fStore, dStore := db.FileStore(), db.DataStore()
	ctx := context.Background()

	ns, err := dStore.Namespaces().GetByName(ctx, core.MagicalGatewayNamespace)
	if err != nil {
		logger.Error("fetching namespace", "error", err)

		return
	}

	files, err := fStore.ForNamespace(ns.Name).ListDirektivFiles(ctx)
	if err != nil {
		logger.Error("listing direktiv files", "error", err)
	}
	endpoints := make([]*core.Endpoint, 0)
	for _, file := range files {
		if file.Typ != filestore.FileTypeEndpoint {
			continue
		}
		data, err := fStore.ForFile(file).GetData(ctx)
		if err != nil {
			logger.Error("read file data", "error", err)

			continue
		}
		item, err := spec.ParseEndpointFile(data)
		if err != nil {
			logger.Error("parse endpoint file", "error", err)

			continue
		}
		endpoints = append(endpoints, &core.Endpoint{
			Method: item.Method,
		})
	}

	gwManager.SetEndpoints(endpoints)
}

func renderServiceManager(db *database.DB, serviceManager core.ServiceManager, logger *zap.SugaredLogger) {
	logger = logger.With("subscriber", "services file watcher")

	fStore, dStore := db.FileStore(), db.DataStore()

	nsList, err := dStore.Namespaces().GetAll(context.Background())
	if err != nil {
		logger.Error("listing namespaces", "error", err)

		return
	}

	funConfigList := []*core.ServiceConfig{}

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
				serviceDef, err := spec.ParseServiceFile(data)
				if err != nil {
					logger.Error("parse service file", "error", err)

					continue
				}
				funConfigList = append(funConfigList, &core.ServiceConfig{
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
					funConfigList = append(funConfigList, &core.ServiceConfig{
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
