package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/flow"
	"github.com/spf13/cobra"
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
		slog.Error("terminating flow (main)", "error", err)
		os.Exit(1)
	}
}

var serverCmd = &cobra.Command{
	Use:  "server",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		circuit := core.NewCircuit(context.Background(), os.Interrupt)

		slog.Info("starting server")
		err := flow.Run(circuit)
		if err != nil {
			slog.Error("initializing", "err", err)
			os.Exit(1)
		}
		slog.Info("server booted successfully")

		// wait until server is done.
		<-circuit.Done()
		slog.Info("terminating server")

		go func() {
			time.Sleep(time.Second * 15)
			slog.Error("ungraceful server termination")
			os.Exit(1)
		}()

		circuit.Wait()
		slog.Info("graceful server termination")
	},
}
