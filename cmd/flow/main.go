package flow

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"strings"

	"github.com/direktiv/direktiv/pkg/flow"
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/spf13/cobra"
	libgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var ctx = context.Background()

var (
	addr string

	stream                             bool
	limit, offset                      int32
	orderField, orderDirection         string
	filterField, filterType, filterVal string

	stdin  bool
	filein string
)

func RunApplication() {
	var err error

	rootCmd.PersistentFlags().StringVar(&addr, "addr", "localhost:8080", "")
	rootCmd.AddCommand(serverCmd)

	rootCmd.AddCommand(serverLogsCmd)
	rootCmd.AddCommand(namespaceLogsCmd)
	rootCmd.AddCommand(workflowLogsCmd)
	rootCmd.AddCommand(instanceLogsCmd)

	rootCmd.AddCommand(eventListenersCmd)
	rootCmd.AddCommand(eventHistoryCmd)
	rootCmd.AddCommand(eventReplayCmd)

	rootCmd.AddCommand(namespaceCmd)
	rootCmd.AddCommand(namespacesCmd)
	rootCmd.AddCommand(createNamespaceCmd)
	rootCmd.AddCommand(deleteNamespaceCmd)

	rootCmd.AddCommand(directoryCmd)
	rootCmd.AddCommand(createDirectoryCmd)
	rootCmd.AddCommand(deleteNodeCmd)
	rootCmd.AddCommand(renameNodeCmd)
	rootCmd.AddCommand(nodeCmd)

	rootCmd.AddCommand(workflowCmd)
	rootCmd.AddCommand(createWorkflowCmd)
	rootCmd.AddCommand(updateWorkflowCmd)

	rootCmd.AddCommand(startWorkflowCmd)
	rootCmd.AddCommand(instanceCmd)
	rootCmd.AddCommand(instancesCmd)
	rootCmd.AddCommand(instanceInputCmd)
	rootCmd.AddCommand(instanceOutputCmd)

	rootCmd.AddCommand(secretsCmd)
	rootCmd.AddCommand(setSecretCmd)
	rootCmd.AddCommand(deleteSecretCmd)

	err = rootCmd.Execute()
	if err != nil {
		exit(err)
	}
}

func addPaginationFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&stream, "stream", false, "")
	cmd.Flags().Int32Var(&limit, "limit", -1, "")
	cmd.Flags().Int32Var(&offset, "offset", 0, "")
	cmd.Flags().StringVar(&orderField, "order.field", "", "")
	cmd.Flags().StringVar(&orderDirection, "order.direction", "", "")
	cmd.Flags().StringVar(&filterField, "filter.field", "", "")
	cmd.Flags().StringVar(&filterType, "filter.type", "", "")
	cmd.Flags().StringVar(&filterVal, "filter.val", "", "")
}

func client() (grpc.FlowClient, io.Closer, error) {
	conn, err := libgrpc.Dial(addr, libgrpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	return grpc.NewFlowClient(conn), conn, nil
}

// Todo evaluate if we can remove this.
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
	desc := status.Convert(err)

	slog.Error("terminating", "status", desc)

	os.Exit(1)
}

var rootCmd = &cobra.Command{
	Use: "flow",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
}

var serverCmd = &cobra.Command{
	Use:  "server",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: yassir: need to be cleaned.
		defer func() {
			if r := recover(); r != nil {
				slog.Info("Recovered in run", "run", r, "stack_trace", string(debug.Stack()))
				panic(r)
			}
		}()
		defer shutdown()

		serverCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		err := flow.Run(serverCtx)
		if err != nil {
			slog.Error("terminating", "error", err)
			os.Exit(1)
		}
	},
}

func shutdown() {
	// just in case, stop DNS server
	pv, err := os.ReadFile("/proc/version")
	if err == nil {
		// this is a direktiv machine, so we press poweroff
		if strings.Contains(string(pv), "#direktiv") {
			log.Printf("direktiv machine, powering off")

			if err := exec.Command("/sbin/poweroff").Run(); err != nil {
				slog.Error("error shutting down", "error", err)
			}
		}
	}
}
