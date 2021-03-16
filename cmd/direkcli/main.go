package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/sisatech/tablewriter"
	cobra "github.com/spf13/cobra"
	"github.com/vorteil/direktiv/pkg/cli/instance"
	log "github.com/vorteil/direktiv/pkg/cli/log"
	"github.com/vorteil/direktiv/pkg/cli/namespace"
	store "github.com/vorteil/direktiv/pkg/cli/store"
	"github.com/vorteil/direktiv/pkg/cli/util"
	"github.com/vorteil/direktiv/pkg/cli/workflow"
	"github.com/vorteil/direktiv/pkg/ingress"
	"github.com/vorteil/vorteil/pkg/elog"
	"google.golang.org/grpc"
)

var flagInputFile string
var flagGRPC string
var flagJSON bool

var conn *grpc.ClientConn
var logger elog.View
var grpcConnection = "127.0.0.1:6666"

func generateCmd(use, short, long string, fn func(cmd *cobra.Command, args []string), c cobra.PositionalArgs) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		Run:   fn,
		Args:  c,
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "direkcli",
	Short: "A CLI for interacting with a direktiv server via gRPC.",
	Long:  ``,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logger = log.GetLogger()
		var err error
		connF, err := cmd.Flags().GetString("grpc")
		if err != nil {
			return err
		}
		if connF == "" {
			connF = grpcConnection
		}

		conn, err = grpc.Dial(connF, grpc.WithInsecure())
		if err != nil {
			return err
		}

		return nil
	},
}

// namespaceCmd
var namespaceCmd = generateCmd("namespaces", "List, create and delete namespaces", "", nil, nil)

// namespaceSendEventCmd
var namespaceSendEventCmd = generateCmd("send NAMESPACE CLOUDEVENTPATH", "Send a cloud event to a namespace", "", func(cmd *cobra.Command, args []string) {
	success, err := namespace.SendEvent(conn, args[0], args[1])
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteRequestJSON(success, true, logger)
	} else {
		logger.Printf(success)
	}
}, cobra.ExactArgs(2))

// namespaceListCmd
var namespaceListCmd = generateCmd("list", "Returns a list of namespaces", "", func(cmd *cobra.Command, args []string) {
	list, err := namespace.List(conn)
	if err != nil {
		logger.Errorf("%s", err.Error())
		os.Exit(1)
	}

	if flagJSON {
		util.WriteJsonList(list, logger)
	} else {
		if len(list) == 0 {
			logger.Printf("No namespaces exist")
			return
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name"})

		for _, namespace := range list {
			table.Append([]string{
				namespace.GetName(),
			})
		}

		table.Render()
	}

}, cobra.ExactArgs(0))

// namespaceCreateCmd
var namespaceCreateCmd = generateCmd("create NAMESPACE", "Create a new namespace", "", func(cmd *cobra.Command, args []string) {
	success, err := namespace.Create(args[0], conn)
	if err != nil {
		logger.Errorf("%s", err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteRequestJSON(success, true, logger)
	} else {
		logger.Printf(success)
	}
}, cobra.ExactArgs(1))

// namespaceDeleteCmd
var namespaceDeleteCmd = generateCmd("delete NAMESPACE", "Deletes a namespace", "", func(cmd *cobra.Command, args []string) {
	success, err := namespace.Delete(args[0], conn)
	if err != nil {
		logger.Errorf("%s", err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteRequestJSON(success, true, logger)
	} else {
		logger.Printf(success)
	}
}, cobra.ExactArgs(1))

// workflowCmd
var workflowCmd = generateCmd("workflows", "List, create, get and execute workflows", "", nil, nil)

// workflowListCmd
var workflowListCmd = generateCmd("list NAMESPACE", "List all workflows under a namespace", "", func(cmd *cobra.Command, args []string) {

	list, err := workflow.List(conn, args[0])
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteJsonList(list, logger)
	} else {
		if len(list) == 0 {
			logger.Printf("No workflows exist under '%s'", args[0])
			return
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID"})

		// Build string array rows
		for _, wf := range list {
			table.Append([]string{
				wf.GetId(),
			})
		}
		table.Render()
	}

}, cobra.ExactArgs(1))

// workflowGetCmd
var workflowGetCmd = generateCmd("get NAMESPACE ID", "Get YAML of a workflow", "", func(cmd *cobra.Command, args []string) {
	success, err := workflow.Get(conn, args[0], args[1])
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteRequestJSON(success, true, logger)
	} else {
		logger.Printf(success)
	}
}, cobra.ExactArgs(2))

// workflowExecuteCmd
var workflowExecuteCmd = generateCmd("execute NAMESPACE ID", "Executes workflow with provided ID", "", func(cmd *cobra.Command, args []string) {
	input, err := cmd.Flags().GetString("input")
	if err != nil {
		logger.Errorf("unable to retrieve input flag")
		os.Exit(1)
	}

	success, err := workflow.Execute(conn, args[0], args[1], input)
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteRequestJSON(success, true, logger)
	} else {
		logger.Printf(success)
	}
}, cobra.ExactArgs(2))

var workflowToggleCmd = generateCmd("toggle NAMESPACE WORKFLOW", "Enables or disables the workflow provided", "", func(cmd *cobra.Command, args []string) {
	success, err := workflow.Toggle(conn, args[0], args[1])
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteRequestJSON(success, true, logger)
	} else {
		logger.Printf(success)
	}
}, cobra.ExactArgs(2))

// workflowAddCmd
var workflowAddCmd = generateCmd("create NAMESPACE WORKFLOW", "Creates a new workflow on provided namespace", "", func(cmd *cobra.Command, args []string) {
	// args[0] should be namespace, args[1] should be path to the workflow file
	success, err := workflow.Add(conn, args[0], args[1])
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteRequestJSON(success, true, logger)
	} else {
		logger.Printf(success)
	}
}, cobra.ExactArgs(2))

// workflowUpdateCmd
var workflowUpdateCmd = generateCmd("update NAMESPACE ID WORKFLOW", "Updates an existing workflow", "", func(cmd *cobra.Command, args []string) {
	success, err := workflow.Update(conn, args[0], args[1], args[2])
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteRequestJSON(success, true, logger)
	} else {
		logger.Printf(success)
	}
}, cobra.ExactArgs(3))

// workflowDeleteCmd
var workflowDeleteCmd = generateCmd("delete NAMESPACE ID", "Deletes an existing workflow", "", func(cmd *cobra.Command, args []string) {
	success, err := workflow.Delete(conn, args[0], args[1])
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteRequestJSON(success, true, logger)
	} else {
		logger.Printf(success)
	}
}, cobra.ExactArgs(2))

// instanceCmd
var instanceCmd = generateCmd("instances", "List, get and retrieve logs for instances", "", nil, nil)

var instanceGetCmd = generateCmd("get ID", "Get details about a workflow instance", "", func(cmd *cobra.Command, args []string) {
	resp, err := instance.Get(conn, args[0])
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteJSON(resp, logger)
	} else {
		logger.Printf("ID: %s", resp.GetId())
		logger.Printf("Input: %s", string(resp.GetInput()))
		logger.Printf("Output: %s", string(resp.GetOutput()))
	}
}, cobra.ExactArgs(1))

var instanceLogsCmd = generateCmd("logs ID", "Grabs all logs for the instance ID provided", "", func(cmd *cobra.Command, args []string) {
	logs, err := instance.Logs(conn, args[0])
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteJSON(logs, logger)
	} else {
		for _, log := range logs {
			fmt.Println(log.GetMessage())
		}
	}

}, cobra.ExactArgs(1))

var instanceListCmd = generateCmd("list NAMESPACE", "List all workflow instances from the provided namespace", "", func(cmd *cobra.Command, args []string) {
	list, err := instance.List(conn, args[0])
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}

	if flagJSON {
		util.WriteJsonList(list, logger)
	} else {
		if len(list) == 0 {
			logger.Printf("No instances exist under '%s'", args[0])
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Status"})

		// Build string array rows
		for _, instance := range list {
			table.Append([]string{
				instance.GetId(),
				instance.GetStatus(),
			})
		}
		table.Render()
	}

}, cobra.ExactArgs(1))

//registriesCmd
var registriesCmd = generateCmd("registries", "List, create and remove registries from provided namespace", "", nil, nil)

var createRegistryCmd = generateCmd("create NAMESPACE URL USER:TOKEN", "Creates a new registry on provided namespace", "", func(cmd *cobra.Command, args []string) {
	// replace : with a ! for args[2] ! is used in direktiv ! gets picked up by bash unfortunately
	args[2] = strings.ReplaceAll(args[2], ":", "!")
	storeV := store.StoreRequest{
		Key:   args[1],
		Value: args[2],
	}

	success, err := store.Create(conn, args[0], &storeV, "registry")
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteRequestJSON(success, true, logger)
	} else {
		logger.Printf(success)
	}
}, cobra.ExactArgs(3))

var removeRegistryCmd = generateCmd("delete NAMESPACE URL", "Deletes a registry from the provided namespace", "", func(cmd *cobra.Command, args []string) {
	success, err := store.Delete(conn, args[0], args[1], "registry")
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteRequestJSON(success, true, logger)
	} else {
		logger.Printf(success)
	}
}, cobra.ExactArgs(2))

var listRegistriesCmd = generateCmd("list NAMESPACE", "Returns a list of registries from the provided namespace", "", func(cmd *cobra.Command, args []string) {
	registries, err := store.List(conn, args[0], "registry")
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	castRegistries := registries.([]*ingress.GetRegistriesResponse_Registry)

	if flagJSON {
		util.WriteJsonList(castRegistries, logger)
	} else {
		if len(castRegistries) == 0 {
			logger.Printf("No registries exist under '%s'", args[0])
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Registry"})

		// Build string array rows
		for _, registry := range castRegistries {
			table.Append([]string{
				registry.GetName(),
			})
		}
		table.Render()
	}

}, cobra.ExactArgs(1))

//secretsCmd
var secretsCmd = generateCmd("secrets", "List, create and delete secrets from the provided namespace", "", nil, nil)

var createSecretCmd = generateCmd("create NAMESPACE KEY VALUE", "Creates a new secret on the provided namespace", "", func(cmd *cobra.Command, args []string) {
	storeV := store.StoreRequest{
		Key:   args[1],
		Value: args[2],
	}

	successMsg, err := store.Create(conn, args[0], &storeV, "secret")
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteRequestJSON(successMsg, true, logger)
	} else {
		logger.Printf(successMsg)
	}
}, cobra.ExactArgs(3))

var removeSecretCmd = generateCmd("delete NAMESPACE KEY", "Deletes a secret from the provided namespace", "", func(cmd *cobra.Command, args []string) {
	success, err := store.Delete(conn, args[0], args[1], "secret")
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
	if flagJSON {
		util.WriteRequestJSON(success, true, logger)
	} else {
		logger.Printf(success)
	}
}, cobra.ExactArgs(2))

var listSecretsCmd = generateCmd("list NAMESPACE", "Returns a list of secrets for the provided namespace", "", func(cmd *cobra.Command, args []string) {
	secrets, err := store.List(conn, args[0], "secret")
	if err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}

	castSecrets := secrets.([]*ingress.GetSecretsResponse_Secret)
	if flagJSON {
		util.WriteJsonList(castSecrets, logger)
	} else {
		if len(castSecrets) == 0 {
			logger.Printf("No secrets exist under '%s'", args[0])
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Secret"})

		// Build string array rows
		for _, secret := range castSecrets {
			table.Append([]string{
				secret.GetName(),
			})
		}
		table.Render()
	}

}, cobra.ExactArgs(1))

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	// Namespace command
	namespaceCmd.AddCommand(namespaceListCmd)
	namespaceCmd.AddCommand(namespaceCreateCmd)
	namespaceCmd.AddCommand(namespaceDeleteCmd)
	namespaceCmd.AddCommand(namespaceSendEventCmd)

	// Workflow commands
	workflowCmd.AddCommand(workflowAddCmd)
	workflowCmd.AddCommand(workflowDeleteCmd)
	workflowCmd.AddCommand(workflowListCmd)
	workflowCmd.AddCommand(workflowUpdateCmd)
	workflowCmd.AddCommand(workflowGetCmd)
	workflowCmd.AddCommand(workflowExecuteCmd)
	workflowCmd.AddCommand(workflowToggleCmd)

	// Workflow instance commands
	instanceCmd.AddCommand(instanceGetCmd)
	instanceCmd.AddCommand(instanceListCmd)
	instanceCmd.AddCommand(instanceLogsCmd)

	// Secrets
	secretsCmd.AddCommand(createSecretCmd)
	secretsCmd.AddCommand(removeSecretCmd)
	secretsCmd.AddCommand(listSecretsCmd)

	// Registries
	registriesCmd.AddCommand(createRegistryCmd)
	registriesCmd.AddCommand(removeRegistryCmd)
	registriesCmd.AddCommand(listRegistriesCmd)

	// Root Commands
	rootCmd.AddCommand(namespaceCmd)
	rootCmd.AddCommand(workflowCmd)
	rootCmd.AddCommand(instanceCmd)
	rootCmd.AddCommand(secretsCmd)
	rootCmd.AddCommand(registriesCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagGRPC, "grpc", "", "", "ip and port for connection GRPC default is 127.0.0.1:6666")
	rootCmd.PersistentFlags().BoolVarP(&flagJSON, "json", "", false, "provides json output")
	// workflowCmd add flag for the namespace
	workflowExecuteCmd.PersistentFlags().StringVarP(&flagInputFile, "input", "", "", "filepath to json input")
}

func main() {
	Execute()
}
