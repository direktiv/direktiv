package workflows

import (
	"path/filepath"

	root "github.com/direktiv/direktiv/cmd/exec/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var infoCmd = &cobra.Command{
	Use:              "info",
	Short:            "Prints detected configuration values for current project",
	PersistentPreRun: root.InitConfigurationAndProject,
	Run: func(cmd *cobra.Command, args []string) {
		pf := viper.GetString("projectFile")
		if pf == "" {
			root.Fail(cmd, "Could not get project directory from the pwd or configuration")
		}

		cmd.Printf("project file: %s\n", pf)
		dir := filepath.Dir(pf)

		cmd.Printf("project directory: %s\n", dir)

		pwd := viper.GetString("directory")
		if pwd == "" {
			root.Fail(cmd, "Could not get working directory")
		}
		cmd.Printf("working directory: %s\n", pwd)
		cmd.Printf("\n")
		cmd.Printf("namespace: %s\n", root.GetNamespace())
		cmd.Printf("URL: %s\n", root.UrlPrefix)

		auth := root.GetAuth()
		printAuth := "***"

		if len(auth) > 6 {
			printAuth = auth[:3] + "***" + auth[len(auth)-3:]
		} else if len(auth) == 0 {
			printAuth = "no token set"
		}

		cmd.Printf("token: %s\n", printAuth)
	},
}
