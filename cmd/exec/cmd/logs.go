package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/r3labs/sse"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:              "logsv2",
	Short:            "Prints the logs for a instance",
	Long:             `Prints the logs for a instance. The process will continue priting logs until the Instance is stopped.`,
	PersistentPreRun: InitConfiguration,
	Run: func(cmd *cobra.Command, args []string) {
		query := "?"
		if instance != "" {
			query += "instance=" + instance
		}
		getLogsV2(cmd, query)
	},
}

type data struct {
	NextPage string
	Data     []map[string]interface{}
}

func getLogsV2(cmd *cobra.Command, query string) {
	urlGet := fmt.Sprintf("%s/logs%s", UrlPrefixV2, query)
	u, err := url.Parse(urlGet)
	if err != nil {
		cmd.PrintErr(err)
	}
	req := &http.Request{
		Method: http.MethodGet,
		URL:    u,
	}
	AddAuthHeaders(req)
	resp, err := http.DefaultClient.Do(
		req,
	)
	if err != nil {
		cmd.PrintErr(err)
	}
	defer resp.Body.Close()
	var d data
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		cmd.PrintErr(err)
	}
	err = json.Unmarshal(body, &d)
	if err != nil {
		cmd.PrintErr(err)
	}
	// if d.NextPage != "" {
	// 	cmd.Println("printing last 200 logs")
	// }
	for _, v := range d.Data {
		cmd.Println(fmt.Sprintf("%v", v))
	}
	urlsse := fmt.Sprintf("%s/logs/subscribe%s", UrlPrefixV2, query)
	clientLogs := sse.NewClient(urlsse)
	clientLogs.Connection.Transport = &http.Transport{
		TLSClientConfig: GetTLSConfig(),
	}

	AddSSEAuthHeaders(clientLogs)

	err = clientLogs.SubscribeWithContext(cmd.Context(), "message", func(msg *sse.Event) {
		data := map[string]interface{}{}

		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			cmd.PrintErr(err)
		}
		cmd.Println(fmt.Sprintf("%v", data))
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to message channel: %v\n", err)
	}
}

var instance string

func init() {
	RootCmd.AddCommand(logsCmd)
	logsCmd.Flags().StringVarP(&instance, "instance", "i", "", "Id of the instance for which to grab the logs.")
}
