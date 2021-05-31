package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

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
func CtxDeadline(ctx context.Context) (context.Context, context.CancelFunc) {
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
	eo := GenerateErrObject(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(eo.Code)
	json.NewEncoder(w).Encode(eo)
}
