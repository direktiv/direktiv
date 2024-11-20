package main

import (
	"log/slog"

	"github.com/direktiv/direktiv/cmd/cmd-exec/pkg/commands"
	"github.com/direktiv/direktiv/cmd/cmd-exec/pkg/server"
)

func main() {
	slog.Info("starting cmd-exec server")

	// Create a new server with the RunCommands handler
	s := server.NewServer[commands.Commands](commands.RunCommands)

	slog.Debug("initialized cmd-exec server")

	// Start the server
	s.Start()

	slog.Info("cmd-exec server is running")
}
