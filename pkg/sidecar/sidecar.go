package sidecar

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/telemetry"
)

const (
	direktivFlowEndpoint  = "DIREKTIV_FLOW_ENDPOINT"
	direktivOpentelemetry = "DIREKTIV_OTEL_BACKEND"
)

func RunApplication(ctx context.Context) {
	sl := new(SignalListener)
	sl.Start()

	slog.Info("listener started")

	otelProvider, err := telemetry.InitOpenTelemetry(ctx, os.Getenv(direktivOpentelemetry))
	if err != nil {
		slog.Error("opentelemetry setup failed", slog.Any("error", err))
	}

	slog.Info("opentelemtry", slog.String("server", os.Getenv(direktivOpentelemetry)))

	local := new(LocalServer)
	local.Start()
	slog.Info("local started")

	network := new(NetworkServer)
	network.local = local
	network.Start()
	slog.Info("network started")

	threads.Wait()

	if otelProvider != nil {
		err = otelProvider.Shutdown(ctx)
	}

	if err != nil {
		slog.Error("shutting down opentelemetry failed", slog.Any("error", err))
	}

	if code := threads.ExitStatus(); code != 0 {
		slog.Error("exiting with exit", "status_code", code)
		os.Exit(code)
	}
}

const (
	SUCCESS = 0
	ERROR   = 1
)

func Shutdown(code int) {
	t := time.Now().UTC()
	threads.Stop(&t, code)
}

func ForceQuit() {
	slog.Warn("performing force-quit")
	os.Exit(1)
}
