package sidecar2

import (
	"log/slog"
	"sync"

	"github.com/direktiv/direktiv/pkg/sidecar2/api"
	"github.com/direktiv/direktiv/pkg/sidecar2/config"
)

func RunApplication() {
	var config config.Config
	var dataMap sync.Map // actionID -> Action
	slog.Debug("starting sidecar APIs")
	api.StartApis(config, &dataMap)
	slog.Debug("started sidecar APIs")
}
