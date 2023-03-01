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
		T    time.Time         `json:"t"`
		Msg  string            `json:"msg"`
		Tags map[string]string `json:"tags"`
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
		Path     string `json:"path"`
		Name     string `json:"name"`
		Parent   string `json:"parent"`
		Revision string `json:"revision"`
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

	relativeDir := GetConfigPath()
	path := GetRelativePath(relativeDir, instance)
	path = GetPath(path)
	urlLogs := fmt.Sprintf("%s/instances/%s/logs%s", UrlPrefix, instance, query)
	clientLogs := sse.NewClient(urlLogs)
	clientLogs.Connection.Transport = &http.Transport{
		TLSClientConfig: GetTLSConfig(),
	}
	Printlog("-------INSTANCE LOGS-------")
	Printlog(urlLogs) //TODO: "bad request" can be returned
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
					if edge.Tags["state"] == "" {
						edge.Tags["state"] = " "
					}
					tags := fmt.Sprintf("(%s:%s) %s/%s\t", edge.Tags["iterator"], edge.Tags["step"], edge.Tags["name"], edge.Tags["state"])
					cmd.PrintErrf("%v: %s\n", edge.T.In(time.Local).Format("02 Jan 06 15:04 MST"), tags+edge.Msg)
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
		Fail("Failed to subscribe to messages channel: %v", err)
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

	Printlog("instance completed with status: %s\n", instanceStatus)
	return fmt.Sprintf("%s/instances/%s/output", UrlPrefix, instance)
}
