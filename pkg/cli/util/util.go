package util

import (
	"context"
	"time"

	"github.com/vorteil/direktiv/pkg/ingress"
	"google.golang.org/grpc"
)

func CreateClient(conn *grpc.ClientConn) (ingress.DirektivIngressClient, context.Context, context.CancelFunc) {
	client := ingress.NewDirektivIngressClient(conn)

	// set context with 3 second timeout
	ctx := context.Background()
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*3))

	cancelConns := func() {
		conn.Close()
		cancel()
	}

	return client, ctx, cancelConns
}
