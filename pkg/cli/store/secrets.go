package store

import (
	"github.com/spf13/cobra"
	"github.com/vorteil/direktiv/pkg/cli/util"
)

// CreateCommandSecrets adds secrets commands
func CreateCommandSecrets() *cobra.Command {

	cmd := util.GenerateCmd("secrets", "List, create and delete secrets from the provided namespace", "", nil, nil)

	cmd.AddCommand(createSecretCmd)
	cmd.AddCommand(removeSecretCmd)
	cmd.AddCommand(listSecretsCmd)

	return cmd

}

var createSecretCmd = util.GenerateCmd("create NAMESPACE KEY VALUE", "Creates a new secret on the provided namespace", "", func(cmd *cobra.Command, args []string) {

}, cobra.ExactArgs(3))

var removeSecretCmd = util.GenerateCmd("delete NAMESPACE KEY", "Deletes a secret from the provided namespace", "", func(cmd *cobra.Command, args []string) {

}, cobra.ExactArgs(2))

var listSecretsCmd = util.GenerateCmd("list NAMESPACE", "Returns a list of secrets for the provided namespace", "", func(cmd *cobra.Command, args []string) {

}, cobra.ExactArgs(1))
