package functions

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/errors"
)

func k8sToGRPCError(err error) error {

	if errors.IsNotFound(err) {
		return status.Error(codes.NotFound, "not found")
	}

	if errors.IsAlreadyExists(err) {
		return status.Error(codes.AlreadyExists, "already exists")
	}

	if errors.IsInvalid(err) {
		return status.Error(codes.InvalidArgument, "invalid")
	}

	return err
}
