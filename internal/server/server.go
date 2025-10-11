package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/direktiv/direktiv/internal/api"
	"github.com/direktiv/direktiv/internal/cluster/cache/memcache"
	"github.com/direktiv/direktiv/internal/cluster/certs"
	"github.com/direktiv/direktiv/internal/cluster/pubsub"
	natspubsub "github.com/direktiv/direktiv/internal/cluster/pubsub/nats"
	"github.com/direktiv/direktiv/internal/compiler"
	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/datastore"
	"github.com/direktiv/direktiv/internal/datastore/datasql"
	"github.com/direktiv/direktiv/internal/engine"
	"github.com/direktiv/direktiv/internal/engine/databus"
	"github.com/direktiv/direktiv/internal/extensions"
	"github.com/direktiv/direktiv/internal/gateway"
	"github.com/direktiv/direktiv/internal/mirroring"
	intNats "github.com/direktiv/direktiv/internal/nats"
	"github.com/direktiv/direktiv/internal/sched"
	"github.com/direktiv/direktiv/internal/secrets"
	"github.com/direktiv/direktiv/internal/service"
	"github.com/direktiv/direktiv/internal/service/registry"
	"github.com/direktiv/direktiv/internal/telemetry"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	_ "github.com/lib/pq" //nolint:revive
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"k8s.io/utils/clock"
)

//nolint:gocognit, maintidx
func Start(lc *lifecycle.Manager) error {
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
	app := api.InitializeArgs{
		Version: &api.Version{
			UnixTime: time.Now().Unix(),
		},
		Config: config,
	}

	// initializing certificate-updater
	{
		slog.Info("initializing certificate-updater")
		cm, err := certs.NewCertificateUpdater(config.DirektivNamespace)
		if err != nil {
			return fmt.Errorf("create certificate-updater, err: %w", err)
		}
		err = cm.Start(lc)
		if err != nil {
			return fmt.Errorf("start certificate-updater, err: %w", err)
		}
	}

	// wait for nats to be up and running and certs are done
	checkNATSConnectivity()

	// Create DB connection
	slog.Info("initializing db connection")
	app.DB, err = initDB(config)
	if err != nil {
		return fmt.Errorf("initialize db, err: %w", err)
	}
	datastore.SymmetricEncryptionKey = config.SecretKey

	// initializing pubsub
	{
		slog.Info("initializing pubsub")
		app.PubSub, err = natspubsub.New(intNats.Connect, slog.Default())
		if err != nil {
			return fmt.Errorf("initialize pubsub, err: %w", err)
		}
		lc.OnShutdown(func() error {
			err := app.PubSub.Close()
			if err != nil {
				return fmt.Errorf("closing pubsub, err: %w", err)
			}

			return nil
		})
	}

	// initializing memcache
	{
		slog.Info("initializing memcache")
		app.Cache, err = memcache.New(app.PubSub, os.Getenv("POD_NAME"), false, slog.Default())
		if err != nil {
			return fmt.Errorf("create memcache, err: %w", err)
		}
		lc.OnShutdown(func() error {
			app.Cache.Close()
			return nil
		})
	}

	// initializing secrets-handler
	{
		slog.Info("initializing secrets-handler")
		app.SecretsManager = secrets.NewManager(app.DB, app.Cache)
	}

	// initializing service-manager
	{
		slog.Info("initializing service-manager")
		fas := func() ([]string, error) {
			beats, err := datasql.NewStore(app.DB).HeartBeats().Since(context.Background(), "life_services", 100)
			if err != nil {
				return nil, err
			}
			list := make([]string, len(beats))
			for i := range beats {
				list[i] = beats[i].Key
			}

			return list, nil
		}

		app.ServiceManager, err = service.NewManager(config, fas)
		if err != nil {
			return fmt.Errorf("create service-manager, err: %w", err)
		}
		err = app.ServiceManager.Start(lc)
		if err != nil {
			return fmt.Errorf("start service-manager, err: %w", err)
		}

		app.PubSub.Subscribe(pubsub.SubjFileSystemChange, func(_ []byte) {
			renderServiceFiles(app.DB, app.ServiceManager)
		})
		app.PubSub.Subscribe(pubsub.SubjNamespacesChange, func(_ []byte) {
			renderServiceFiles(app.DB, app.ServiceManager)
		})
		// call at least once before booting
		renderServiceFiles(app.DB, app.ServiceManager)
	}

	// initializing engine
	{
		// prepare compiler
		comp, err := compiler.NewCompiler(app.DB, app.Cache)
		if err != nil {
			return fmt.Errorf("creating compiler, err: %w", err)
		}

		slog.Info("initializing engine-nats")
		nc, err := intNats.Connect()
		if err != nil {
			return fmt.Errorf("create engine-nats, err: %w", err)
		}
		js, err := intNats.SetupJetStream(context.Background(), nc)
		// TODO: remove this dev code.
		if err != nil {
			err = fmt.Errorf("reset streams, err: %w", err)
		}
		if err != nil {
			return fmt.Errorf("create engine-nats, err: %w", err)
		}
		lc.OnShutdown(func() error {
			return nc.Drain()
		})

		slog.Info("initializing engine")
		app.Engine, err = engine.NewEngine(
			databus.New(js),
			comp,
			js,
		)
		if err != nil {
			return fmt.Errorf("create engine, err: %w", err)
		}
		err = app.Engine.Start(lc)
		if err != nil {
			return fmt.Errorf("start engine, err: %w", err)
		}

		slog.Info("initializing scheduler")
		app.Scheduler = sched.New(js, clock.RealClock{}, slog.With("component", "scheduler"))
		err = app.Scheduler.Start(lc)
		if err != nil {
			return fmt.Errorf("start scheduler, err: %w", err)
		}
	}

	// initializing registry-manager
	{
		slog.Info("initializing registry-manager")
		app.RegistryManager, err = registry.NewManager()
		if err != nil {
			return fmt.Errorf("create registry-manager, err: %w", err)
		}
	}

	// initializing gateway-manager
	{
		slog.Info("initializing gateway manager")
		app.GatewayManager = gateway.NewManager(app.SecretsManager)

		app.PubSub.Subscribe(pubsub.SubjFileSystemChange, func(_ []byte) {
			renderGatewayFiles(app.DB, app.GatewayManager)
		})
		app.PubSub.Subscribe(pubsub.SubjNamespacesChange, func(_ []byte) {
			renderGatewayFiles(app.DB, app.GatewayManager)
		})
		// call at least once before booting
		renderGatewayFiles(app.DB, app.GatewayManager)
	}

	// initializing extensions
	{
		if extensions.Initialize != nil {
			slog.Info("initializing extensions")
			if err = extensions.Initialize(app.DB, app.PubSub, config); err != nil {
				return fmt.Errorf("initializing extensions, err: %w", err)
			}
		}
	}

	// TODO: backend jobs should be created by lc.Go()
	// start mirror process cleanup
	go mirroring.RunCleanMirrorProcesses(lc.Context(), app.DB)

	// initializing api-serer
	{
		slog.Info("initializing api server")
		srv, err := api.New(app)
		if err != nil {
			return fmt.Errorf("create api-server, err: %w", err)
		}
		err = srv.Start(lc)
		if err != nil {
			return fmt.Errorf("start api-server, err: %w", err)
		}
		lc.OnShutdown(func() error {
			err := srv.Close(context.Background())
			if err != nil {
				return fmt.Errorf("close api-server, err: %w", err)
			}

			return nil
		})
	}

	return nil
}

func initDB(config *core.Config) (*gorm.DB, error) {
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

	return db, nil
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
			nc, err := intNats.Connect()
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
