package flow

import (
	"context"
	"net"
)

type internal struct {
	*server
	listener net.Listener
}

func initInternalServer(ctx context.Context, srv *server) (*internal, error) {
	var err error

	internal := &internal{server: srv}

	internal.listener, err = net.Listen("tcp", ":7777") //nolint:gosec
	if err != nil {
		return nil, err
	}

	return internal, nil
}
