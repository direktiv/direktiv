package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/util"
)

var errNamespaceRegex = fmt.Errorf("namespace name must match the regex pattern `%s`", util.RegexPattern)
var errWorkflowRegex = fmt.Errorf("workflow id must match the regex pattern `%s`", util.RegexPattern)
var errSecretRegex = fmt.Errorf("secret key must match the regex pattern `%s`", util.VarRegexPattern)

func closeVerbose(x io.Closer, log io.Writer) {
	if log == nil {
		log = os.Stdout
	}

	err := x.Close()
	if err != nil {
		/* #nosec */
		_, _ = log.Write([]byte(err.Error()))
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
	/* #nosec */
	_, _ = w.Write(retData)
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
	/* #nosec */
	_ = json.NewEncoder(w).Encode(eo)
}

func ErrSSEResponse(w http.ResponseWriter, flusher http.Flusher, err error) {
	eo := GenerateErrObject(err)

	b, err := json.Marshal(eo)
	if err != nil {
		log.Errorf("FAILED to marshal sse error: %v", eo)
	}

	_, err = w.Write([]byte(fmt.Sprintf("event: error\ndata: %s\n\n", string(b))))
	if err != nil {
		log.Errorf("FAILED to write sse error: %s", string(b))
	}

	flusher.Flush()
}

func setupSEEWriter(w http.ResponseWriter) (http.Flusher, error) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return flusher, fmt.Errorf("streaming unsupported")
	}

	return flusher, nil
}
