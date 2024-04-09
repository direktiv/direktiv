package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/flow"
	"github.com/direktiv/direktiv/pkg/refactor/core"
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
		appCtx, appCancel := signal.NotifyContext(context.Background(), os.Interrupt)

		circuit := &core.Circuit{
			Context: appCtx,
			Cancel:  appCancel,
			WG:      sync.WaitGroup{},
		}

		slog.Info("starting server")
		err := flow.Run(circuit)
		if err != nil {
			slog.Error("initializing", "err", err)
			os.Exit(1)
		}
		slog.Info("server booted successfully")

		// wait until server is done.
		<-appCtx.Done()
		slog.Info("terminating server")

		go func() {
			time.After(time.Second * 5)
			slog.Error("ungraceful server termination")
			os.Exit(1)
		}()

		circuit.WG.Wait()
		slog.Info("graceful server termination")
	},
}
