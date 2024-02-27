package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	UrlPrefix   string
	UrlPrefixV2 string
)

func init() {
	RootCmd.PersistentFlags().StringP("profile", "P", "", "Select the named profile configuration file.")

	RootCmd.PersistentFlags().StringP("addr", "a", "", "Target direktiv api address.")
	RootCmd.PersistentFlags().StringP("namespace", "n", "", "Target namespace to execute workflow on.")
	RootCmd.PersistentFlags().StringP("auth", "t", "", "Authenticate request with token or apikey.")
	RootCmd.PersistentFlags().Bool("insecure", true, "Accept insecure https connections.")

	err := viper.BindPFlags(RootCmd.PersistentFlags())
	if err != nil {
		panic(fmt.Errorf("error binding configuration flags: %w", err))
	}

	viper.SetEnvPrefix("direktiv")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

var RootCmd = &cobra.Command{
	Use: ToolName,
}

func cmdPrepareSharedValues() {
	// Load Config From flags / config
	addr := getAddr()
	namespace := GetNamespace()

	maxSize = GetMaxSize()
	http.DefaultTransport.(*http.Transport).TLSClientConfig = GetTLSConfig()

	UrlPrefix = fmt.Sprintf("%s/api/namespaces/%s", strings.Trim(addr, "/"), strings.Trim(namespace, "/"))
	UrlPrefixV2 = fmt.Sprintf("%s/api/v2/namespaces/%s", strings.Trim(addr, "/"), strings.Trim(namespace, "/"))

}
