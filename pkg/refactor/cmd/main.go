package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/direktiv/direktiv/pkg/refactor/api"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/function"
	"github.com/direktiv/direktiv/pkg/refactor/pubsub"
	"github.com/direktiv/direktiv/pkg/refactor/webapi"
	"go.uber.org/zap"
)

func NewMain(db *database.DB, pbus pubsub.Bus, logger *zap.SugaredLogger) *sync.WaitGroup {
	funcManager, err := function.NewManagerFromK8s()
	if err != nil {
		log.Fatalf("error creating functions client: %v\n", err)
	}

	wg := &sync.WaitGroup{}
	done := make(chan struct{})

	pbus.Subscribe(func(_ string) {
		subscriberServicesChanges(db, funcManager, logger)
	},
		pubsub.WorkflowCreate,
		pubsub.WorkflowUpdate,
		pubsub.WorkflowDelete,
		pubsub.FunctionCreate,
		pubsub.FunctionUpdate,
		pubsub.FunctionDelete,
		pubsub.MirrorSync,
	)
	subscriberServicesChanges(db, funcManager, logger)

	go func() {
		// Listen for syscall signals for process to interrupt/quit
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		close(done)
	}()

	// Start functions manager
	wg.Add(1)
	funcManager.Start(done, wg)

	// Start api v2 server
	wg.Add(1)
	webapi.Start(funcManager, "0.0.0.0:6667", done, wg)

	return wg
}

func subscriberServicesChanges(db *database.DB, funcManager *function.Manager, logger *zap.SugaredLogger) {
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
				serviceDef := api.ParseService(data)
				if serviceDef == nil {
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
				serviceDef, err := api.ParseWorkflowFunctionDefinition(data)
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
