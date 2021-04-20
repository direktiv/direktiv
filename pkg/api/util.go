package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type errObject struct {
	Code    int
	Message string
}

func CtxDeadline() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	return context.WithDeadline(ctx, time.Now().Add(GRPCCommandTimeout))
}

func paginationParams(r *http.Request) (offset, limit int) {
	if x, ok := r.URL.Query()["offset"]; ok && len(x) > 0 {
		offset, _ = strconv.Atoi(x[0])
	}
	if x, ok := r.URL.Query()["limit"]; ok && len(x) > 0 {
		limit, _ = strconv.Atoi(x[0])
	}
	return
}

// ErrResponse creates error based on grpc error
func ErrResponse(w http.ResponseWriter, code int, err error) {

	st, ok := status.FromError(err)
	eo := &errObject{
		Code:    999,
		Message: err.Error(),
	}
	if ok {
		eo = &errObject{
			Code:    int(st.Code()),
			Message: st.Message(),
		}
	}

	respCode := 500
	switch eo.Code {
	case int(codes.NotFound):
		{
			respCode = 404
		}
	case int(codes.AlreadyExists):
		{
			respCode = 409
		}
	}

	w.WriteHeader(respCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eo)

}
