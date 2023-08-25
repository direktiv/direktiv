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
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
)

type RouteManager interface {
	AddExtraRoutes(r *chi.Mux) error
	IsAuthenticated(next http.Handler) http.Handler
}

func (s *Server) addUIRoutes() {

	log.Debug().Msg("adding ui routes")
	filesDir := http.Dir(s.conf.Server.Assets)

	log.Info().Msgf("using ui files in %s", filesDir)

	indexFile, err := os.ReadFile(filepath.Join(s.conf.Server.Assets, "index.html"))
	if err != nil {
		// log.Panic().
	}

	// s.chi.Use(rm.IsAuthenticated)

	s.chi.Group(func(router chi.Router) {
		// router.Use(s.routeManager.IsAuthenticated)
		serverFiles := func(wsf http.ResponseWriter, rsf *http.Request) {

			// rFile := rsf.URL.RequestURI()

			deliver := func(deliverFile []byte) {

				// mimeType := http.DetectContentType(deliverFile)

				wsf.Write(deliverFile)
			}

			rf := rsf.URL.RequestURI()
			if rf == "/" || rf == "" {
				// deliver index
				deliver(indexFile)
				return
			}

			f := filepath.Join(s.conf.Server.Assets, rsf.URL.RequestURI())
			b, err := os.ReadFile(f)
			if errors.Is(err, os.ErrNotExist) {
				log.Debug().Msgf("delivering index for %s", f)
				// deliver index
				deliver(indexFile)
				return
			} else if err != nil {
				// error
			}

			// deliver file
			deliver(b)

			// if
			// read index.html if it is root dir
			// if rFile == "/" || rFile == "" {
			// 	rFile = "index.html"
			// }

			// _, err := rf(rFile)
			// // f := filepath.Join(s.conf.Server.Assets, rFile)
			// // _, err := os.ReadFile(f)

			// if errors.Is(err, os.ErrNotExist) {
			// 	rFile = "index.html"
			// } else if err != nil {
			// 	fmt.Printf("just err %v\n", err)
			// }
			// if _, err := os.Stat("/path/to/whatever"); err == nil {
			// 	// path/to/whatever exists

			//   } else if errors.Is(err, os.ErrNotExist) {
			// 	// path/to/whatever does *not* exist

			//   } else {
			// 	// Schrodinger: file may or may not exist. See err for details.

			// 	// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence

			// fs := http.StripPrefix(pathPrefix, http.FileServer(filesDir))
			// fs.ServeHTTP(wsf, rsf)
		}

		// r.Get("/ui", func(w http.ResponseWriter, r *http.Request) {
		// 	// rctx := chi.RouteContext(r.Context())
		// 	// pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		// 	// filesDir := http.Dir(s.conf.Server.Assets)
		// 	// fs := http.StripPrefix(pathPrefix, http.FileServer(filesDir))
		// 	// fs.ServeHTTP(w, r)
		// 	serverFiles(w, r)
		// })

		// router.Get("/ui", func(w http.ResponseWriter, r *http.Request) {
		// 	fmt.Println("REDIRECT!!!!")
		// 	// rctx := chi.RouteContext(r.Context())
		// 	// pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		// 	// filesDir := http.Dir(s.conf.Server.Assets)
		// 	// fs := http.StripPrefix(pathPrefix, http.FileServer(filesDir))
		// 	// fs.ServeHTTP(w, r)
		// 	// serverFiles(w, r)
		// 	http.Redirect(w, r, "/ui/", http.StatusPermanentRedirect)
		// })

		router.Get("/*", func(w http.ResponseWriter, r *http.Request) {

			// router.Use(middleware.WithValue(routePath, "ui"))

			// r.WithValue("1", "2")

			// rctx := chi.RouteContext(r.Context())
			// pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
			// filesDir := http.Dir(s.conf.Server.Assets)
			// fs := http.StripPrefix(pathPrefix, http.FileServer(filesDir))
			// fs.ServeHTTP(w, r)
			serverFiles(w, r)
		})
	})

	// s.chi.Get("/ui/*", func(w http.ResponseWriter, r *http.Request) {
	// 	rctx := chi.RouteContext(r.Context())
	// 	pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
	// 	filesDir := http.Dir(s.conf.Server.Assets)
	// 	fs := http.StripPrefix(pathPrefix, http.FileServer(filesDir))
	// 	fs.ServeHTTP(w, r)
	// })

	// uiRoutes.Use(s.routeManager.IsAuthenticated())
	// uiRoutes.Static("/", s.conf.Server.Assets)
}

// func (s *Server) addAPIRoutes() {

// 	// apiRoutes := s.gin.Group("/api")
// 	// apiRoutes.Use(s.routeManager.IsAuthenticated())
// 	// apiRoutes.Any("/*apiPath", reverseProxy)

// 	// TODO add all API routes here
// 	// apiRoutes.GET("/dummy", dummy)

// 	// apiRoutesv2 := apiRoutes.Group("/v2")
// 	// apiRoutesv2.GET("/dummy", dummy)

// }

// func dummy(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{
// 		"hello": "it's signin",
// 	})
// }

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
