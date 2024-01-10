package main

import (
	"github.com/direktiv/direktiv/cmd/cmd-exec/pkg/commands"
	"github.com/direktiv/direktiv/cmd/cmd-exec/pkg/server"
)

func main() {
	s := server.NewServer[commands.Commands](commands.RunCommands)
	s.Start()
}
