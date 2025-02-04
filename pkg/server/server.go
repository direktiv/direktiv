package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/direktiv/direktiv/pkg/api"
	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/extensions"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/flow"
	"github.com/direktiv/direktiv/pkg/gateway"
	"github.com/direktiv/direktiv/pkg/instancestore"
	"github.com/direktiv/direktiv/pkg/mirror"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/pubsub"
	pubsubSQL "github.com/direktiv/direktiv/pkg/pubsub/sql"
	"github.com/direktiv/direktiv/pkg/service"
	"github.com/direktiv/direktiv/pkg/service/registry"
	"github.com/direktiv/direktiv/pkg/tracing"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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

func Run(circuit *core.Circuit) error {
	var err error
	config := &core.Config{}
	if err := env.Parse(config); err != nil {
		return fmt.Errorf("parsing env variables: %w", err)
	}
	if err := config.Init(); err != nil {
		return fmt.Errorf("init config, err: %w", err)
	}
	initSLog(config)

	// Create App struct
	app := core.App{
		Version: &core.Version{
			UnixTime: time.Now().Unix(),
		},
		Config: config,
	}

	// Create DB connection
	slog.Info("initializing db connection")
	db, err := initDB(config)
	if err != nil {
		return fmt.Errorf("initialize db, err: %w", err)
	}
	datastore.SymmetricEncryptionKey = config.SecretKey

	// Create Raw DB connection
	slog.Info("initializing raw db connection")
	rawDB, err := sql.Open("postgres", config.DB)
	if err == nil {
		err = rawDB.Ping()
	}
	if err != nil {
		return fmt.Errorf("creating raw db driver, err: %w", err)
	}

	// Create Bus
	slog.Info("initializing pubsub2")
	coreBus, err := pubsubSQL.NewPostgresCoreBus(rawDB, app.Config.DB)
	if err != nil {
		return fmt.Errorf("creating pubsub core bus, err: %w", err)
	}
	bus := pubsub.NewBus(coreBus)
	circuit.Start(func() error {
		err := bus.Loop(circuit)
		if err != nil {
			return fmt.Errorf("pubsub bus loop, err: %w", err)
		}

		return nil
	})

	// Initialize legacy server
	slog.Info("initializing legacy server")
	srv, err := flow.InitLegacyServer(circuit, config, bus, db, rawDB)
	if err != nil {
		return fmt.Errorf("initialize legacy server, err: %w", err)
	}

	instanceManager := &instancestore.InstanceManager{
		Start:  srv.Engine.StartWorkflow,
		Cancel: srv.Engine.CancelInstance,
	}

	// Create service manager
	slog.Info("initializing service manager")
	app.ServiceManager, err = service.NewManager(config)
	if err != nil {
		return fmt.Errorf("initializing service manager, err: %w", err)
	}

	// Setup GetServiceURL function
	service.SetupGetServiceURLFunc(config)

	circuit.Start(func() error {
		err := app.ServiceManager.Run(circuit)
		if err != nil {
			return fmt.Errorf("service manager, err: %w", err)
		}

		return nil
	})

	// Create registry manager
	slog.Info("initializing registry manager")
	app.RegistryManager, err = registry.NewManager()
	if err != nil {
		slog.Error("registry manager", "error", err)
		panic(err)
	}

	// Create endpoint manager
	slog.Info("initializing gateway manager")
	app.GatewayManager = gateway.NewManager(db)

	// Create syncNamespace function
	slog.Info("initializing sync namespace routine")
	app.SyncNamespace = func(namespace any, mirrorConfig any) (any, error) {
		ns := namespace.(*datastore.Namespace)            //nolint:forcetypeassert
		mConfig := mirrorConfig.(*datastore.MirrorConfig) //nolint:forcetypeassert
		proc, err := srv.MirrorManager.NewProcess(context.Background(), ns, datastore.ProcessTypeSync)
		if err != nil {
			return nil, err
		}

		go func() {
			srv.MirrorManager.Execute(context.Background(), proc, mConfig, &mirror.DirektivApplyer{NamespaceID: ns.ID})
			err := srv.Bus.Publish(&pubsub.NamespacesChangeEvent{
				Action: "sync",
				Name:   ns.Name,
			})
			if err != nil {
				slog.Error("pubsub publish", "error", err)
			}
		}()

		return proc, nil
	}

	srv.Bus.Subscribe(&pubsub.FileSystemChangeEvent{}, func(_ string) {
		renderServiceFiles(db, app.ServiceManager)
	})
	srv.Bus.Subscribe(&pubsub.NamespacesChangeEvent{}, func(_ string) {
		renderServiceFiles(db, app.ServiceManager)
	})
	// Call at least once before booting
	renderServiceFiles(db, app.ServiceManager)

	srv.Bus.Subscribe(&pubsub.FileSystemChangeEvent{}, func(data string) {
		event := &pubsub.FileSystemChangeEvent{}
		err := json.Unmarshal([]byte(data), event)
		if err != nil {
			panic("Logic Error could not parse file system change event")
		}

		err = srv.ConfigureWorkflow(event)
		if err != nil {
			slog.Error("configure workflow", "error", err)
		}
	})

	slog.Debug("Rendering event-listeners on server start")
	err = flow.RenderAllStartEventListeners(circuit.Context(), db)
	if err != nil {
		slog.Error("rendering event listener on server start", "error", err)
	}
	slog.Debug("Completed rendering event-listeners on server start")

	// endpoint manager
	bus.Subscribe(&pubsub.FileSystemChangeEvent{}, func(_ string) {
		renderGatewayFiles(db, app.GatewayManager)
	})
	bus.Subscribe(&pubsub.NamespacesChangeEvent{}, func(_ string) {
		renderGatewayFiles(db, app.GatewayManager)
	})
	// initial loading of routes and consumers
	renderGatewayFiles(db, app.GatewayManager)

	// initialize extensions
	if extensions.Initialize != nil {
		extensions.Initialize(db, bus, config)
	}

	// Start api v2 server
	err = api.Initialize(circuit, app, db, bus, instanceManager, srv.Engine.WakeEventsWaiter, srv.Engine.EventsInvoke)
	if err != nil {
		return fmt.Errorf("initializing api v2, err: %w", err)
	}
	slog.Info("api server v2 started")

	return nil
}

func initDB(config *core.Config) (*database.DB, error) {
	gormConf := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel:                  logger.Silent,
				IgnoreRecordNotFoundError: true,
			},
		),
	}

	var err error
	var db *gorm.DB
	//nolint:intrange
	for i := 0; i < 10; i++ {
		slog.Info("connecting to database...")

		db, err = gorm.Open(postgres.New(postgres.Config{
			DSN:                  config.DB,
			PreferSimpleProtocol: false, // disables implicit prepared statement usage
			// Conn:                 edb.DB(),
		}), gormConf)
		if err == nil {
			slog.Info("successfully connected to the database.")

			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		return nil, err
	}

	res := db.Exec(database.Schema)
	if res.Error != nil {
		return nil, fmt.Errorf("provisioning schema, err: %w", res.Error)
	}
	slog.Info("Schema provisioned successfully")

	if extensions.AdditionalSchema != "" {
		res = db.Exec(extensions.AdditionalSchema)
		if res.Error != nil {
			return nil, fmt.Errorf("provisioning additional schema, err: %w", res.Error)
		}
		slog.Info("Additional schema provisioned successfully")
	}

	gdb, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("modifying gorm driver, err: %w", err)
	}

	slog.Debug("Database connection pool limits set", "maxIdleConns", 32, "maxOpenConns", 16)
	gdb.SetMaxIdleConns(32)
	gdb.SetMaxOpenConns(16)

	return database.NewDB(db), nil
}

func initSLog(cfg *core.Config) {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelInfo)

	logDebug := cfg.LogDebug
	if logDebug {
		slog.Info("logging is set to debug")
		lvl.Set(slog.LevelDebug)
	}
	handlers := tracing.NewContextHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	}))
	slogger := slog.New(
		tracing.TeeHandler{
			handlers,
			tracing.EventHandler{},
		})

	slog.SetDefault(slogger)
}
