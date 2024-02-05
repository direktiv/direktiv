package dinit

import (
	"io"
	"io/fs"
	"os"
)

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
