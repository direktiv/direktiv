package tsengine

import (
	"log"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/direktiv/direktiv/pkg/tsengine"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func RunApplication() {

	// parsing config
	cfg := tsengine.Config{}
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}
	setLogLevel(cfg.LogLevel)

	// loggingCtx = tracing.WithTrack(context.Background(), tracing.BuildNamespaceTrack(args.Namespace.Name))
	// slog.Error("Failed to parse workflow definition.", tracing.GetSlogAttributesWithError(loggingCtx, err)...)

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
		DSN:                  cfg.DBConfig,
		PreferSimpleProtocol: false,
	}), gormConf)
	if err != nil {
		panic(err)
	}

	srv, err := tsengine.NewServer(cfg, db)
	if err != nil {
		panic(err)
	}

	panic(srv.Start())
}

func setLogLevel(level string) {

	ll := slog.LevelDebug
	switch level {
	case "info":
		ll = slog.LevelInfo
	case "warn":
		ll = slog.LevelWarn
	case "error":
		ll = slog.LevelError
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: ll})
	logger := slog.New(handler)
	slog.SetDefault(logger)

}
