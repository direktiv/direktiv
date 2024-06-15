package main

import (
	"fmt"

	"github.com/direktiv/direktiv/cmd/exec/cli"
)

// _ "github.com/direktiv/direktiv/cmd/exec/cmd/config"
// _ "github.com/direktiv/direktiv/cmd/exec/cmd/events"
// _ "github.com/direktiv/direktiv/cmd/exec/cmd/logs"
// _ "github.com/direktiv/direktiv/cmd/exec/cmd/workflows"

func main() {
	err := cli.RootCmd.Execute()
	if err != nil {
		fmt.Printf("command failed: %s \n", err.Error())
	}
}

// var RootCmd = &cobra.Command{
// 	Use: "direktivctl",
// }
