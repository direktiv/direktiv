package core

import (
	"github.com/direktiv/direktiv/pkg/refactor/service"
)

type Version struct {
	UnixTime int64 `json:"unix_time"`
}

type App struct {
	Version          *Version
	FunctionsManager *service.Manager
}
