package info

import (
	"os"

	root "github.com/direktiv/direktiv/cmd/exec/cmd"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Prints detected configuration values for current project",
	Run: func(cmd *cobra.Command, args []string) {
		pf, err := root.ProjectFolder()
		if err != nil {
			root.Fail("could not get project directory: %s", err.Error())
		}

		cmd.Printf("project directory: %s\n", pf)

		pwd, err := os.Getwd()
		if err != nil {
			root.Fail("could not get working directory: %s", err.Error())
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

func init() {
	root.RootCmd.AddCommand(infoCmd)
}
