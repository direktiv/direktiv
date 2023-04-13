package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var UrlPrefix string

func init() {
	RootCmd.PersistentFlags().StringP("profile", "P", "", "Select the named profile from the loaded multi-profile configuration file.")
	RootCmd.PersistentFlags().StringP("config", "C", "", "Point to a specific configuration file.")

	RootCmd.PersistentFlags().StringP("addr", "a", "", "Target direktiv api address.")
	RootCmd.PersistentFlags().StringP("namespace", "n", "", "Target namespace to execute workflow on.")
	RootCmd.PersistentFlags().StringP("auth", "t", "", "Authenticate request with token or apikey.")
	RootCmd.PersistentFlags().Bool("insecure", true, "Accept insecure https connections")

	err := viper.BindPFlags(RootCmd.PersistentFlags())
	if err != nil {
		Fail("error binding configuration flags: %v", err)
	}

	viper.SetEnvPrefix("direktiv")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

var RootCmd = &cobra.Command{
	Use: "direktivctl",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		loadConfig(cmd)
		cmdPrepareSharedValues()
		if err := pingNamespace(); err != nil {
			log.Fatalf("%v", err)
		}
	},
}

func cmdPrepareSharedValues() {
	// Load Config From flags / config
	addr := getAddr()
	namespace := GetNamespace()

	if cfgMaxSize := viper.GetInt64("max-size"); cfgMaxSize > 0 {
		maxSize = cfgMaxSize
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = GetTLSConfig()

	UrlPrefix = fmt.Sprintf("%s/api/namespaces/%s", strings.Trim(addr, "/"), strings.Trim(namespace, "/"))
}
