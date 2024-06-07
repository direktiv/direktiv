package config

import (
	"os"
	"path/filepath"
	"strings"

	root "github.com/direktiv/direktiv/cmd/exec/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var confInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a new profile-configuration.",
	Long: "Creates a new " + root.DefaultProfileConfigName + " in the ~/" + root.DefaultProfileConfigPath + ` directory use the global flags to populate the config values. If a old configuration-file exists it will be renamed to ` + root.DefaultProfileConfigPath + `.bak

Example : init -P myserver --addr 192.168.122.232 -n ns
`,

	Run: func(cmd *cobra.Command, args []string) {
		profile := viper.GetString("profile")
		if profile == "" {
			root.Fail(cmd, "Error: Pls provide a profile name via the profile flag")
		}
		addr := viper.GetString("addr")
		if addr == "" {
			root.Fail(cmd, "Error: Pls provide the addr of the direktiv server via the addr flag")
		}
		ns := viper.GetString("namespace")
		if ns == "" {
			root.Fail(cmd, "Error: Pls provide the namespace of the direktiv server via the namespace flag")
		}
		cf, err := root.GetConfigFilePath()
		if err != nil {
			root.Fail(cmd, "Could not parse config directory: %s", err.Error())
		}
		err = backupOldConfIfItExists(cf)
		if err != nil {
			root.Fail(cmd, "Error whil creating a backup of the old config, %s", err)
		}

		err = os.MkdirAll(filepath.Dir(cf), os.ModePerm)
		if err != nil {
			root.Fail(cmd, "Could not create config directory: %s", err.Error())
		}

		profileConf := root.ProfileConfig{
			Addr:      addr,
			Namespace: ns,
			Auth:      root.GetAuth(),
			MaxSize:   root.GetMaxSize(),
		}

		profiles := make(map[string]root.ProfileConfig)
		profiles[profile] = profileConf

		data, err := yaml.Marshal(&profiles)
		if err != nil {
			root.Fail(cmd, "%s", err)
		}
		err = os.WriteFile(cf, data, 0o664)
		if err != nil {
			root.Fail(cmd, "%s", err)
		}
	},
}

var confAddCmd = &cobra.Command{
	Use: "add",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := root.LoadProfileConfig()
		if err != nil {
			root.Fail(cmd, "Got an error: %v", err)
		}
	},
	Short: "Add a profile to a existing profile-configuration.",
	Long: "Add a profile to a existing " + root.DefaultProfileConfigName + ` use the global flags to populate the config values.

Examples: 	add-config -P myserver --addr 192.168.122.232 -n ns
			add-config -P myserver --addr http://192.168.122.232 -n ns
			add-config -P myserver --addr https://192.168.122.232 -n ns

If the profile exists it will be overwritten.
`,

	Run: func(cmd *cobra.Command, args []string) {
		profile := viper.GetString("profile")
		if profile == "" {
			root.Fail(cmd, "Error: Pls provide a profile name via the profile flag")
		}
		addr := viper.GetString("addr")
		if addr == "" {
			root.Fail(cmd, "Error: Pls provide the addr of the direktiv server via the addr flag")
		}
		if !strings.HasPrefix(addr, "http") {
			addr = "http://" + addr
		}
		ns := viper.GetString("namespace")
		if ns == "" {
			root.Fail(cmd, "Error: Pls provide the namespace of the direktiv server via the namespace flag")
		}
		cf, err := root.GetConfigFilePath()
		if err != nil {
			root.Fail(cmd, "Could not parse config directory: %s", err.Error())
		}

		profileConf := root.ProfileConfig{
			Addr:      addr,
			Namespace: ns,
			Auth:      root.GetAuth(),
			MaxSize:   root.GetMaxSize(),
		}

		profiles := root.Config.Profiles
		profiles[profile] = profileConf

		data, err := yaml.Marshal(&profiles)
		if err != nil {
			root.Fail(cmd, "%s", err)
		}
		err = os.WriteFile(cf, data, 0o664)
		if err != nil {
			root.Fail(cmd, "%s", err)
		}
	},
}

func backupOldConfIfItExists(cf string) error {
	_, err := os.Stat(cf)
	if err == nil {
		e := os.Rename(cf, cf+".bak")
		if e != nil {
			return e
		}
	}
	return nil
}
