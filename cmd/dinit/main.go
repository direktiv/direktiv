// Package dinit provides a simple mechanism to prepare containers for action without
// a server listening to port 8080. This enables Direktiv to use standard containers from
// e.g. DockerHub.
package dinit

import (
	"io"
	"io/fs"
	"log/slog"
	"os"
)

// RunApplication runs if the Direktiv binary starts up as `init`. This happens when an action
// uses the special command `/usr/share/direktiv/direktiv-cmd` in the configuration. It copies
// the direktiv-cmd binary to a shared location of the function container and the container is
// able to use this as a server to execute commands.
func RunApplication() {
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
}
