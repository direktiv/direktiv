package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/direktiv/direktiv/pkg/api"
	"github.com/direktiv/direktiv/pkg/cache"
	"github.com/direktiv/direktiv/pkg/certificates"
	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/engine"
	engineStore "github.com/direktiv/direktiv/pkg/engine/store"
	"github.com/direktiv/direktiv/pkg/extensions"
	"github.com/direktiv/direktiv/pkg/gateway"
	"github.com/direktiv/direktiv/pkg/natsclient"
	"github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/direktiv/direktiv/pkg/secrets"
	"github.com/direktiv/direktiv/pkg/service"
	"github.com/direktiv/direktiv/pkg/service/registry"
	"github.com/direktiv/direktiv/pkg/telemetry"
	_ "github.com/lib/pq" //nolint:revive
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//nolint:gocognit
func Start(circuit *core.Circuit) error {
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

	// create certs for communication
	slog.Info("initializing certificate updater")
	cm, err := certificates.NewCertificateUpdater(config.DirektivNamespace)
	if err != nil {
		return fmt.Errorf("initialize certificate updater, err: %w", err)
	}
	cm.Start(circuit)

	// wait for nats to be up and running and certs are done
	checkNATSConnectivity()

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
	slog.Info("initializing pubsub")
	nc, err := natsclient.Connect()
	if err != nil {
		return fmt.Errorf("can not connect to nats")
	}

	pubSub := pubsub.NewPubSub(nc)
	circuit.Go(func() error {
		err := pubSub.Loop(circuit)
		if err != nil {
			return fmt.Errorf("pubsub bus loop, err: %w", err)
		}

		return nil
	})
	app.PubSub = pubSub

	// creates bus with pub sub
	cache, err := cache.NewCache(pubSub, os.Getenv("POD_NAME"), false)
	circuit.Go(func() error {
		cache.Run(circuit)
		if err != nil {
			return fmt.Errorf("pubsub bus loop, err: %w", err)
		}

		return nil
	})
	app.Cache = cache

	slog.Info("initializing secrets handler")
	app.SecretsManager = secrets.NewManager(db, cache)

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

	circuit.Go(func() error {
		err := app.ServiceManager.Run(circuit)
		if err != nil {
			return fmt.Errorf("service manager, err: %w", err)
		}

		return nil
	})

	// Create js engine
	nc, err = natsclient.Connect()
	if err != nil {
		return fmt.Errorf("can not connect to nats")
	}
	eStore, err := engineStore.NewStore(circuit.Context(), nc)
	if err != nil {
		return fmt.Errorf("initializing engine, err: %w", err)
	}
	app.Engine, err = engine.NewEngine(db, eStore)
	if err != nil {
		return fmt.Errorf("initializing engine, err: %w", err)
	}
	circuit.Go(func() error {
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
	app.GatewayManager = gateway.NewManager(app)

	// Create syncNamespace function
	slog.Info("initializing sync namespace routine")
	// TODO: fix app.SyncNamespace init.

	pubSub.Subscribe(core.FileSystemChangeEvent, func(_ []byte) {
		renderServiceFiles(db, app.ServiceManager)
	})
	pubSub.Subscribe(core.NamespacesChangeEvent, func(_ []byte) {
		renderServiceFiles(db, app.ServiceManager)
	})
	// Call at least once before booting
	renderServiceFiles(db, app.ServiceManager)

	// endpoint manager
	pubSub.Subscribe(core.FileSystemChangeEvent, func(_ []byte) {
		renderGatewayFiles(db, app.GatewayManager)
	})
	pubSub.Subscribe(core.NamespacesChangeEvent, func(_ []byte) {
		renderGatewayFiles(db, app.GatewayManager)
	})
	// initial loading of routes and consumers
	renderGatewayFiles(db, app.GatewayManager)

	// initialize extensions
	if extensions.Initialize != nil {
		slog.Info("initializing extensions")
		if err = extensions.Initialize(db, pubSub, config); err != nil {
			return fmt.Errorf("initializing extensions, err: %w", err)
		}
	}

	// Start api server
	slog.Info("initializing api server")
	srv, err := api.Initialize(circuit, app, db)
	if err != nil {
		return fmt.Errorf("initializing api server, err: %w", err)
	}

	circuit.Go(func() error {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("shutdown api server, err: %w", err)
		}

		return nil
	})

	circuit.Go(func() error {
		<-circuit.Done()

		slog.Info("shutdown api server...")
		shutdownCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			slog.Error("shutdown api server", "err", err)
		}
		slog.Info("shutdown api server successful")

		return nil
	})

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

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  config.DB,
		PreferSimpleProtocol: false, // disables implicit prepared statement usage
		// Conn:                 edb.DB(),
	}), gormConf)
	if err != nil {
		return nil, err
	}

	res := db.Exec(database.Schema)
	if res.Error != nil {
		return nil, fmt.Errorf("provisioning schema, err: %w", res.Error)
	}
	slog.Info("schema provisioned successfully")

	if extensions.AdditionalSchema != "" {
		res = db.Exec(extensions.AdditionalSchema)
		if res.Error != nil {
			return nil, fmt.Errorf("provisioning additional schema, err: %w", res.Error)
		}
		slog.Info("additional schema provisioned successfully")
	}

	gdb, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("modifying gorm driver, err: %w", err)
	}

	slog.Debug("database connection pool limits set", "maxIdleConns", 32, "maxOpenConns", 16)
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

func checkNATSConnectivity() {
	// waiting for nats to be available
	// this waits for certificates as well
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			slog.Info("checking nats connection")
			nc, err := natsclient.Connect()
			if err == nil {
				nc.Close()
				slog.Info("nats available")

				return
			}
			slog.Error("nats connection not available", slog.Any("error", err))
		case <-time.After(2 * time.Minute):
			// can not recover from nats not connecting
			panic("cannot connect to nats")
		}
	}
}
