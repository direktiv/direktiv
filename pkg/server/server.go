package server

import (
	"context"
	"database/sql"
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
	"github.com/direktiv/direktiv/pkg/engine"
	store2 "github.com/direktiv/direktiv/pkg/engine/store"
	"github.com/direktiv/direktiv/pkg/extensions"
	"github.com/direktiv/direktiv/pkg/gateway"
	"github.com/direktiv/direktiv/pkg/pubsub"
	pubsubSQL "github.com/direktiv/direktiv/pkg/pubsub/sql"
	"github.com/direktiv/direktiv/pkg/service"
	"github.com/direktiv/direktiv/pkg/service/registry"
	"github.com/direktiv/direktiv/pkg/telemetry"
	"github.com/direktiv/direktiv/pkg/utils"
	"github.com/nats-io/nats.go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//nolint:gocognit
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
		err = rawDB.PingContext(context.Background())
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

	// Create service manager
	slog.Info("initializing service manager")
	app.ServiceManager, err = service.NewManager(config, func() ([]string, error) {
		beats, err := db.DataStore().HeartBeats().Since(context.Background(), "life_services", 100)
		if err != nil {
			return nil, err
		}
		list := make([]string, len(beats))
		for i := range beats {
			list[i] = beats[i].Key
		}

		return list, nil
	})
	if err != nil {
		return fmt.Errorf("initializing service manager, err: %w", err)
	}

	circuit.Start(func() error {
		err := app.ServiceManager.Run(circuit)
		if err != nil {
			return fmt.Errorf("service manager, err: %w", err)
		}

		return nil
	})

	// Create js engine
	store, err := store2.NewStore(circuit.Context(), "nats://nats:4222", "foooo")
	if err != nil {
		return fmt.Errorf("initializing engine, err: %w", err)
	}
	app.Engine, err = engine.NewEngine(db, store)
	if err != nil {
		return fmt.Errorf("initializing engine, err: %w", err)
	}
	circuit.Start(func() error {
		err := app.Engine.Start(circuit)
		if err != nil {
			return fmt.Errorf("engine, err: %w", err)
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
	// TODO: fix app.SyncNamespace init.

	bus.Subscribe(&pubsub.FileSystemChangeEvent{}, func(_ string) {
		renderServiceFiles(db, app.ServiceManager)
	})
	bus.Subscribe(&pubsub.NamespacesChangeEvent{}, func(_ string) {
		renderServiceFiles(db, app.ServiceManager)
	})
	// Call at least once before booting
	renderServiceFiles(db, app.ServiceManager)

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
		slog.Info("initializing extensions")
		if err = extensions.Initialize(db, bus, config); err != nil {
			return fmt.Errorf("initializing extensions, err: %w", err)
		}
	}

	// Start api v2 server
	err = api.Initialize(circuit, app, db, bus)
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

	var db *gorm.DB
	var err error

	utils.Retry(time.Second, 10, func() error {
		slog.Info("connecting to database...")
		db, err = gorm.Open(postgres.New(postgres.Config{
			DSN:                  config.DB,
			PreferSimpleProtocol: false, // disables implicit prepared statement usage
			// Conn:                 edb.DB(),
		}), gormConf)

		return err
	})
	if err != nil {
		return nil, err
	}
	slog.Info("successfully connected to the database")

	var nc *nats.Conn
	utils.Retry(time.Second, 10, func() error {
		slog.Info("test connection to nats...")
		nc, err = nats.Connect("nats://nats:4222")

		return err
	})
	if err != nil {
		return nil, err
	}
	nc.Drain()
	slog.Info("successfully connected to the nats")

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

	ctxHandler := telemetry.NewContextHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				// force time format to full length with trainling zeros
				a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02T15:04:05.000000000Z"))
			}

			return a
		},
	}))

	slogger := slog.New(ctxHandler)

	slog.SetDefault(slogger)
}
