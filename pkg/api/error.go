package api

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

// Errors GRPC
func convertGRPCStatusCodeToHTTPCode(code codes.Code) int {
	switch code {
	case codes.Canceled:
		return 400
	case codes.Unknown:
		return 400
	case codes.InvalidArgument:
		return 406
	case codes.DeadlineExceeded:
		return 400
	case codes.NotFound:
		return 400
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return 400
	case codes.ResourceExhausted:
		return 400
	case codes.FailedPrecondition:
		return 400
	case codes.Aborted:
		return 400
	case codes.OutOfRange:
		return 400
	case codes.Unimplemented:
		return 400
	case codes.Internal:
		return 400
	case codes.Unavailable:
		return 400
	case codes.DataLoss:
		return 400
	case codes.Unauthenticated:
		return 400
	}

	return 500
}
