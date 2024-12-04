package cmdserver

import (
	"log/slog"

	"github.com/direktiv/direktiv/pkg/cmdserver/pkg/commands"
	"github.com/direktiv/direktiv/pkg/cmdserver/pkg/server"
)

func Start() {
	slog.Info("starting cmd-exec server")

	// Create a new server with the RunCommands handler
	s := server.NewServer[commands.Commands](commands.RunCommands)

	slog.Debug("initialized cmd-exec server")

	// Start the server
	s.Start()

	slog.Info("cmd-exec server is running")
}
