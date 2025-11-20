package sidecar

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/server"
	"github.com/direktiv/direktiv/internal/telemetry"
)

func RunApplication(ctx context.Context) {
	server.InitSLog(&core.Config{
		LogDebug: false,
	})

	err := waitForUserContainer()
	if err != nil {
	}

	err = telemetry.InitOpenTelemetry(ctx, os.Getenv("DIREKTIV_OTEL_BACKEND"))
	if err != nil {
		slog.Warn("cannot init opentelemtry in sidecar", slog.Any("error", err))
	}
	slog.Info("opentelemetry configured", slog.String("addr", os.Getenv("DIREKTIV_OTEL_BACKEND")))

	sidecar := newSidecar()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	sidecar.start()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	slog.Info("shutting down server")

	err = sidecar.stop(ctx)
	if err != nil {
		slog.Error("shutting down server failed", slog.Any("error", err))
		return
	}

	slog.Info("server stopped")
}

func waitForUserContainer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for user container")
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", "localhost:8080", 1*time.Second)
			if err == nil {
				conn.Close()
				return nil
			}
		}
	}
}

// legacy logging, can be removed later.
type requestMap struct {
	mu      sync.Mutex
	syncMap sync.Map
}

func (rm *requestMap) Add(id string, log telemetry.LogObject) {
	rm.syncMap.Store(id, log)
}

func (rm *requestMap) Remove(id string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
}

func (rm *requestMap) Get(id string) telemetry.LogObject {
	lo, ok := rm.syncMap.Load(id)
	if !ok {
		return telemetry.LogObject{}
	}

	return lo.(telemetry.LogObject)
}
