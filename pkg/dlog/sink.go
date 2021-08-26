package dlog

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

// HTTPWriter interface implementation
type HTTPWriter struct {
	conn net.Conn
	url  *url.URL

	client *http.Client
}

// NewHTTPSink write logs to http / fluentbit
func NewHTTPSink(url *url.URL) (zap.Sink, error) {

	tw := &HTTPWriter{
		url: url,
	}

	tr := &http.Transport{
		MaxIdleConnsPerHost: 1024,
	}
	tw.client = &http.Client{Transport: tr}

	return tw, nil
}

// Close interface implementation
func (tw *HTTPWriter) Close() error {
	return nil
}

// Write interface implementation
func (tw *HTTPWriter) Write(p []byte) (int, error) {

	resp, err := tw.client.Post(tw.url.String(), "application/json", bytes.NewBuffer(p))
	if err != nil {
		log.Println("err", err)
		return 0, err
	}

	io.Copy(ioutil.Discard, resp.Body) // <= NOTE
	defer resp.Body.Close()

	return len(p), err
}

// Sync interface implementation
func (tw *HTTPWriter) Sync() error {
	return nil
}
