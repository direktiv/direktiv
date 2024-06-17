package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.SilenceErrors = true

	RootCmd.PersistentFlags().StringP("profile", "p", "", "Name of the profile")
	RootCmd.PersistentFlags().StringP("address", "a", "", "Target direktiv api host address")
	RootCmd.PersistentFlags().StringP("namespace", "n", "", "Target namespace for Direktiv")
	RootCmd.PersistentFlags().StringP("token", "t", "", "Authenticate request with token")
	RootCmd.PersistentFlags().Bool("insecure", true, "Accept insecure https connections")

	viper.SetConfigName("profiles")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.direktiv")
	viper.AddConfigPath(".")
}

var RootCmd = &cobra.Command{
	Use: "direktivctl",
}

// nolint: errcheck
func bindArgs() {
	viper.SetEnvPrefix("direktiv")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	viper.BindEnv(fmt.Sprintf("%s.token", RootCmd.PersistentFlags().Lookup("profile").Value.String()), "DIREKTIV_TOKEN")
	viper.BindEnv(fmt.Sprintf("%s.address", RootCmd.PersistentFlags().Lookup("profile").Value.String()), "DIREKTIV_ADDRESS")
	viper.BindEnv(fmt.Sprintf("%s.namespace", RootCmd.PersistentFlags().Lookup("profile").Value.String()), "DIREKTIV_NAMESPACE")
	viper.BindEnv(fmt.Sprintf("%s.insecure", RootCmd.PersistentFlags().Lookup("profile").Value.String()), "DIREKTIV_INSECURE")

	viper.BindPFlag(fmt.Sprintf("%s.address", RootCmd.PersistentFlags().Lookup("profile").Value.String()), RootCmd.PersistentFlags().Lookup("address"))
	viper.BindPFlag(fmt.Sprintf("%s.token", RootCmd.PersistentFlags().Lookup("profile").Value.String()), RootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag(fmt.Sprintf("%s.insecure", RootCmd.PersistentFlags().Lookup("profile").Value.String()), RootCmd.PersistentFlags().Lookup("insecure"))
	viper.BindPFlag(fmt.Sprintf("%s.namespace", RootCmd.PersistentFlags().Lookup("profile").Value.String()), RootCmd.PersistentFlags().Lookup("namespace"))
}

func readConfig() error {
	if err := viper.ReadInConfig(); err != nil {
		// nolint: errorlint
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
		// first time run with no config file
		dir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		dir = filepath.Join(dir, ".direktiv")

		err = os.MkdirAll(dir, 0o755)
		if err != nil {
			return err
		}
		_, err = os.Create(filepath.Join(dir, "profiles.yaml"))
		if err != nil {
			return err
		}

		return viper.ReadInConfig()
	}

	return nil
}

func prepareCommand(cmd *cobra.Command) (profile, error) {
	p := profile{}

	err := readConfig()
	if err != nil {
		return p, err
	}
	bindArgs()
	cmd.Printf("config file: %s\n", viper.ConfigFileUsed())

	var profiles map[string]profile
	err = viper.Unmarshal(&profiles)
	if err != nil {
		return p, err
	}

	name := RootCmd.PersistentFlags().Lookup("profile").Value.String()

	p = profiles[name]

	// check for namespace
	if p.Namespace == "" {
		return p, fmt.Errorf("no namespace provided or profile name incorrect")
	}

	// check for url
	if p.Address == "" {
		return p, fmt.Errorf("no address provided or profile name incorrect")
	}

	return p, nil
}
