package target

import (
	"io"
	"net/http"
	"os"

	"github.com/direktiv/direktiv/pkg/tracing"
)

func doRequest(r *http.Request, method, url string, body io.ReadCloser) (*http.Response, error) {
	client := http.Client{}
	ctx := r.Context()
	req, err := http.NewRequestWithContext(ctx, method, url, body)

	endTrace := tracing.TraceGWHTTPRequest(ctx, req, "direktiv/gateway")
	defer endTrace()
	if err != nil {
		return nil, err
	}

	// add api key if required
	if os.Getenv("DIREKTIV_API_KEY") != "" {
		req.Header.Set("Direktiv-Api-Key", os.Getenv("DIREKTIV_API_KEY"))
	}

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}
