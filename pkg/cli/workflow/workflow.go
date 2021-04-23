package workflow

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/vorteil/direktiv/pkg/cli/util"
)

type workflowObject struct {
	UID       string `json:"uid"`
	ID        string `json:"id"`
	Revision  int    `json:"revision"`
	Active    bool   `json:"active"`
	Workflow  string `json:"workflow"`
	Createdat struct {
		Seconds int `json:"seconds"`
		Nanos   int `json:"nanos"`
	} `json:"createdAt"`
}

type workflowList struct {
	Workflows []workflowObject
	Offset    int `json:"offset"`
	Limit     int `json:"limit"`
	Total     int `json:"total"`
}

type executed struct {
	Instanceid string `json:"instanceId"`
}

// CreateCommand create the namespace command and subcommands
func CreateCommand() *cobra.Command {

	cmd := util.GenerateCmd("workflows", "List, create, get and execute workflows", "", nil, nil)

	cmd.AddCommand(workflowAddCmd)
	cmd.AddCommand(workflowUpdateCmd)
	cmd.AddCommand(workflowDeleteCmd)
	cmd.AddCommand(workflowListCmd)
	cmd.AddCommand(workflowGetCmd)
	cmd.AddCommand(workflowExecuteCmd)
	cmd.AddCommand(workflowToggleCmd)

	return cmd

}

var workflowGetCmd = util.GenerateCmd("get NAMESPACE NAME", "Get YAML of a workflow", "", func(cmd *cobra.Command, args []string) {

	wf, err := util.DoRequest(http.MethodGet, fmt.Sprintf("/namespaces/%s/workflows/%s",
		args[0], args[1]), util.NONECt, nil)
	if err != nil {
		log.Fatalf("error toggling workflow: %v", err)
	}

	var workflow workflowObject
	err = json.Unmarshal(wf, &workflow)
	if err != nil {
		log.Fatalf("can not parse response: %v, %v", err, string(wf))
	}

	d, err := base64.StdEncoding.DecodeString(workflow.Workflow)
	if err != nil {
		log.Fatalf("can not decode workflow: %v, %v", err, string(wf))
	}

	fmt.Printf("%v", string(d))

}, cobra.ExactArgs(2))

var workflowToggleCmd = util.GenerateCmd("toggle NAMESPACE WORKFLOW", "Enables or disables the workflow provided", "", func(cmd *cobra.Command, args []string) {

	wf, err := util.DoRequest(http.MethodPut, fmt.Sprintf("/namespaces/%s/workflows/%s/toggle",
		args[0], args[1]), util.NONECt, nil)
	if err != nil {
		log.Fatalf("error toggling workflow: %v", err)
	}

	var w workflowObject
	err = json.Unmarshal(wf, &w)
	if err != nil {
		log.Fatalf("can not parse response: %v, %v", err, string(wf))
	}

	fmt.Printf("workflow %s enabled: %v\n", w.ID, w.Active)

}, cobra.ExactArgs(2))

var workflowAddCmd = util.GenerateCmd("create NAMESPACE FILE", "Creates a new workflow on provided namespace", "", func(cmd *cobra.Command, args []string) {

	// read file
	f, err := ioutil.ReadFile(args[1])
	if err != nil {
		log.Fatalf("can not create workflow: %v", err)
	}
	st := string(f)

	wf, err := util.DoRequest(http.MethodPost, fmt.Sprintf("/namespaces/%s/workflows",
		args[0]), util.YAMLCt, &st)
	if err != nil {
		log.Fatalf("error creating workflow: %v", err)
	}

	var workflow workflowObject
	err = json.Unmarshal(wf, &workflow)
	if err != nil {
		log.Fatalf("can not parse response: %v, %v", err, string(wf))
	}

	log.Printf("workflow '%s' created", workflow.ID)

}, cobra.ExactArgs(2))

var workflowUpdateCmd = util.GenerateCmd("update NAMESPACE NAME FILE", "Updates an existing workflow", "", func(cmd *cobra.Command, args []string) {

	// read file
	f, err := ioutil.ReadFile(args[2])
	if err != nil {
		log.Fatalf("can not create workflow: %v", err)
	}
	st := string(f)

	wf, err := util.DoRequest(http.MethodPut, fmt.Sprintf("/namespaces/%s/workflows/%s",
		args[0], args[1]), util.YAMLCt, &st)
	if err != nil {
		log.Fatalf("error updating workflow: %v", err)
	}

	var workflow workflowObject
	err = json.Unmarshal(wf, &workflow)
	if err != nil {
		log.Fatalf("can not parse response: %v, %v", err, string(wf))
	}

	log.Printf("workflow '%s' updated", workflow.ID)

}, cobra.ExactArgs(3))

var workflowDeleteCmd = util.GenerateCmd("delete NAMESPACE NAME", "Deletes an existing workflow", "", func(cmd *cobra.Command, args []string) {

	_, err := util.DoRequest(http.MethodDelete, fmt.Sprintf("/namespaces/%s/workflows/%s",
		args[0], args[1]), util.NONECt, nil)
	if err != nil {
		log.Fatalf("error creating workflow: %v", err)
	}

	fmt.Printf("workflow %s deleted\n", args[1])

}, cobra.ExactArgs(2))

var workflowListCmd = util.GenerateCmd("list NAMESPACE", "List all workflows under a namespace", "", func(cmd *cobra.Command, args []string) {

	wfs, err := util.DoRequest(http.MethodGet, fmt.Sprintf("/namespaces/%s/workflows/",
		args[0]), util.NONECt, nil)
	if err != nil {
		log.Fatalf("error getting workflows: %v", err)
	}

	var w workflowList
	err = json.Unmarshal(wfs, &w)
	if err != nil {
		log.Fatalf("error getting workflows: %v", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name"})
	for _, workflow := range w.Workflows {
		table.Append([]string{
			workflow.ID,
		})
	}
	table.Render()

}, cobra.ExactArgs(1))

// workflowExecuteCmd
var workflowExecuteCmd = util.GenerateCmd("execute NAMESPACE ID [INPUT FILE]", "Executes workflow with provided ID", "", func(cmd *cobra.Command, args []string) {

	var st string
	if len(args) > 2 {
		f, err := ioutil.ReadFile(args[2])
		if err != nil {
			log.Fatalf("can not create workflow: %v", err)
		}
		st = string(f)
	}

	exe, err := util.DoRequest(http.MethodPost, fmt.Sprintf("/namespaces/%s/workflows/%s/execute",
		args[0], args[1]), util.NONECt, &st)
	if err != nil {
		log.Fatalf("error creating workflow: %v", err)
	}

	var e executed
	err = json.Unmarshal(exe, &e)
	if err != nil {
		log.Fatalf("error executing workflows: %v", err)
	}

	fmt.Printf("%s\n", e.Instanceid)

}, cobra.MinimumNArgs(2))
