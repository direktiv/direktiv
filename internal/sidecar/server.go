package sidecar

import (
	"context"
	"log"
	"log/slog"
	"net/http"
)

type sidecar struct {
	external *externalServer
	internal *internalServer
}

func newSidecar() *sidecar {
	return &sidecar{
		internal: newInternalServer(),
		external: newExternalServer(),
	}
}

func (sc *sidecar) start() {
	slog.Info("starting sidecar")
	go func() {
		if err := sc.internal.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error starting internal server: %v\n", err)
		}
	}()

	go func() {
		if err := sc.external.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error starting external server: %v\n", err)
		}
	}()
}

func (sc *sidecar) stop(ctx context.Context) error {
	err := sc.internal.server.Shutdown(ctx)
	if err != nil {
		return err
	}

	return sc.external.server.Shutdown(ctx)
}
