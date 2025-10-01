package cmdserver

import (
	"log/slog"

	"github.com/direktiv/direktiv/internal/cmdserver/pkg/server"
)

func Start() {
	slog.Info("starting cmd-exec server")

	// Create a new server with the RunCommands handler
	s := server.NewServer(server.RunCommands)

	slog.Debug("initialized cmd-exec server")

	// Start the server
	s.Start()

	slog.Info("cmd-exec server is running")
}
