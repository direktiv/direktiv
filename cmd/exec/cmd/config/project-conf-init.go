package config

import (
	"os"
	"path/filepath"

	root "github.com/direktiv/direktiv/cmd/exec/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var confPInitCmd = &cobra.Command{
	Use:   "project-init",
	Short: "Creates direktiv-project from current directory.",
	Long:  "Creates " + root.ConfigFileName + " in current directory to mark it as direktiv-project.",
	PreRun: func(cmd *cobra.Command, args []string) {
		err := root.InitWD()
		if err != nil {
			root.Fail(cmd, "Error: %s", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		pf := viper.GetString("directory")
		conf := make(map[string][]string)
		conf["Ignore"] = []string{""}
		data, err := yaml.Marshal(&conf)
		if err != nil {
			root.Fail(cmd, "%s", err)
		}
		path := filepath.Join(pf, root.ConfigFileName)
		err = os.WriteFile(path, data, 0o664)
		if err != nil {
			root.Fail(cmd, "%s", err)
		}
	},
}

func init() {
	root.RootCmd.AddCommand(confPInitCmd)
}
