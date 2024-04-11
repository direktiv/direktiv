package target

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/util"
)

func doRequest(w http.ResponseWriter, r *http.Request, method, url string, body io.ReadCloser) *http.Response {

	client := http.Client{}
	ctx := r.Context()
	req, err := http.NewRequestWithContext(ctx, method, url, body)

	endTrace := util.TraceGWHTTPRequest(ctx, req, "direktiv/flow")
	defer endTrace()
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"can not create request", err)

		return nil
	}

	// add api key if required
	if os.Getenv("DIREKTIV_API_KEY") != "" {
		req.Header.Set("Direktiv-Token", os.Getenv("DIREKTIV_API_KEY"))
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		plugins.ReportError(r.Context(), w, http.StatusInternalServerError,
			"can not execute flow", err)

		return nil
	}

	// error handling
	errorCode := resp.Header.Get("Direktiv-Instance-Error-Code")
	errorMessage := resp.Header.Get("Direktiv-Instance-Error-Message")
	instance := resp.Header.Get("Direktiv-Instance-Id")

	if errorCode != "" {
		msg := fmt.Sprintf("%s: %s (%s)", errorCode, errorMessage, instance)
		plugins.ReportError(r.Context(), w, resp.StatusCode,
			"error executing workflow", fmt.Errorf(msg))

		return nil
	}

	// direktiv requests always respond with 200, workflow errors are handled in the previous check
	if resp.StatusCode >= http.StatusMultipleChoices {
		plugins.ReportError(r.Context(), w, resp.StatusCode,
			"can not execute flow", fmt.Errorf(resp.Status))

		return nil
	}

	return resp
}
