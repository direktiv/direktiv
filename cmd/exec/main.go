package main

import (
	"fmt"

	root "github.com/direktiv/direktiv/cmd/exec/cmd"
	_ "github.com/direktiv/direktiv/cmd/exec/cmd/config"
	_ "github.com/direktiv/direktiv/cmd/exec/cmd/events"
	_ "github.com/direktiv/direktiv/cmd/exec/cmd/logs"
	_ "github.com/direktiv/direktiv/cmd/exec/cmd/workflows"
)

func main() {
	err := root.RootCmd.Execute()
	if err != nil {
		fmt.Printf("command failed: %s \n", err.Error())
	}
}
