package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/direktiv/direktiv/pkg/flow"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/status"
)

func runApplication() {
	var err error
	var addr string

	rootCmd := &cobra.Command{
		Use: "flow",
	}

	rootCmd.PersistentFlags().StringVar(&addr, "addr", "localhost:8080", "")
	rootCmd.AddCommand(serverCmd)

	err = rootCmd.Execute()
	if err != nil {
		desc := status.Convert(err)
		slog.Error("terminating flow (main)", "status", desc, "error", err)
		os.Exit(1)
	}
}

var serverCmd = &cobra.Command{
	Use:  "server",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		serverCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer func() {
			cancel()
			slog.Info("graceful shutdown initiated.")
		}()
		slog.Info("starting server")
		err := flow.Run(serverCtx)
		if err != nil {
			slog.Error("server termination due to error", "error", err)
			os.Exit(1)
		}
	},
}
