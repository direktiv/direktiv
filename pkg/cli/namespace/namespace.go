package namespace

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/vorteil/direktiv/pkg/cli/util"
)

// CreateCommand create the namespace command and subcommands
func CreateCommand() *cobra.Command {

	cmd := util.GenerateCmd("namespaces", "List, create and delete namespaces", "", nil, nil)

	cmd.AddCommand(listCmd)
	cmd.AddCommand(createCmd)
	cmd.AddCommand(deleteCmd)

	return cmd

}

type namespacesList struct {
	Namespaces []struct {
		Name      string `json:"name"`
		CreatedAt struct {
			Seconds int `json:"seconds"`
			Nanos   int `json:"nanos"`
		} `json:"createdAt"`
	} `json:"namespaces"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

var listCmd = util.GenerateCmd("list", "Returns a list of namespaces", "", func(cmd *cobra.Command, args []string) {

	ns, err := util.DoRequest(http.MethodGet, "/namespaces/", util.NONECt, nil)
	if err != nil {
		log.Fatalf("error getting namespaces: %v", err)
	}

	var r namespacesList
	err = json.Unmarshal(ns, &r)
	if err != nil {
		log.Fatalf("error getting namespaces: %v", err)
	}

	if len(r.Namespaces) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name"})
		for _, namespace := range r.Namespaces {
			table.Append([]string{
				namespace.Name,
			})
		}
		table.Render()
	} else {
		log.Printf("no namespaces are available. use 'direkcli namespaces create NAMESPACE' to create one")
	}

}, cobra.ExactArgs(0))

var createCmd = util.GenerateCmd("create NAME", "Creates a namespaces", "", func(cmd *cobra.Command, args []string) {

	_, err := util.DoRequest(http.MethodPost, fmt.Sprintf("/namespaces/%s", args[0]), util.NONECt, nil)
	if err != nil {
		log.Fatalf("error creating namespace: %v", err)
	}

	log.Printf("namespace %s created", args[0])

}, cobra.ExactArgs(1))

var deleteCmd = util.GenerateCmd("delete NAME", "Deletes a namespace", "", func(cmd *cobra.Command, args []string) {

	_, err := util.DoRequest(http.MethodDelete, fmt.Sprintf("/namespaces/%s", args[0]), util.NONECt, nil)
	if err != nil {
		log.Fatalf("error deleting namespace: %v", err)
	}

	log.Printf("namespace %s deleted", args[0])

}, cobra.ExactArgs(1))
