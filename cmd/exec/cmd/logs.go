package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/r3labs/sse"
	"github.com/spf13/cobra"
)

var instance string

func init() {
	RootCmd.AddCommand(logsCmd)
	logsCmd.Flags().StringVarP(&instance, "instance", "i", "", "Id of the instance for which to grab the logs.")
	RootCmd.AddCommand(logsCmdSSE)
	logsCmdSSE.Flags().StringVarP(&instance, "instance", "i", "", "Id of the instance for which to grab the logs.")
}

var logsCmd = &cobra.Command{
	Use:              "logs",
	Short:            "Prints the logs for an instance",
	PersistentPreRun: InitConfiguration,
	Run: func(cmd *cobra.Command, args []string) {
		query := "?instance=" + instance
		if instance == "" {
			cmd.PrintErr("Instance parameter is required.\n")
			return
		}

		urlGet := fmt.Sprintf("%s/logs%s", UrlPrefixV2, query)
		out := func(msg string) {
			cmd.Println(msg)
		}
		// Call GetLogsV2 with the command's context and a wrapper around cmd.Printf for logging
		err := GetLogsV2(cmd.Context(), out, urlGet)
		if err != nil {
			cmd.PrintErr("Error: ", err)
		}
	},
}

var logsCmdSSE = &cobra.Command{
	Use:              "logs-follow",
	Short:            "Subscribes to logs for a instance",
	PersistentPreRun: InitConfiguration,
	Run: func(cmd *cobra.Command, args []string) {
		query := "?instance=" + instance
		if instance == "" {
			cmd.PrintErr("Instance parameter is required.\n")
			return
		}
		urlsse := fmt.Sprintf("%s/logs/subscribe%s", UrlPrefixV2, query)
		out := func(msg string) {
			cmd.Println(msg)
		}
		err := GetLogsSSE(cmd.Context(), out, urlsse)
		if err != nil {
			cmd.PrintErr("Error: ", err)
		}
	},
}

type data struct {
	NextPage string
	Data     []map[string]interface{}
}

type LoggerFunc func(msg string)

func GetLogsSSE(ctx context.Context, printToConsole LoggerFunc, urlsse string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	printToConsole("Subscribing to logs")

	clientLogs := sse.NewClient(urlsse)
	clientLogs.Connection.Transport = &http.Transport{
		TLSClientConfig: GetTLSConfig(),
	}

	AddSSEAuthHeaders(clientLogs)

	errCh := make(chan error, 1)

	go func() {
		err := clientLogs.SubscribeWithContext(ctx, "message", func(msg *sse.Event) {
			data := map[string]interface{}{}

			if err := json.Unmarshal(msg.Data, &data); err != nil {
				cancel()
				errCh <- err
				return
			}

			printToConsole(FormatLogEntry(data))

			if wf, ok := data["workflow"].(map[string]interface{}); ok && wf["status"] == "completed" {
				printToConsole("Instance SSE complete")
				cancel()
				errCh <- nil
				return
			}
		})
		if err != nil {
			errCh <- err
		}
	}()

	err := <-errCh
	return err
}

func GetLogsV2(ctx context.Context, printToConsole LoggerFunc, urlGet string) error {
	printToConsole("Getting top 200 log entries")

	u, err := url.Parse(urlGet)
	if err != nil {
		return err
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    u,
		Header: http.Header{},
	}
	AddAuthHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var d data
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &d)
	if err != nil {
		return err
	}

	for _, v := range d.Data {
		printToConsole(FormatLogEntry(v))
	}

	return nil
}

func FormatLogEntry(data map[string]interface{}) string {
	var logEntries []string
	for key, value := range data {
		// Special handling for nested maps
		if nestedMap, ok := value.(map[string]interface{}); ok {
			nestedEntries := make([]string, 0, len(nestedMap))
			for k, v := range nestedMap {
				nestedEntries = append(nestedEntries, fmt.Sprintf("%s: %v", k, v))
			}
			logEntries = append(logEntries, fmt.Sprintf("%s: {%s}", key, strings.Join(nestedEntries, ", ")))
		} else {
			logEntries = append(logEntries, fmt.Sprintf("%s: %v", key, value))
		}
	}
	return strings.Join(logEntries, " | ")
}
