package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/r3labs/sse"
	"github.com/spf13/cobra"
)

type LogResponse struct {
	PageInfo struct {
		TotalCount int `json:"total"`
		Limit      int `json:"limit"`
		Offset     int `json:"offset"`
	} `json:"pageInfo"`
	Results []struct {
		T     time.Time         `json:"t"`
		Msg   string            `json:"msg"`
		Level string            `json:"level"`
		Tags  map[string]string `json:"tags"`
	} `json:"results"`
	Namespace string `json:"namespace"`
	Instance  string `json:"instance"`
}

type InstanceResponse struct {
	Namespace string `json:"namespace"`
	Instance  struct {
		CreatedAt    time.Time `json:"createdAt"`
		UpdatedAt    time.Time `json:"updatedAt"`
		ID           string    `json:"id"`
		As           string    `json:"as"`
		Status       string    `json:"status"`
		ErrorCode    string    `json:"errorCode"`
		ErrorMessage string    `json:"errorMessage"`
	} `json:"instance"`
	InvokedBy string   `json:"invokedBy"`
	Flow      []string `json:"flow"`
	Workflow  struct {
		Path   string `json:"path"`
		Name   string `json:"name"`
		Parent string `json:"parent"`
	} `json:"workflow"`
}

type FilterQueryInstance struct {
	Typ     string
	Filter  string
	Payload []string
}

func (fq FilterQueryInstance) Query() string {
	value := ""
	for i, v := range fq.Payload {
		value += v
		if i < len(fq.Payload)-1 {
			value += "::"
		}
	}
	return fmt.Sprintf("?filter.field=%s&filter.type=%s&filter.val=%s", fq.Filter, fq.Typ, value)
}

func GetLogs(cmd *cobra.Command, instance string, query string) (urlOutput string) {
	instanceStatus := "pending"

	urlLogs := fmt.Sprintf("%s/instances/%s/logs%s", UrlPrefix, instance, query)
	clientLogs := sse.NewClient(urlLogs)
	clientLogs.Connection.Transport = &http.Transport{
		TLSClientConfig: GetTLSConfig(),
	}
	cmd.Println("-------INSTANCE LOGS-------")
	cmd.Println(urlLogs)
	cmd.Println("---------------------------")
	AddSSEAuthHeaders(clientLogs)

	logsChannel := make(chan *sse.Event)
	err := clientLogs.SubscribeChan("messages", logsChannel)
	if err != nil {
		log.Fatalf("Failed to subscribe to messages channel: %v\n", err)
	}

	// Get Logs
	go func() {
		for {
			msg := <-logsChannel
			if msg == nil {
				break
			}

			// Skip heartbeat
			if len(msg.Data) == 0 {
				continue
			}

			var logResp LogResponse
			err = json.Unmarshal(msg.Data, &logResp)
			if err != nil {
				log.Fatalln(err)
			}

			if len(logResp.Results) > 0 {
				for _, edge := range logResp.Results {
					prefix := ""
					if len(edge.Tags) > 0 {
						prefix = buildPrefix(edge.Tags)
					}
					prefix = printFormated(edge.Level) + prefix
					//nolint:gosmopolitan
					cmd.Printf("%v: %s %s\n", edge.T.In(time.Local).Format("02 Jan 06 15:04 MST"), prefix, edge.Msg)
				}
			}
		}
	}()

	urlInstance := fmt.Sprintf("%s/instances/%s", UrlPrefix, instance)
	clientInstance := sse.NewClient(urlInstance)
	clientInstance.Connection.Transport = &http.Transport{
		TLSClientConfig: GetTLSConfig(),
	}

	AddSSEAuthHeaders(clientInstance)

	channelInstance := make(chan *sse.Event)
	err = clientInstance.SubscribeChan("messages", channelInstance)
	if err != nil {
		Fail(cmd, "Failed to subscribe to messages channel: %v", err)
	}

	for {
		msg := <-channelInstance
		if msg == nil {
			break
		}

		// Skip heartbeat
		if len(msg.Data) == 0 {
			continue
		}

		var instanceResp InstanceResponse
		err = json.Unmarshal(msg.Data, &instanceResp)
		if err != nil {
			log.Fatalf("Failed to read instance response: %v\n", err)
		}

		if instanceResp.Instance.Status != instanceStatus {
			time.Sleep(500 * time.Millisecond)
			instanceStatus = instanceResp.Instance.Status
			clientLogs.Unsubscribe(logsChannel)
			clientInstance.Unsubscribe(channelInstance)
			break
		}
	}

	cmd.Printf("instance completed with status: %s\n", instanceStatus)
	return fmt.Sprintf("%s/instances/%s/output", UrlPrefix, instance)
}

func buildPrefix(tags map[string]string) string {
	if tags["state-id"] == "" {
		tags["state-id"] = " "
	}
	caller := fmt.Sprintf("%s/%s", tags["workflow"], tags["state-id"])
	loop_index := ""
	if val, ok := tags["loop-index"]; ok {
		loop_index = "/i-" + val
	}
	prefix := caller + loop_index
	prefixLen := len(prefix)

	// if prefixLen < 8 {
	// 	prefix += "\t"
	// }
	// if prefixLen < 12 {
	// 	prefix += "\t"
	// }
	for i := prefixLen; i < 25; i++ {
		prefix += " "
	}
	if prefixLen < 24 {
		prefix += "\t"
	}
	if prefixLen < 32 {
		prefix += "\t"
	}
	return prefix
}

func printFormated(level string) string {
	switch level {
	case "debug":
		return ""
	case "info":
		return ""
	case "error":
		return "\033[0;31m"
	case "fatal":
		return "\033[1;95m"
	}
	return ""
}
