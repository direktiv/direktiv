package config

import (
	root "github.com/direktiv/direktiv/cmd/exec/cmd"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "config-related commands.",
}

func init() {
	root.RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(confPInitCmd)
	configCmd.AddCommand(confInitCmd)
	configCmd.AddCommand(confAddCmd)
}
