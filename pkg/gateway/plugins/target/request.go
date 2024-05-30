package target

import (
	"io"
	"net/http"
	"os"

	"github.com/direktiv/direktiv/pkg/gateway/plugins"
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

	return resp
}
