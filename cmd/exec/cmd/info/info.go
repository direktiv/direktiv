package info

import (
	"os"

	root "github.com/direktiv/direktiv/cmd/exec/cmd"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Prints basic info about command",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := os.Getwd()
		if err != nil {
			root.Fail("could not get working directory: %s", err.Error())
		}
		cmd.Printf("directory: %s\n", path)
		cmd.Printf("URL: %s\n", root.UrlPrefix)

		cf, err := root.ConfigFilePath()
		if err != nil {
			root.Fail("could not get config folder: %s", err.Error())
		}

		cmd.Printf("used config: %s\n", cf)

		pf, _ := root.ProjectFolder()
		// if err != nil {
		// 	root.Fail("could not get project folder: %s", err.Error())
		// }

		cmd.Printf("project dir: %s\n", pf)
		cmd.Printf("namespace: %s\n", root.GetNamespace())

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
