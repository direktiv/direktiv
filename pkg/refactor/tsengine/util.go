package tsengine

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	direktivErrorCodeHeader    = "Direktiv-ErrorCode"
	direktivErrorMessageHeader = "Direktiv-ErrorMessage"

	direktivErrorInternal = "io.direktiv.internal"
)

type errorStruct struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

func writeError(w http.ResponseWriter, code, msg string) {
	w.Header().Add(direktivErrorCodeHeader, code)
	w.Header().Add(direktivErrorMessageHeader, msg)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(500)
	e := errorStruct{
		Code: code,
		Msg:  msg,
	}
	json.NewEncoder(w).Encode(&e)

}

func CreateMultiPartForm(prefix string, flow, flowPath string, secrets map[string]string,
	files map[string]io.Reader) (io.Reader, *multipart.Writer, chan error) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	errCh := make(chan error)

	go func() {
		// adding flow to request
		flowPartName := fmt.Sprintf("%s-%s", prefix, flowPath)
		writer.WriteField(flowPartName, flow)

		for k, v := range files {
			partName := fmt.Sprintf("%s-file_%s", prefix, k)
			part, err := writer.CreateFormFile(partName, k)
			if err != nil {
				errCh <- err
			}

			_, err = io.Copy(part, v)
			if err != nil {
				errCh <- err
			}
		}

		for k, v := range secrets {
			writer.WriteField(fmt.Sprintf("%s-secret_%s", prefix, k), v)
		}

		err := writer.Close()
		if err != nil {
			errCh <- err
		}

		errCh <- nil
	}()

	return pr, writer, errCh
}

func copyFile(src, dst string) (int64, error) {
	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()

	return io.Copy(destination, source)
}
