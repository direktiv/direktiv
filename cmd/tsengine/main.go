package tsengine

import (
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/direktiv/direktiv/pkg/tsengine"
)

func RunApplication() {

	// parsing config
	cfg := tsengine.Config{}
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}
	setLogLevel(cfg.LogLevel)

	srv, err := tsengine.NewServer(cfg, nil)
	if err != nil {
		panic(err)
	}

	srv.Start()
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
