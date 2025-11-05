package cmdserver

import (
	"context"
	"log/slog"
	"os"

	"github.com/direktiv/direktiv/internal/cmdserver/pkg/server"
	"github.com/direktiv/direktiv/internal/core"
	intServer "github.com/direktiv/direktiv/internal/server"
	"github.com/direktiv/direktiv/internal/telemetry"
)

func Start() {
	intServer.InitSLog(&core.Config{
		LogDebug: false,
	})

	err := telemetry.InitOpenTelemetry(context.Background(), os.Getenv("DIREKTIV_OTEL_BACKEND"))
	if err != nil {
		slog.Warn("cannot init opentelemtry in sidecar", slog.Any("error", err))
	}
	slog.Info("opentelemetry configured", slog.String("addr", os.Getenv("DIREKTIV_OTEL_BACKEND")))

	slog.Info("starting cmd-exec server")

	// Create a new server with the RunCommands handler
	s := server.NewServer()

	slog.Debug("initialized cmd-exec server")

	// Start the server
	s.Start()

	slog.Info("cmd-exec server is running")
}
