package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

func ErrResponse(w http.ResponseWriter, code int, err error) {
	e := fmt.Errorf("unknown error")
	c := http.StatusInternalServerError

	if code != 0 {
		c = code
	}

	if err != nil {
		e = err
	}

	w.WriteHeader(c)
	io.Copy(w, strings.NewReader(e.Error()))
}
