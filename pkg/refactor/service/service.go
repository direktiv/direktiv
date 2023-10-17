// nolint
package service

import (
	"errors"
)

var ErrNotFound = errors.New("ErrNotFound")

const (
	httpsProxy = "HTTPS_PROXY"
	httpProxy  = "HTTP_PROXY"
	noProxy    = "NO_PROXY"

	containerUser    = "direktiv-container"
	containerSidecar = "direktiv-sidecar"
)
