package store

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

type secretsObject struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type secretsList struct {
	Secrets []struct {
		Name string `json:"name"`
	} `json:"secrets"`
}

// CreateCommandSecrets adds secrets commands
func CreateCommandSecrets() *cobra.Command {

	cmd := util.GenerateCmd("secrets", "List, create and delete secrets from the provided namespace", "", nil, nil)

	cmd.AddCommand(createSecretCmd)
	cmd.AddCommand(removeSecretCmd)
	cmd.AddCommand(listSecretsCmd)

	return cmd

}

var createSecretCmd = util.GenerateCmd("create NAMESPACE KEY VALUE", "Creates a new secret on the provided namespace", "", func(cmd *cobra.Command, args []string) {

	var so secretsObject
	so.Name = args[1]
	so.Data = args[2]

	b, err := json.Marshal(&so)
	if err != nil {
		log.Fatalf("can not parse secret: %v", err)
	}

	in := string(b)
	_, err = util.DoRequest(http.MethodPost, fmt.Sprintf("/namespaces/%s/secrets/",
		args[0]), util.JSONCt, &in)
	if err != nil {
		log.Fatalf("error creating secret: %v", err)
	}

	fmt.Printf("secret %s created\n", args[1])

}, cobra.ExactArgs(3))

var removeSecretCmd = util.GenerateCmd("delete NAMESPACE KEY", "Deletes a secret from the provided namespace", "", func(cmd *cobra.Command, args []string) {

	var so secretsObject
	so.Name = args[1]

	b, err := json.Marshal(&so)
	if err != nil {
		log.Fatalf("can not parse secret: %v", err)
	}
	in := string(b)

	_, err = util.DoRequest(http.MethodDelete, fmt.Sprintf("/namespaces/%s/secrets/",
		args[0]), util.NONECt, &in)
	if err != nil {
		log.Fatalf("error getting secrets: %v", err)
	}

	fmt.Printf("secret %s deleted\n", args[1])

}, cobra.ExactArgs(2))

var listSecretsCmd = util.GenerateCmd("list NAMESPACE", "Returns a list of secrets for the provided namespace", "", func(cmd *cobra.Command, args []string) {

	s, err := util.DoRequest(http.MethodGet, fmt.Sprintf("/namespaces/%s/secrets/",
		args[0]), util.NONECt, nil)
	if err != nil {
		log.Fatalf("error getting secrets: %v", err)
	}

	var r secretsList
	err = json.Unmarshal(s, &r)
	if err != nil {
		log.Fatalf("error gettting secrets: %v", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name"})
	for _, ss := range r.Secrets {
		table.Append([]string{
			ss.Name,
		})
	}

	table.Render()

}, cobra.ExactArgs(1))
