package main

import (
	"os"
	"strings"

	"github.com/direktiv/direktiv/pkg/cmdserver"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/auth"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/inbound"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/outbound"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/target"
)

func main() {
	if strings.Contains(os.Args[0], "direktiv-cmd") {
		cmdserver.Start()

		return
	}

	runApplication()
}
