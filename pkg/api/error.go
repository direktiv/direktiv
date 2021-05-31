package api

import (
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	humanErrorInvalidRegex string = "must be less than 36 characters and may only use lowercase letters, numbers, and “-_”"
)

// ErrObject for grpc
type ErrObject struct {
	Code     int
	Message  string
	grpcCode codes.Code
}

var grpcErrorHttpCodeMap = map[codes.Code]int{
	codes.Canceled:           http.StatusBadRequest,
	codes.Unknown:            http.StatusBadRequest,
	codes.InvalidArgument:    http.StatusNotAcceptable,
	codes.DeadlineExceeded:   http.StatusBadRequest,
	codes.NotFound:           http.StatusNotFound,
	codes.AlreadyExists:      http.StatusConflict,
	codes.PermissionDenied:   http.StatusBadRequest,
	codes.ResourceExhausted:  http.StatusBadRequest,
	codes.FailedPrecondition: http.StatusBadRequest,
	codes.Aborted:            http.StatusBadRequest,
	codes.OutOfRange:         http.StatusBadRequest,
	codes.Unimplemented:      http.StatusBadRequest,
	codes.Internal:           http.StatusBadRequest,
	codes.Unavailable:        http.StatusBadRequest,
	codes.DataLoss:           http.StatusBadRequest,
	codes.Unauthenticated:    http.StatusBadRequest,
}

// convertGRPCStatusCodeToHTTPCode - Errors GRPC
func convertGRPCStatusCodeToHTTPCode(code codes.Code) int {
	if val, ok := grpcErrorHttpCodeMap[code]; ok {
		return val
	}

	return http.StatusInternalServerError
}

func GenerateErrObject(err error) *ErrObject {
	eo := new(ErrObject)
	if st, ok := status.FromError(err); ok {
		eo.Code = convertGRPCStatusCodeToHTTPCode(st.Code())
		eo.Message = st.Message()
		eo.grpcCode = st.Code()
	} else {
		eo.Code = 999
		eo.Message = err.Error()
	}

	// Handle Certain Erros
	if eo.isRegexError() {
		eo.Message = strings.Replace(eo.Message, `must match regex: ^[a-z][a-z0-9._-]{1,34}[a-z0-9]$`, humanErrorInvalidRegex, 1)
	}

	return eo
}

func (e *ErrObject) GrpcCode() codes.Code {
	return e.grpcCode
}

func (e *ErrObject) isRegexError() (ok bool) {
	if e.grpcCode != codes.InvalidArgument {
		ok = false
	} else if strings.HasSuffix(e.Message, `^[a-z][a-z0-9._-]{1,34}[a-z0-9]$`) {
		ok = true
	}

	return ok
}
