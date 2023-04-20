package logs

import (
	root "github.com/direktiv/direktiv/cmd/exec/cmd"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Prints the logs for a instance and all child instances",
	Long: `Prints the logs for a instance and all child instances. The process will continue priting logs until the Instance is stopped.
EXAMPLE: logs --addr http://192.168.1.1 --namespace admin --instance 9b0e45b5-5e7e-4006-93e7-6764e8379e98
Use the additional flags to can to retrieve and watch logs from a specific instance whithout Child instances. 
EXAMPLE: logs --addr http://192.168.1.1 --namespace admin --instance 9b0e45b5-5e7e-4006-93e7-6764e8379e98 --filter ID --type MATCH 9b0e45b5-5e7e-4006-93e7-6764e8379e98
It is also possible to watch logs of a specific workflow, state-id and array-index by providing the values as arguments the array-index is optional
EXAMPLE: logs --addr http://192.168.1.1 --namespace admin -instance 9b0e45b5-5e7e-4006-93e7-6764e8379e98 --filter QUERY --type MATCH getwf getter 2
To filter by the level of the log use -filter LEVEL --type MATCH level or -filter LEVEL --type STARTING info, supported levels are debug, info, error and panic
`,
	PersistentPreRun: root.InitConfiguration,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := uuid.Parse(instance)
		if err != nil {
			cmd.PrintErrln("instance id should be formated as 9b0e45b5-5e7e-4006-93e7-6764e8379e98")
			cmd.Print(cmd.Help())
			return
		}
		var fq root.FilterQueryInstance
		fq.Payload = args
		fq.Filter = filter
		fq.Typ = filterTyp
		query := ""
		if fq.Filter != "" && fq.Typ != "" {
			query = fq.Query()
		}
		root.GetLogs(cmd, instance, query)
	},
}

var (
	filterTyp string
	filter    string
	instance  string
)

func init() {
	root.RootCmd.AddCommand(logsCmd)
	logsCmd.Flags().StringVarP(&instance, "instance", "i", "", "Id of the instance for which to grab the logs.")
	logsCmd.Flags().StringVarP(&filter, "filter", "f", "", "Id of the filter.")
	logsCmd.Flags().StringVar(&filterTyp, "type", "", "Type of the filter.")
}
