package api

import (
	"encoding/json"
	"net/http"
)

// var errNamespaceRegex = fmt.Errorf("namespace name must match the regex pattern `%s`", util.RegexPattern)
// var errWorkflowRegex = fmt.Errorf("workflow id must match the regex pattern `%s`", util.RegexPattern)
// var errSecretRegex = fmt.Errorf("secret key must match the regex pattern `%s`", util.VarRegexPattern)
//
// const filenameRegexp = `^[^\s\.\,\/\*]*$`
//
// func sanitizeFileName(str string) error {
//
// 	pass, err := regexp.MatchString(filenameRegexp, str)
// 	if err != nil {
// 		return err
// 	}
//
// 	if !pass {
// 		return fmt.Errorf("file name contains invalid characters ('.', '..', '*', etc.)")
// 	}
//
// 	return nil
// }
//
// func writeData(resp interface{}, w http.ResponseWriter) {
// 	// Write Data
// 	retData, err := json.Marshal(resp)
// 	if err != nil {
// 		ErrResponse(w, err)
// 		return
// 	}
//
// 	w.WriteHeader(http.StatusOK)
// 	/* #nosec */
// 	_, _ = w.Write(retData)
// }
//
// // CtxDeadline defines default request deadline
// func CtxDeadline(ctx context.Context) (context.Context, context.CancelFunc) {
// 	return context.WithDeadline(ctx, time.Now().Add(GRPCCommandTimeout))
// }
//
// func paginationParams(r *http.Request) (offset, limit int) {
// 	if x, ok := r.URL.Query()["offset"]; ok && len(x) > 0 {
// 		offset, _ = strconv.Atoi(x[0])
// 	}
// 	if x, ok := r.URL.Query()["limit"]; ok && len(x) > 0 {
// 		limit, _ = strconv.Atoi(x[0])
// 	}
// 	return
// }
//
// ErrResponse creates error based on grpc error
func ErrResponse(w http.ResponseWriter, err error) {
	eo := GenerateErrObject(err)
	respCode := ConvertGRPCStatusCodeToHTTPCode(eo.Code)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(respCode)
	/* #nosec */
	_ = json.NewEncoder(w).Encode(eo)
}

//
// // SSE Util functions
//
// func ErrSSEResponse(w http.ResponseWriter, flusher http.Flusher, err error) {
// 	eo := GenerateErrObject(err)
//
// 	b, err := json.Marshal(eo)
// 	if err != nil {
// 		logger.Errorf("FAILED to marshal sse error: %v", eo)
// 	}
//
// 	_, err = w.Write([]byte(fmt.Sprintf("event: error\ndata: %s\n\n", string(b))))
// 	if err != nil {
// 		logger.Errorf("FAILED to write sse error: %s", string(b))
// 	}
//
// 	flusher.Flush()
// }
//
// func ErrSSEResponseSimple(w http.ResponseWriter, flusher http.Flusher, data []byte) {
// 	_, err := w.Write([]byte(fmt.Sprintf("event: error\ndata: %s\n\n", string(data))))
// 	if err != nil {
// 		logger.Errorf("FAILED to write sse error: %s", string(data))
// 	}
//
// 	flusher.Flush()
// }
//
// func SetupSEEWriter(w http.ResponseWriter) (http.Flusher, error) {
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
// 	w.Header().Set("Content-Type", "text/event-stream")
// 	w.Header().Set("Cache-Control", "no-cache")
// 	w.Header().Set("Connection", "keep-alive")
//
// 	flusher, ok := w.(http.Flusher)
// 	if !ok {
// 		return flusher, fmt.Errorf("streaming unsupported")
// 	}
//
// 	return flusher, nil
// }
//
// func WriteSSEJSONData(w http.ResponseWriter, flusher http.Flusher, data interface{}) error {
// 	b, err := json.Marshal(data)
// 	if err != nil {
// 		err = fmt.Errorf("client recieved bad data: %w", err)
// 		logger.Error(err)
// 		return err
// 	}
//
// 	return WriteSSEData(w, flusher, b)
// }
//
// func WriteSSEData(w http.ResponseWriter, flusher http.Flusher, data []byte) error {
// 	_, err := w.Write([]byte(fmt.Sprintf("data: %s\n\n", string(data))))
// 	if err != nil {
// 		err = fmt.Errorf("client failed to write data: %w", err)
// 		logger.Error(err)
// 		return err
// 	}
//
// 	flusher.Flush()
// 	return nil
// }
//
// func SendSSEHeartbeat(w http.ResponseWriter, flusher http.Flusher) {
// 	_, err := w.Write([]byte(fmt.Sprintf("data: %s\n\n", "")))
// 	if err != nil {
// 		ErrSSEResponse(w, flusher, fmt.Errorf("client failed to write hearbeat: %w", err))
// 		logger.Error(fmt.Errorf("client failed to write hearbeat: %w", err))
// 	}
//
// 	flusher.Flush()
// }
