package main

import (
	_ "github.com/ChannelMeter/iso8601duration"
	_ "github.com/coreos/go-oidc/v3/oidc"
	"github.com/direktiv/direktiv/cmd/cli"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/auth"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/inbound"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/outbound"
	_ "github.com/direktiv/direktiv/pkg/gateway/plugins/target"
	_ "github.com/hashicorp/golang-lru/v2"
)

func main() {
	cli.Run()
}
