package sidecar

import (
	"log/slog"
	"sync"

	"github.com/direktiv/direktiv/cmd/sidecar/api"
	"github.com/direktiv/direktiv/cmd/sidecar/config"
)

func main() {
	var config config.Config
	var dataMap sync.Map // actionID -> Action
	slog.Debug("starting sidecar APIs")
	api.StartApis(config, &dataMap)
	slog.Debug("started sidecar APIs")
}
