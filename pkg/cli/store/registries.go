package store

import (
	"github.com/spf13/cobra"
	"github.com/vorteil/direktiv/pkg/cli/util"
)

// CreateCommandRegistries adds registry commands
func CreateCommandRegistries() *cobra.Command {

	cmd := util.GenerateCmd("registries", "List, create and remove registries from provided namespace", "", nil, nil)

	cmd.AddCommand(createRegistryCmd)
	cmd.AddCommand(removeRegistryCmd)
	cmd.AddCommand(listRegistriesCmd)

	return cmd

}

var createRegistryCmd = util.GenerateCmd("create NAMESPACE URL USER:TOKEN", "Creates a new registry on provided namespace", "", func(cmd *cobra.Command, args []string) {
	createSecret(args[1], args[2], args[0], "registries")
}, cobra.ExactArgs(3))

var removeRegistryCmd = util.GenerateCmd("delete NAMESPACE URL", "Deletes a registry from the provided namespace", "", func(cmd *cobra.Command, args []string) {
	deleteSecrets(args[0], args[1], "registries")
}, cobra.ExactArgs(2))

var listRegistriesCmd = util.GenerateCmd("list NAMESPACE", "Returns a list of registries from the provided namespace", "", func(cmd *cobra.Command, args []string) {
	listSecrets(args[0], "registries")
}, cobra.ExactArgs(1))
