package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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

		uploader, err := newUploader("", p)
		if err != nil {
			return err
		}

		b, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}

		url := fmt.Sprintf("%s/api/v2/namespaces/%s/events/broadcast", p.Address, p.Namespace)
		resp, err := uploader.sendRequest("POST", url, b)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			var errJson errorResponse
			err = json.Unmarshal(b, &errJson)
			if err != nil {
				return err
			}

			return fmt.Errorf(errJson.Error.Message)
		}

		fmt.Println("event sent")

		return nil
	},
}
