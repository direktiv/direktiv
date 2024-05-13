package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/direktiv/direktiv/pkg/jqer"
	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
)

type jxController struct{}

func (c *jxController) mountRouter(r chi.Router) {
	r.Post("/", c.handler)
}

type jxRequest struct {
	JX   []byte `json:"jx"`
	Data []byte `json:"data"`
}

type jxResponse struct {
	JX     []byte   `json:"jx"`
	Data   []byte   `json:"data"`
	Output [][]byte `json:"output"`
	Logs   []byte   `json:"logs"`
}

func (c *jxController) handler(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	var req jxRequest

	err = json.Unmarshal(data, &req)
	if err != nil {
		writeNotJSONError(w, err)

		return
	}

	var query interface{}
	var document interface{}

	err = json.Unmarshal(req.Data, &document)
	if err != nil {
		err = fmt.Errorf("invalid 'data': %w", err)
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: err.Error(),
		})

		return
	}

	err = yaml.Unmarshal(req.JX, &query)
	if err != nil {
		err = fmt.Errorf("invalid 'jx': %w", err)
		writeError(w, &Error{
			Code:    "request_data_invalid",
			Message: err.Error(),
		})

		return
	}

	resp := &jxResponse{
		JX:     req.JX,
		Data:   req.Data,
		Output: make([][]byte, 0),
	}

	buf := new(bytes.Buffer)

	var failed bool
	var firstResult interface{}

	results, err := jqer.Evaluate(document, query) //nolint:contextcheck
	if err != nil {
		buf.WriteString(fmt.Sprintf("failure: %s\n", err.Error()))
		failed = true
	}

	if len(results) > 0 {
		firstResult = results[0]
	}

	for i := range results {
		result, _ := json.MarshalIndent(results[i], "", "  ")
		resp.Output = append(resp.Output, result)
	}

	resp.Logs = buf.Bytes()
	if resp.Logs == nil {
		resp.Logs = make([]byte, 0)
	}

	asserts := r.URL.Query()["assert"]

	for _, assert := range asserts {
		switch assert {
		case "success":
			if failed {
				writeJxError(w, resp, &Error{
					Code:    "assert_success",
					Message: err.Error(),
				})

				return
			}
		case "array":
			_, ok := firstResult.([]interface{})
			if !ok {
				writeJxError(w, resp, &Error{
					Code:    "assert_array",
					Message: "result is not an array",
				})

				return
			}

		case "object":
			_, ok := firstResult.(map[string]interface{})
			if !ok {
				writeJxError(w, resp, &Error{
					Code:    "assert_object",
					Message: "result is not an object",
				})

				return
			}

		default:
			writeJxError(w, resp, &Error{
				Code:    "request_data_invalid",
				Message: fmt.Sprintf("unknown assert: %s", assert),
			})

			return
		}
	}

	writeJSON(w, resp)
}

func writeJxError(w http.ResponseWriter, resp *jxResponse, err *Error) {
	httpStatus := http.StatusBadRequest

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	payLoad := struct {
		Error *Error      `json:"error"`
		Data  *jxResponse `json:"data"`
	}{
		Error: err,
		Data:  resp,
	}

	_ = json.NewEncoder(w).Encode(payLoad)
}
