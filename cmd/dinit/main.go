// Package dinit provides a simple mechanism to prepare containers for action without
// a server listening to port 8080. This enables Direktiv to use standard containers from
// e.g. DockerHub.
package dinit

import (
	"io"
	"io/fs"
	"os"
)

// RunApplication runs if the Direktiv binary starts up as `init`. This happens when a action
// uses the special command `/usr/share/direktiv/direktiv-cmd` in the configuration. It copies
// the direktiv-cmd binary to a shared location of the function container and the container is
// able to use this as a server to execute commands.
func RunApplication() {
	perm := 0o755

	// copies the command server to the shared directory in kubernetes
	err := os.MkdirAll("/usr/share/direktiv", fs.FileMode(perm))
	if err != nil {
		panic(err)
	}

	source, err := os.Open("/bin/direktiv-cmd")
	if err != nil {
		panic(err)
	}

	destination, err := os.Create("/usr/share/direktiv/direktiv-cmd")
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(destination, source)
	if err != nil {
		panic(err)
	}

	err = os.Chmod("/usr/share/direktiv/direktiv-cmd", fs.FileMode(perm))
	if err != nil {
		panic(err)
	}

	destination.Close()
	source.Close()
}
