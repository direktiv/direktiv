package main

import (
	"github.com/direktiv/direktiv/cmd/cli"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/auth"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/inbound"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/outbound"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/target"
)

func main() {
	cli.Run()
}
