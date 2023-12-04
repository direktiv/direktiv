package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
)

type RouteManager interface {
	AddExtraRoutes(r *chi.Mux) error
}

const (
	uiPrefix = "/ui"
)

func (s *Server) addUIRoutes() {

	log.Debug().Msg("adding ui routes")
	filesDir := http.Dir(s.conf.Server.Assets)

	log.Info().Msgf("using ui files in %s", filesDir)

	indexFile, err := os.ReadFile(filepath.Join(s.conf.Server.Assets, "index.html"))
	if err != nil {
		log.Fatal().Err(err).Msg("can not read index file")
	}

	s.chi.Group(func(router chi.Router) {
		serverFiles := func(wsf http.ResponseWriter, rsf *http.Request) {

			deliver := func(deliverFile []byte) {

				mimeType := http.DetectContentType(deliverFile)

				ext := filepath.Ext(rsf.URL.RequestURI())
				switch ext {
				case ".css":
					mimeType = "text/css"
				case ".js":
					mimeType = "text/javascript"
				}

				wsf.Header().Set("Content-Type", mimeType)
				wsf.Write(deliverFile)
			}

			rf := rsf.URL.RequestURI()
			if rf == uiPrefix || rf == uiPrefix+"/" {
				deliver(indexFile)
				return
			}

			ss := strings.Replace(rsf.URL.RequestURI(), "/ui/", "", 1)

			f := filepath.Join(s.conf.Server.Assets, ss)
			b, err := os.ReadFile(f)
			if errors.Is(err, os.ErrNotExist) {
				log.Debug().Msgf("delivering index for %s", f)
				deliver(indexFile)
				return
			} else if err != nil {
				SendError(wsf, err, http.StatusInternalServerError,
					ErrorServerInternal, "file delivery failed")
			}

			// deliver file, it exists
			deliver(b)
		}

		router.Get("/ui", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/ui/", http.StatusPermanentRedirect)
		})

		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/ui/", http.StatusPermanentRedirect)
		})

		router.Get("/ui/*", func(w http.ResponseWriter, r *http.Request) {
			serverFiles(w, r)
		})

		router.Get("/ui/assets/*", func(w http.ResponseWriter, r *http.Request) {
			serverFiles(w, r)
		})

	})

}

func ReverseProxy(r *http.Request, w http.ResponseWriter, urlTarget string) {

	target, err := url.Parse(urlTarget)
	if err != nil {
		SendError(w, err, http.StatusInternalServerError, ErrorServerInternal,
			fmt.Sprintf("can not parse url %s", urlTarget))
		return
	}

	log.Debug().Msgf("proxy url %s", urlTarget)

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.FlushInterval = 100 * time.Millisecond
	var InsecureTransport http.RoundTripper = &http.Transport{
		// TLSClientConfig: &tls.Config{InsecureSkipVerify: rc.Conf.Backend.Insecure},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   24 * time.Hour,
			KeepAlive: 24 * time.Hour,
		}).Dial,
		TLSHandshakeTimeout: 60 * time.Second,
	}
	proxy.Transport = InsecureTransport

	proxy.Director = func(req *http.Request) {
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}

	proxy.ServeHTTP(w, r)
}
