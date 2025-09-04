package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/direktiv/direktiv/internal/api"
	"github.com/direktiv/direktiv/internal/cache"
	"github.com/direktiv/direktiv/internal/cluster/certs"
	"github.com/direktiv/direktiv/internal/cluster/pubsub"
	natspubsub "github.com/direktiv/direktiv/internal/cluster/pubsub/nats"
	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/datastore"
	"github.com/direktiv/direktiv/internal/datastore/datasql"
	"github.com/direktiv/direktiv/internal/engine"
	engineStore "github.com/direktiv/direktiv/internal/engine/store"
	"github.com/direktiv/direktiv/internal/extensions"
	"github.com/direktiv/direktiv/internal/gateway"
	"github.com/direktiv/direktiv/internal/secrets"
	"github.com/direktiv/direktiv/internal/service"
	"github.com/direktiv/direktiv/internal/service/registry"
	"github.com/direktiv/direktiv/internal/telemetry"
	database2 "github.com/direktiv/direktiv/pkg/database"
	_ "github.com/lib/pq" //nolint:revive
	"github.com/nats-io/nats.go"
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
	app := api.InitializeArgs{
		Version: &api.Version{
			UnixTime: time.Now().Unix(),
		},
		Config: config,
	}

	// create certs for communication
	slog.Info("initializing certificate updater")
	cm, err := certs.NewCertificateUpdater(config.DirektivNamespace)
	if err != nil {
		return fmt.Errorf("initialize certificate updater, err: %w", err)
	}
	cm.Start(circuit)

	// wait for nats to be up and running and certs are done
	checkNATSConnectivity()

	// Create DB connection
	slog.Info("initializing db connection")
	app.DB, err = initDB(config)
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

	// Create EventBus
	slog.Info("initializing pubsub")
	nc, err := natsConnect()
	if err != nil {
		return fmt.Errorf("can not connect to nats")
	}

	pubSub := natspubsub.New(nc)
	circuit.Go(func() error {
		<-circuit.Done()
		err := nc.Drain()
		if err != nil {
			return fmt.Errorf("nats pubsub drain, err: %w", err)
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
	app.SecretsManager = secrets.NewManager(app.DB, cache)

	// Create service manager
	slog.Info("initializing service manager")
	app.ServiceManager, err = service.NewManager(config, func() ([]string, error) {
		beats, err := datasql.NewStore(app.DB).HeartBeats().Since(context.Background(), "life_services", 100)
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
	nc, err = natsConnect()
	if err != nil {
		return fmt.Errorf("can not connect to nats")
	}
	eStore, err := engineStore.NewStore(circuit.Context(), nc)
	if err != nil {
		return fmt.Errorf("initializing engine, err: %w", err)
	}
	app.Engine, err = engine.NewEngine(app.DB, eStore)
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
	app.GatewayManager = gateway.NewManager(app.SecretsManager)

	// Create syncNamespace function
	slog.Info("initializing sync namespace routine")
	// TODO: fix app.SyncNamespace init.

	pubSub.Subscribe(pubsub.SubjFileSystemChange, func(_ []byte) {
		renderServiceFiles(app.DB, app.ServiceManager)
	})
	pubSub.Subscribe(pubsub.SubjNamespacesChange, func(_ []byte) {
		renderServiceFiles(app.DB, app.ServiceManager)
	})
	// Call at least once before booting
	renderServiceFiles(app.DB, app.ServiceManager)

	// endpoint manager
	pubSub.Subscribe(pubsub.SubjFileSystemChange, func(_ []byte) {
		renderGatewayFiles(app.DB, app.GatewayManager)
	})
	pubSub.Subscribe(pubsub.SubjNamespacesChange, func(_ []byte) {
		renderGatewayFiles(app.DB, app.GatewayManager)
	})
	// initial loading of routes and consumers
	renderGatewayFiles(app.DB, app.GatewayManager)

	// initialize extensions
	if extensions.Initialize != nil {
		slog.Info("initializing extensions")
		if err = extensions.Initialize(app.DB, pubSub, config); err != nil {
			return fmt.Errorf("initializing extensions, err: %w", err)
		}
	}

	// Start api server
	slog.Info("initializing api server")
	srv, err := api.Initialize(circuit, app)
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

	res := db.Exec(database2.Schema)
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
			nc, err := natsConnect()
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

func natsConnect() (*nats.Conn, error) {
	// set the deployment name in dns names
	deploymentName := os.Getenv("DIREKTIV_DEPLOYMENT_NAME")

	return nats.Connect(
		fmt.Sprintf("tls://%s-nats.default.svc:4222", deploymentName),
		nats.ClientTLSConfig(
			func() (tls.Certificate, error) {
				cert, err := tls.LoadX509KeyPair("/etc/direktiv-tls/server.crt",
					"/etc/direktiv-tls/server.key")
				if err != nil {
					slog.Error("cannot create certificate pair", slog.Any("error", err))
					return tls.Certificate{}, err
				}

				return cert, nil
			},
			func() (*x509.CertPool, error) {
				caCert, err := os.ReadFile("/etc/direktiv-tls/ca.crt")
				if err != nil {
					return nil, err
				}
				caPool := x509.NewCertPool()
				if !caPool.AppendCertsFromPEM(caCert) {
					slog.Error("cannot create certificate pair", slog.Any("error", err))
					return nil, err
				}

				return caPool, nil
			},
		),
	)
}
