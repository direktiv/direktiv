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

type registriesList struct {
	Registries []struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"registries"`
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
	createSecret(args[1], args[2], args[0], "secrets")
}, cobra.ExactArgs(3))

var removeSecretCmd = util.GenerateCmd("delete NAMESPACE KEY", "Deletes a secret from the provided namespace", "", func(cmd *cobra.Command, args []string) {
	deleteSecrets(args[0], args[1], "secrets")
}, cobra.ExactArgs(2))

var listSecretsCmd = util.GenerateCmd("list NAMESPACE", "Returns a list of secrets for the provided namespace", "", func(cmd *cobra.Command, args []string) {
	listSecrets(args[0], "secrets")
}, cobra.ExactArgs(1))

func listSecrets(ns, t string) {

	s, err := util.DoRequest(http.MethodGet, fmt.Sprintf("/namespaces/%s/%s/",
		ns, t), util.NONECt, nil)
	if err != nil {
		log.Fatalf("error getting secrets: %v", err)
	}

	if t == "registries" {
		var r registriesList
		err = json.Unmarshal(s, &r)
		if err != nil {
			log.Fatalf("error gettting secrets: %v", err)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name"})
		for _, ss := range r.Registries {
			table.Append([]string{
				ss.Name,
			})
		}
		table.Render()
	} else {
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

	}
}

func createSecret(name, data, ns, t string) {

	var so secretsObject
	so.Name = name
	so.Data = data

	b, err := json.Marshal(&so)
	if err != nil {
		log.Fatalf("can not parse secret: %v", err)
	}

	in := string(b)
	_, err = util.DoRequest(http.MethodPost, fmt.Sprintf("/namespaces/%s/%s/", ns, t), util.JSONCt, &in)
	if err != nil {
		log.Fatalf("error creating secret: %v", err)
	}

	fmt.Printf("%s created\n", name)

}

func deleteSecrets(ns, name, t string) {

	var so secretsObject
	so.Name = name

	b, err := json.Marshal(&so)
	if err != nil {
		log.Fatalf("can not parse secret: %v", err)
	}
	in := string(b)

	_, err = util.DoRequest(http.MethodDelete, fmt.Sprintf("/namespaces/%s/%s/",
		ns, t), util.NONECt, &in)
	if err != nil {
		log.Fatalf("error getting secrets: %v", err)
	}

	fmt.Printf("%s deleted\n", name)

}
