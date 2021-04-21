package main

import (
	"fmt"
	"log"
	"os"

	cobra "github.com/spf13/cobra"
	// log "github.com/vorteil/direktiv/pkg/cli/log"
	"github.com/vorteil/direktiv/pkg/cli/instance"
	"github.com/vorteil/direktiv/pkg/cli/namespace"
	store "github.com/vorteil/direktiv/pkg/cli/store"
	"github.com/vorteil/direktiv/pkg/cli/util"
	"github.com/vorteil/direktiv/pkg/cli/workflow"
)

var (
	flagURL  string
	flagSkip bool
)

const (
	endpoint = "DIREKTIV_CLI_ENDPOINT"
)

func generateCmd(use, short, long string, fn func(cmd *cobra.Command, args []string), c cobra.PositionalArgs) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		Run:   fn,
		Args:  c,
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "direkcli",
	Short: "A CLI for interacting with a direktiv server via API.",
	Long:  ``,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		util.DirektivURL = "http://localhost"

		if os.Getenv(endpoint) != "" {
			util.DirektivURL = os.Getenv(endpoint)
		}

		urlValue, err := cmd.Flags().GetString("url")
		if err != nil {
			log.Printf("url arg invalid: %v", err)
		}
		if len(urlValue) > 0 {
			util.DirektivURL = urlValue
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	// Root Commands
	rootCmd.AddCommand(namespace.CreateCommand())
	rootCmd.AddCommand(workflow.CreateCommand())
	rootCmd.AddCommand(store.CreateCommandRegistries())
	rootCmd.AddCommand(store.CreateCommandSecrets())
	rootCmd.AddCommand(instance.CreateCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagURL, "url", "", "", "name and port for connection, default is 127.0.0.1:80. Overwrite with env DIREKTIV_CLI_ENDPOINT")
	rootCmd.PersistentFlags().BoolVarP(&flagSkip, "skipVerify", "", false, "skip certificate validation")
}

func main() {

	log.SetFlags(0)
	log.SetOutput(new(util.CLILogWriter))

	Execute()
}
