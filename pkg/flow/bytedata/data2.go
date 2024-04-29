package bytedata

import (
	"github.com/direktiv/direktiv/pkg/flow/grpc"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	enginerefactor "github.com/direktiv/direktiv/pkg/refactor/engine"
	"github.com/direktiv/direktiv/pkg/refactor/instancestore"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertInstanceToGrpcInstance(instance *enginerefactor.Instance) *grpc.Instance {
	return &grpc.Instance{
		CreatedAt:    timestamppb.New(instance.Instance.CreatedAt),
		UpdatedAt:    timestamppb.New(instance.Instance.UpdatedAt),
		Id:           instance.Instance.ID.String(),
		As:           instance.Instance.WorkflowPath,
		Status:       instance.Instance.Status.String(),
		ErrorCode:    instance.Instance.ErrorCode,
		ErrorMessage: string(instance.Instance.ErrorMessage),
		Invoker:      instance.Instance.Invoker,
	}
}

func ConvertInstancesToGrpcInstances(instances []instancestore.InstanceData) []*grpc.Instance {
	list := make([]*grpc.Instance, 0)
	for idx := range instances {
		instance := &instances[idx]
		list = append(list, &grpc.Instance{
			CreatedAt:    timestamppb.New(instance.CreatedAt),
			UpdatedAt:    timestamppb.New(instance.UpdatedAt),
			Id:           instance.ID.String(),
			As:           instance.WorkflowPath,
			Status:       instance.Status.String(),
			ErrorCode:    instance.ErrorCode,
			ErrorMessage: string(instance.ErrorMessage),
			Invoker:      instance.Invoker,
		})
	}

	return list
}

func ConvertNamespaceToGrpc(item *datastore.Namespace) *grpc.Namespace {
	ns := &grpc.Namespace{
		Name: item.Name,

		CreatedAt: timestamppb.New(item.CreatedAt),
		UpdatedAt: timestamppb.New(item.UpdatedAt),
	}

	return ns
}
