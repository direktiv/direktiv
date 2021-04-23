package instance

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/vorteil/direktiv/pkg/cli/util"
)

type instanceList struct {
	Workflowinstances []struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		Begintime struct {
			Seconds int `json:"seconds"`
			Nanos   int `json:"nanos"`
		} `json:"beginTime"`
	} `json:"workflowInstances"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type instanceObject struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Invokedby string `json:"invokedBy"`
	Revision  int    `json:"revision"`
	Begintime struct {
		Seconds int `json:"seconds"`
		Nanos   int `json:"nanos"`
	} `json:"beginTime"`
	Endtime struct {
		Seconds int `json:"seconds"`
		Nanos   int `json:"nanos"`
	} `json:"endTime"`
	Flow   []string `json:"flow"`
	Input  string   `json:"input"`
	Output string   `json:"output"`
}

type instanceLogs struct {
	Workflowinstancelogs []struct {
		Timestamp struct {
			Seconds int `json:"seconds"`
		} `json:"timestamp"`
		Message string `json:"message"`
	} `json:"workflowInstanceLogs"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// CreateCommand adds instance commands
func CreateCommand() *cobra.Command {

	cmd := util.GenerateCmd("instances", "List, get and retrieve logs for instances", "", nil, nil)

	cmd.AddCommand(instanceGetCmd)
	cmd.AddCommand(instanceListCmd)
	cmd.AddCommand(instanceLogsCmd)

	return cmd

}

var instanceGetCmd = util.GenerateCmd("get ID", "Get details about a workflow instance", "", func(cmd *cobra.Command, args []string) {

	i, err := util.DoRequest(http.MethodGet, fmt.Sprintf("/instances/%s", args[0]),
		util.NONECt, nil)
	if err != nil {
		log.Fatalf("error getting instance: %v", err)
	}

	var io instanceObject
	err = json.Unmarshal(i, &io)
	if err != nil {
		log.Fatalf("error getting instance: %v", err)
	}

	in, err := base64.StdEncoding.DecodeString(io.Input)
	if err != nil {
		log.Fatalf("can not decode workflow: %v, %v", err, string(i))
	}

	out, err := base64.StdEncoding.DecodeString(io.Output)
	if err != nil {
		log.Fatalf("can not decode workflow: %v, %v", err, string(i))
	}

	fmt.Printf("Input: %v\nOutput: %v", string(in), string(out))

}, cobra.ExactArgs(1))

var instanceLogsCmd = util.GenerateCmd("logs ID", "Gets all logs for the instance ID provided", "", func(cmd *cobra.Command, args []string) {

	i, err := util.DoRequest(http.MethodGet, fmt.Sprintf("/instances/%s/logs?offset=0&limit=300", args[0]),
		util.NONECt, nil)
	if err != nil {
		log.Fatalf("error getting instance: %v", err)
	}
	var il instanceLogs
	err = json.Unmarshal(i, &il)
	if err != nil {
		log.Fatalf("error getting instance: %v", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Time", "Log"})
	for _, l := range il.Workflowinstancelogs {
		t := time.Unix(int64(l.Timestamp.Seconds), 0)
		table.Append([]string{
			t.String(),
			l.Message,
		})
	}
	table.Render()

}, cobra.ExactArgs(1))

var instanceListCmd = util.GenerateCmd("list NAMESPACE", "List all workflow instances from the provided namespace", "", func(cmd *cobra.Command, args []string) {

	i, err := util.DoRequest(http.MethodGet, fmt.Sprintf("/instances/%s", args[0]),
		util.NONECt, nil)
	if err != nil {
		log.Fatalf("error getting instances: %v", err)
	}

	var il instanceList
	err = json.Unmarshal(i, &il)
	if err != nil {
		log.Fatalf("error getting instances: %v", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Status"})
	for _, instances := range il.Workflowinstances {
		table.Append([]string{
			instances.ID,
			instances.Status,
		})
	}
	table.Render()

}, cobra.ExactArgs(1))
