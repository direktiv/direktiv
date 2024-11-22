package sidecar

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/tracing"
)

const (
	direktivFlowEndpoint  = "DIREKTIV_FLOW_ENDPOINT"
	direktivOpentelemetry = "DIREKTIV_OTLP"
)

func RunApplication(ctx context.Context) {
	sl := new(SignalListener)
	sl.Start()

	fmt.Printf("listener started\n")

	openTelemetryBackend := os.Getenv(direktivOpentelemetry)

	telend, err := tracing.InitTelemetry(ctx, openTelemetryBackend, "direktiv/sidecar", "direktiv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize telemetry: %v\n", err)
		os.Exit(1)
	}
	defer telend()

	local := new(LocalServer)
	local.Start()
	fmt.Printf("local started\n")

	network := new(NetworkServer)
	network.local = local
	network.Start()
	fmt.Printf("network started\n")

	threads.Wait()

	if code := threads.ExitStatus(); code != 0 {
		slog.Error("Exiting with exit.", "status_code", code)
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
	slog.Warn("Performing force-quit.")
	os.Exit(1)
}
