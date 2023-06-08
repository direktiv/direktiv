package flow

import (
	"context"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/version"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) Build(ctx context.Context, in *emptypb.Empty) (*grpc.BuildResponse, error) {
	var resp grpc.BuildResponse
	resp.Build = version.Version
	return &resp, nil
}
