package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
)

type RouteManager interface {
	// IsAuthenticated() gin.HandlerFunc
	AddExtraRoutes(r *chi.Mux) error

	IsAuthenticated(next http.Handler) http.Handler
}

var (
	routePath = "path"
)

func (s *Server) addUIRoutes() {
	log.Debug().Msg("adding ui routes")
	// uiRoutes := s.gin.Group("/ui")

	// router.Route("/users", func(r chi.Router) {
	filesDir := http.Dir(s.conf.Server.Assets)

	s.chi.Group(func(router chi.Router) {
		router.Use(s.routeManager.IsAuthenticated)
		// r.Use(middleware.RedirectSlashes)

		serverFiles := func(wsf http.ResponseWriter, rsf *http.Request) {
			rctx := chi.RouteContext(rsf.Context())
			pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
			fs := http.StripPrefix(pathPrefix, http.FileServer(filesDir))
			fs.ServeHTTP(wsf, rsf)
		}

		// r.Get("/ui", func(w http.ResponseWriter, r *http.Request) {
		// 	// rctx := chi.RouteContext(r.Context())
		// 	// pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		// 	// filesDir := http.Dir(s.conf.Server.Assets)
		// 	// fs := http.StripPrefix(pathPrefix, http.FileServer(filesDir))
		// 	// fs.ServeHTTP(w, r)
		// 	serverFiles(w, r)
		// })

		router.Get("/ui", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("REDIRECT!!!!")
			// rctx := chi.RouteContext(r.Context())
			// pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
			// filesDir := http.Dir(s.conf.Server.Assets)
			// fs := http.StripPrefix(pathPrefix, http.FileServer(filesDir))
			// fs.ServeHTTP(w, r)
			// serverFiles(w, r)
			http.Redirect(w, r, "/ui/", http.StatusPermanentRedirect)
		})

		router.Get("/ui/*", func(w http.ResponseWriter, r *http.Request) {

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

func (s *Server) addAPIRoutes() {

	// apiRoutes := s.gin.Group("/api")
	// apiRoutes.Use(s.routeManager.IsAuthenticated())
	// apiRoutes.Any("/*apiPath", reverseProxy)

	// TODO add all API routes here
	// apiRoutes.GET("/dummy", dummy)

	// apiRoutesv2 := apiRoutes.Group("/v2")
	// apiRoutesv2.GET("/dummy", dummy)

}

// func dummy(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{
// 		"hello": "it's signin",
// 	})
// }

func reverseProxy(c *gin.Context) {

	target, _ := url.Parse("https://jensgwork.local")

	apiPath := c.Param("apiPath")
	fmt.Printf("REDIRECT %v", apiPath)

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

	// proxy.Director = func(req *http.Request) {
	// 	req.Header = c.Request.Header
	// 	req.Host = target.Host
	// 	req.URL.Scheme = target.Scheme
	// 	req.URL.Host = rc.backendURL.Host
	// 	// apiPrefix = "/api"
	// 	req.URL.Path = fmt.Sprintf("%s/%s", "/api", c.Param("apiPath"))
	// }

	proxy.ServeHTTP(c.Writer, c.Request)

}
