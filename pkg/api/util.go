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

// ErrObject for grpc
type ErrObject struct {
	Code    int
	Message string
}

func writeData(resp interface{}, w http.ResponseWriter) {
	// Write Data
	retData, err := json.Marshal(resp)
	if err != nil {
		ErrResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(retData)
}

// CtxDeadline defines default request deadline
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
func ErrResponse(w http.ResponseWriter, err error) {

	st, ok := status.FromError(err)
	eo := &ErrObject{
		Code:    999,
		Message: err.Error(),
	}
	if ok {
		eo = &ErrObject{
			Code:    int(st.Code()),
			Message: st.Message(),
		}
	}

	respCode := 400
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
