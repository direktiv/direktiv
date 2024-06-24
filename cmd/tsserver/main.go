package tsserver

import (
	"log/slog"

	"github.com/direktiv/direktiv/pkg/tsengine"
)

func RunApplication() {
	srv, err := tsengine.NewServer()
	if err != nil {
		slog.Error("failed to start the ts-engine server", "error", err)
		panic(err)
	}

	panic(srv.Start())
}
