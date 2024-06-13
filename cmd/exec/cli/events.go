package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var eventCmd = &cobra.Command{
	Use:   "events",
	Short: "Sends and lists events",
}

func init() {
	RootCmd.AddCommand(eventCmd)
	eventCmd.AddCommand(sendEventCmd)
	// configCmd.AddCommand(confAddCmd)
	// configCmd.AddCommand(deleteProfileCmd)
}

var sendEventCmd = &cobra.Command{
	Use:   "send [path to event file]",
	Args:  cobra.ExactArgs(1),
	Short: "Sends the file as cloudevent to Direktiv.",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := prepareCommand(cmd)
		if err != nil {
			return err
		}

		fmt.Println(p)

		return nil
	},
}
