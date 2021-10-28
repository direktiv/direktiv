package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/direktiv/direktiv/pkg/dlog"
	"github.com/direktiv/direktiv/pkg/flow"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/util"
	"go.uber.org/zap"
	libgrpc "google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var ctx = context.Background()

var (
	addr string

	stream                             bool
	after, before                      string
	first, last                        int32
	orderField, orderDirection         string
	filterField, filterType, filterVal string

	stdin  bool
	filein string
)

var (
	logger *zap.SugaredLogger
)

func main() {

	var err error

	logger, err = dlog.ApplicationLogger("flow")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	rootCmd.PersistentFlags().StringVar(&addr, "addr", "localhost:8080", "")
	rootCmd.AddCommand(serverCmd)

	rootCmd.AddCommand(serverLogsCmd)
	rootCmd.AddCommand(namespaceLogsCmd)
	rootCmd.AddCommand(workflowLogsCmd)
	rootCmd.AddCommand(instanceLogsCmd)

	rootCmd.AddCommand(namespaceCmd)
	rootCmd.AddCommand(namespacesCmd)
	rootCmd.AddCommand(createNamespaceCmd)
	rootCmd.AddCommand(deleteNamespaceCmd)
	rootCmd.AddCommand(renameNamespaceCmd)

	rootCmd.AddCommand(directoryCmd)
	rootCmd.AddCommand(createDirectoryCmd)
	rootCmd.AddCommand(deleteNodeCmd)
	rootCmd.AddCommand(renameNodeCmd)
	rootCmd.AddCommand(nodeCmd)

	rootCmd.AddCommand(workflowCmd)
	rootCmd.AddCommand(createWorkflowCmd)
	rootCmd.AddCommand(updateWorkflowCmd)
	rootCmd.AddCommand(saveHeadCmd)
	rootCmd.AddCommand(discardHeadCmd)
	rootCmd.AddCommand(tagsCmd)
	rootCmd.AddCommand(refsCmd)
	rootCmd.AddCommand(revisionsCmd)

	rootCmd.AddCommand(startWorkflowCmd)
	rootCmd.AddCommand(instanceCmd)
	rootCmd.AddCommand(instancesCmd)
	rootCmd.AddCommand(instanceInputCmd)
	rootCmd.AddCommand(instanceOutputCmd)

	rootCmd.AddCommand(routerCmd)
	rootCmd.AddCommand(editRouterCmd)
	rootCmd.AddCommand(secretsCmd)
	rootCmd.AddCommand(setSecretCmd)
	rootCmd.AddCommand(deleteSecretCmd)

	rootCmd.AddCommand(testsCmd)
	testsCmd.Flags().BoolVarP(&skipLongTests, "quick", "q", false, "")
	testsCmd.Flags().IntVarP(&parallelTests, "clients", "c", 1, "")
	testsCmd.Flags().DurationVarP(&instanceTimeout, "instance-timeout", "t", time.Second*5, "")
	testsCmd.Flags().DurationVarP(&testTimeout, "test-timeout", "T", time.Second*10, "")

	err = rootCmd.Execute()
	if err != nil {
		exit(err)
	}

}

func addPaginationFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&stream, "stream", false, "")
	cmd.Flags().StringVar(&after, "after", "", "")
	cmd.Flags().Int32Var(&first, "first", 0, "")
	cmd.Flags().StringVar(&before, "before", "", "")
	cmd.Flags().Int32Var(&last, "last", 0, "")
	cmd.Flags().StringVar(&orderField, "order.field", "", "")
	cmd.Flags().StringVar(&orderDirection, "order.direction", "", "")
	cmd.Flags().StringVar(&filterField, "filter.field", "", "")
	cmd.Flags().StringVar(&filterType, "filter.type", "", "")
	cmd.Flags().StringVar(&filterVal, "filter.val", "", "")
}

func client() (grpc.FlowClient, io.Closer, error) {

	conn, err := libgrpc.Dial(addr, libgrpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}

	return grpc.NewFlowClient(conn), conn, nil

}

func print(x interface{}) {

	data, err := protojson.MarshalOptions{
		Multiline:       true,
		EmitUnpopulated: true,
	}.Marshal(x.(proto.Message))
	if err != nil {
		exit(err)
	}

	s := string(data)

	fmt.Fprintf(os.Stdout, "%s\n", s)

}

func exit(err error) {

	desc := libgrpc.ErrorDesc(err)

	logger.Error(fmt.Sprintf("%s", desc))

	os.Exit(1)

}

var rootCmd = &cobra.Command{
	Use: "flow",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

	},
}

var serverCmd = &cobra.Command{
	Use:  "server CONFIG_FILE",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		defer shutdown()

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		conf, err := util.ReadConfig(args[0])
		if err != nil {
			exit(err)
		}

		err = flow.Run(ctx, logger, conf)
		if err != nil {
			exit(err)
		}

	},
}

func shutdown() {

	// just in case, stop DNS server
	pv, err := ioutil.ReadFile("/proc/version")
	if err == nil {

		// this is a direktiv machine, so we press poweroff
		if strings.Contains(string(pv), "#direktiv") {

			log.Printf("direktiv machine, powering off")

			if err := exec.Command("/sbin/poweroff").Run(); err != nil {
				fmt.Println("error shutting down:", err)
			}

		}

	}

}
