package api

import (
	"context"
	"encoding/json"
<<<<<<< HEAD
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
=======
	"io"
	"net/http"
	"os"
>>>>>>> main
	"strconv"
	"time"
)

func closeVerbose(x io.Closer, log io.Writer) {
	if log == nil {
		log = os.Stdout
	}

	err := x.Close()
	if err != nil {
		log.Write([]byte(err.Error()))
	}
}

const filenameRegexp = `^[^\s\.\,\/\*]*$`

func sanitizeFileName(str string) error {

	pass, err := regexp.MatchString(filenameRegexp, str)
	if err != nil {
		return err
	}

	if !pass {
		return fmt.Errorf("file name contains invalid characters ('.', '..', '*', etc.)")
	}

	return nil
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
	respCode := ConvertGRPCStatusCodeToHTTPCode(eo.Code)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(respCode)
	json.NewEncoder(w).Encode(eo)
}
