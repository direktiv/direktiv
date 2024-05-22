package api

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// GenericErrorCode - Reserved status code for generic non grpc errors.
	GenericErrorCode codes.Code = 50
)

// ErrObject for grpc.
type ErrObject struct {
	Code    codes.Code
	Message string
}

var grpcErrorHTTPCodeMap = map[codes.Code]int{
	codes.Canceled:           http.StatusBadRequest,
	codes.Unknown:            http.StatusInternalServerError,
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
	GenericErrorCode:         http.StatusInternalServerError,
}

// ConvertGRPCStatusCodeToHTTPCode - Convert Grpc Code errors to http response codes.
func ConvertGRPCStatusCodeToHTTPCode(code codes.Code) int {
	if val, ok := grpcErrorHTTPCodeMap[code]; ok {
		return val
	}

	return http.StatusInternalServerError
}

// GenerateErrObject - Unwrap grpc errors into ErrorObject.
func GenerateErrObject(err error) *ErrObject {
	eo := new(ErrObject)
	if st, ok := status.FromError(err); ok {
		eo.Code = st.Code()
		eo.Message = st.Message()
	} else {
		eo.Code = GenericErrorCode
		eo.Message = err.Error()
	}

	// // Handle Certain Erros
	// if eo.isRegexError() {
	// 	eo.Message = strings.Replace(eo.Message, `must match regex: ^[a-z][a-z0-9._-]{1,34}[a-z0-9]$`, humanErrorInvalidRegex, 1)
	// }

	return eo
}

// func (e *ErrObject) isRegexError() (ok bool) {
// 	if  e.Code !=   codes.InvalidArgument {
// 		ok =  false
// 	} else if strings.HasSuffix(e.Message, `^[a-z][a-z0-9._-]{1,34}[a-z0-9]$`) {
// 		ok = true
// 	}
//
// 	return ok
// }
