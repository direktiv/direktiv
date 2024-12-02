package main

import (
	"context"
	"io"
	"io/fs"
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
	rootCmd.AddCommand(serverCmd, dinitCmd)

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

// command dinitCmd provides a simple mechanism to prepare containers for action without
// a server listening to port 8080. This enables Direktiv to use standard containers from
// e.g. DockerHub.
var dinitCmd = &cobra.Command{
	Use:  "dinit",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("starting dinit")

		perm := 0o755
		sharedDir := "/usr/share/direktiv"
		cmdBinary := "/app/direktiv-cmd"
		targetBinary := "/usr/share/direktiv/direktiv-cmd"

		slog.Info("starting RunApplication", "sharedDir", sharedDir, "cmdBinary", cmdBinary)

		// Ensure the shared directory exists
		slog.Info("creating shared directory", "path", sharedDir)
		err := os.MkdirAll(sharedDir, fs.FileMode(perm))
		if err != nil {
			slog.Error("failed to create shared directory", "path", sharedDir, "error", err)
			panic(err)
		}

		// Open source file
		slog.Info("opening source binary", "path", cmdBinary)
		source, err := os.Open(cmdBinary)
		if err != nil {
			slog.Error("failed to open source binary", "path", cmdBinary, "error", err)
			panic(err)
		}
		defer source.Close()

		// Create destination file
		slog.Info("creating destination binary", "path", targetBinary)
		destination, err := os.Create(targetBinary)
		if err != nil {
			slog.Error("failed to create destination binary", "path", targetBinary, "error", err)
			panic(err)
		}
		defer destination.Close()

		// Copy the source binary to the destination
		slog.Info("copying binary", "source", cmdBinary, "destination", targetBinary)
		_, err = io.Copy(destination, source)
		if err != nil {
			slog.Error("failed to copy binary", "source", cmdBinary, "destination", targetBinary, "error", err)
			panic(err)
		}

		// Change permissions of the target binary
		slog.Info("setting permissions", "path", targetBinary, "permissions", perm)
		err = os.Chmod(targetBinary, fs.FileMode(perm))
		if err != nil {
			slog.Error("failed to set permissions", "path", targetBinary, "permissions", perm, "error", err)
			panic(err)
		}

		slog.Info("RunApplication completed successfully")
	},
}
