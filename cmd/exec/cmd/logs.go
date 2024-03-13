package cmd

import (
	"context"
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
		query := "?instance=" + instance
		if instance == "" {
			cmd.PrintErr("instance param is required /n")
			return
		}
		GetLogsV2(cmd, query)
	},
}

type data struct {
	NextPage string
	Data     []map[string]interface{}
}

func GetLogsSSE(cmd *cobra.Command, query string) {
	urlGet := fmt.Sprintf("%s/logs%s", UrlPrefixV2, query)
	cmd.Println(fmt.Sprintf("getting top 200 log entries %v", urlGet))
	ctx, cancel := context.WithCancel(context.TODO())

	urlsse := fmt.Sprintf("%s/logs/subscribe%s", UrlPrefixV2, query)
	clientLogs := sse.NewClient(urlsse)
	clientLogs.Connection.Transport = &http.Transport{
		TLSClientConfig: GetTLSConfig(),
	}

	AddSSEAuthHeaders(clientLogs)
	err := clientLogs.SubscribeWithContext(ctx, "message", func(msg *sse.Event) {
		data := map[string]interface{}{}

		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			cmd.PrintErr(err)
		}
		cmd.Println(fmt.Sprintf("%v", data))
		wf := data["workflow"]
		wfContext, ok := wf.(map[string]interface{})
		if ok && wfContext["status"] == "completed" {
			ctx.Done()
			cmd.Println("instance SSE complete")
			cancel()
			return
		}
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to message channel: %v\n", err)
	}
}

func GetLogsV2(cmd *cobra.Command, query string) {
	urlGet := fmt.Sprintf("%s/logs%s", UrlPrefixV2, query)
	cmd.Println(fmt.Sprintf("getting top 200 log entries %v", urlGet))
	u, err := url.Parse(urlGet)
	if err != nil {
		cmd.PrintErr(err)
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    u,
		Header: http.Header{},
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

	for _, v := range d.Data {
		cmd.Println(fmt.Sprintf("%v /n", v))
	}
}

var instance string

func init() {
	RootCmd.AddCommand(logsCmd)
	logsCmd.Flags().StringVarP(&instance, "instance", "i", "", "Id of the instance for which to grab the logs.")
}
