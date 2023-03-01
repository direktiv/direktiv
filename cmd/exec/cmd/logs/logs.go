package logs

import (
	root "github.com/direktiv/direktiv/cmd/exec/cmd"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Prints logs for a instance",
	Run: func(cmd *cobra.Command, args []string) {
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
	// outputFlag     string
	// execNoPushFlag bool
	filterTyp string
	filter    string
	instance  string
)

func init() {
	root.RootCmd.AddCommand(logsCmd)
	logsCmd.Flags().StringVarP(&instance, "instance id", "i", "", "Id of the instance for which to grab the logs.")
	logsCmd.Flags().StringVarP(&filter, "filter id", "f", "", "Id of the filter.")
	logsCmd.Flags().StringVar(&filterTyp, "type", "", "Type of the filter.")
}
