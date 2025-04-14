package target

import (
	"io"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func doRequest(r *http.Request, method, url string, body io.ReadCloser) (*http.Response, error) {
	// use the otelhttp object
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	ctx := r.Context()
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	// add api key if required
	if os.Getenv("DIREKTIV_API_KEY") != "" {
		req.Header.Set("Direktiv-Api-Key", os.Getenv("DIREKTIV_API_KEY"))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
