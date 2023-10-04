package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/api"
	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/function"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
	"go.uber.org/zap"
)

func NewMain(db *database.DB, pbus pubsub.Bus, logger *zap.SugaredLogger) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	done := make(chan struct{})

	// Create functions manager
	funcManager, err := function.NewManagerFromK8s()
	if err != nil {
		log.Fatalf("error creating functions client: %v\n", err)
	}
	// Start functions manager
	wg.Add(1)
	funcManager.Start(done, wg)

	// Create App
	app := &core.App{
		Version: &core.Version{
			UnixTime: time.Now().Unix(),
		},
		FunctionsManager: funcManager,
	}

	pbus.Subscribe(func(_ string) {
		renderServicesInFunctionsManager(db, funcManager, logger)
	},
		pubsub.WorkflowCreate,
		pubsub.WorkflowUpdate,
		pubsub.WorkflowDelete,
		pubsub.FunctionCreate,
		pubsub.FunctionUpdate,
		pubsub.FunctionDelete,
		pubsub.MirrorSync,
	)
	// Call at least once before booting
	renderServicesInFunctionsManager(db, funcManager, logger)

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

func renderServicesInFunctionsManager(db *database.DB, funcManager *function.Manager, logger *zap.SugaredLogger) {
	logger = logger.With("subscriber", "services file watcher")

	fStore, dStore := db.FileStore(), db.DataStore()

	nsList, err := dStore.Namespaces().GetAll(context.Background())
	if err != nil {
		logger.Error("listing namespaces", "error", err)

		return
	}

	funConfigList := []*function.Config{}

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
				serviceDef, err := spec.ParseService(data)
				if err != nil {
					logger.Error("parse service file", "error", err)

					continue
				}
				funConfigList = append(funConfigList, &function.Config{
					Namespace:   ns.Name,
					ServicePath: file.Path,
					Image:       serviceDef.Image,
					CMD:         serviceDef.Cmd,
					Size:        serviceDef.Size,
					Scale:       serviceDef.Scale,
				})
			} else if file.Typ == filestore.FileTypeWorkflow {
				serviceDef, err := spec.ParseWorkflowFunctionDefinition(data)
				if err != nil {
					logger.Error("parse workflow service def", "error", err)

					continue
				}
				if serviceDef.Typ == "knative-workflow" {
					funConfigList = append(funConfigList, &function.Config{
						Namespace:    ns.Name,
						WorkflowPath: file.Path,
						Image:        serviceDef.Image,
						CMD:          serviceDef.Cmd,
						Size:         serviceDef.Size,
						Scale:        serviceDef.Scale,
					})
				}
			}
		}
	}

	funcManager.SetServices(funConfigList)
}
