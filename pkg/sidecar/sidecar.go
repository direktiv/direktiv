package sidecar

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

const (
	direktivFlowEndpoint  = "DIREKTIV_FLOW_ENDPOINT"
	direktivOpentelemetry = "DIREKTIV_OTLP"
)

func RunApplication(ctx context.Context) {
	sl := new(SignalListener)
	sl.Start()

	fmt.Printf("listener started\n")

	// openTelemetryBackend := os.Getenv(direktivOpentelemetry)

	// telend, err := tracing.InitTelemetry(ctx, openTelemetryBackend, "direktiv/sidecar", "direktiv")
	// if err != nil {
	// 	slog.Warn("failed to initialize telemetry, but continuing", "error", err)
	// } else {
	// 	defer telend()
	// }

	local := new(LocalServer)
	local.Start()
	fmt.Printf("local started\n")

	network := new(NetworkServer)
	network.local = local
	network.Start()
	fmt.Printf("network started\n")

	threads.Wait()

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
