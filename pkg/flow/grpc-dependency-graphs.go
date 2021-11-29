package flow

import (
	"context"

	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (flow *flow) GlobalDependencyGraph(ctx context.Context, req *emptypb.Empty) (*grpc.DependencyGraphResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	g, err := flow.getCompleteDependencyGraph()
	if err != nil {
		return nil, err
	}

	data := marshal(g)

	var resp grpc.DependencyGraphResponse

	resp.Data = []byte(data)

	return &resp, nil

}

func (flow *flow) NamespacedDependencyGraph(ctx context.Context, req *grpc.NamespacedDependencyGraphRequest) (*grpc.DependencyGraphResponse, error) {

	flow.sugar.Debugf("Handling gRPC request: %s", this())

	g, err := flow.getNamespacedDependencyGraph(req.GetNamespace())
	if err != nil {
		return nil, err
	}

	data := marshal(g)

	var resp grpc.DependencyGraphResponse

	resp.Data = []byte(data)

	return &resp, nil

}
