package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/vorteil/direktiv/pkg/util"
)

var errNamespaceRegex = fmt.Errorf("namespace name must match the regex pattern `%s`", util.RegexPattern)
var errWorkflowRegex = fmt.Errorf("workflow id must match the regex pattern `%s`", util.RegexPattern)
var errSecretRegex = fmt.Errorf("secret key must match the regex pattern `%s`", util.VarRegexPattern)

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
