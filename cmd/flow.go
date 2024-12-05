// nolint:forbidigo
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/cmdserver"
	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/flow"
	"github.com/direktiv/direktiv/pkg/sidecar"
	"github.com/spf13/cobra"
)

func runApplication() {
	startCmd := &cobra.Command{
		Use:   "start SERVICE_NAME",
		Short: "Starts the specified direktiv service",
		Long: `The "start" command starts a direktiv service. 
You need to specify the SERVICE_NAME as an argument.`,
	}

	eventCmd := &cobra.Command{
		Use:   "event COMMAND",
		Short: "Executes a direktiv event command",
	}
	eventCmd.AddCommand(eventSendCmd)

	instancesCmd := &cobra.Command{
		Use:   "filesystem",
		Short: "Execute flows and push files",
	}
	instancesCmd.AddCommand(instancesPushCmd)
	instancesCmd.AddCommand(instancesExecCmd)
	instancesExecCmd.PersistentFlags().Bool("push", true, "Push before execute.")

	startCmd.AddCommand(startAPICmd, startSidecarCmd, startDinitCmd, startCommandServerCmd)

	rootCmd := &cobra.Command{
		Use:   "direktiv",
		Short: "This CLI is for lunching Direktiv stacks and interacting its APIs",
		Args:  cobra.ExactArgs(1),
	}

	rootCmd.AddCommand(startCmd, eventCmd)

	err := rootCmd.Execute()
	if err != nil {
		slog.Error("terminating (main)", "error", err)
		os.Exit(1)
	}
}

var startAPICmd = &cobra.Command{
	Use:   "api",
	Short: "direktiv API service",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("starting 'api' service...")

		circuit := core.NewCircuit(context.Background(), os.Interrupt)

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

// command startDinitCmd provides a simple mechanism to prepare containers for action without
// a server listening to port 8080. This enables Direktiv to use standard containers from
// e.g. DockerHub.
var startDinitCmd = &cobra.Command{
	Use:   "dinit",
	Short: "a helper service for direktiv sidecar",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("starting 'dinit' service...")

		perm := 0o755
		sharedDir := "/usr/share/direktiv"
		cmdBinary := "/app/direktiv"
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

var startSidecarCmd = &cobra.Command{
	Use:   "sidecar",
	Short: "direktiv sidecar service, this service manage action request to user containers",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("starting 'sidecar' service...")
		sidecar.RunApplication(context.Background())
	},
}

var startCommandServerCmd = &cobra.Command{
	Use:   "cmdserver",
	Short: "direktiv cmdserver service, this service is part of direktiv sidecar stack",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("starting 'cmdserver' service...")
		cmdserver.Start()
	},
}

var eventSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Sends a file as cloudevent to Direktiv",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := prepareCommand()

		uploader, err := newUploader("", p)
		if err != nil {
			return err
		}

		b, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}

		url := fmt.Sprintf("%s/api/v2/namespaces/%s/events/broadcast", p.Address, p.Namespace)
		resp, err := uploader.sendRequest("POST", url, b)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			var errJSON errorResponse
			err = json.Unmarshal(b, &errJSON)
			if err != nil {
				return err
			}

			return fmt.Errorf(errJSON.Error.Message)
		}

		fmt.Println("event sent")

		return nil
	},
}
