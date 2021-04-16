package api

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func ctxDeadline() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	return context.WithDeadline(ctx, time.Now().Add(GRPCCommandTimeout))
}

func tlsConfig(certDir, certType string, insecure bool) (*tls.Config, error) {

	config := &tls.Config{
		Rand:               rand.Reader,
		InsecureSkipVerify: insecure,
	}

	_, err := os.Stat(certDir)
	if len(certDir) == 0 || os.IsNotExist(err) {
		return config, nil
	}

	keyPath := filepath.Join(certDir, fmt.Sprintf("%s.key", certType))
	certPath := filepath.Join(certDir, fmt.Sprintf("%s.pem", certType))
	caPath := filepath.Join(certDir, "ca.pem")

	if _, err := os.Stat(keyPath); err == nil {
		certs, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return config, err
		}
		config.Certificates = []tls.Certificate{certs}
	}

	if _, err := os.Stat(caPath); err == nil {

		pool := x509.NewCertPool()
		pem, err := ioutil.ReadFile(caPath)
		if err != nil {
			return config, err
		}

		if !pool.AppendCertsFromPEM(pem) {
			err := fmt.Errorf("can not append cert to CA")
			return config, err
		}

		config.ClientAuth = tls.RequireAndVerifyClientCert
		config.ClientCAs = pool
	}

	return config, nil

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

func errResponse(w http.ResponseWriter, code int, err error) {
	e := fmt.Errorf("unknown error")
	c := http.StatusInternalServerError

	if code != 0 {
		c = code
	}

	if err != nil {
		e = err
	}

	w.WriteHeader(c)
	io.Copy(w, strings.NewReader(e.Error()))
}
