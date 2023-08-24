package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "github.com/gin-contrib/sessions"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Server struct {
	conf *Config
	// gin  *gin.Engine
	chi *chi.Mux
	srv *http.Server

	routeManager RouteManager
}

var (
	EnvAPIKey = "DIREKTIV_APIKEY"
)

const (
	APITokenHeader = "direktiv-token"
)

func NewServer(conf *Config, rm RouteManager) *Server {

	r := chi.NewRouter()
	r.Use(LoggerMiddleware(&log.Logger))

	log.Info().Msgf("listening to %s", conf.Server.Listen)

	srv := &http.Server{
		Addr:    conf.Server.Listen,
		Handler: r,
	}

	s := &Server{
		conf:         conf,
		chi:          r,
		srv:          srv,
		routeManager: rm,
	}

	s.addUIRoutes()
	s.addAPIRoutes()

	err := rm.AddExtraRoutes(r)
	if err != nil {
		log.Panic().Msgf("could not add extra routes: %s", err.Error())
	}

	chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Debug().Msgf("%s [%s]", route, method)
		return nil
	})

	return s
}

func LoggerMiddleware(logger *zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			log := logger.With().Logger()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				t2 := time.Now()
				// if rec := recover(); rec != nil {
				// 	log.Error().
				// 		Str("type", "error").
				// 		Timestamp().
				// 		Interface("recover_info", rec).
				// 		Bytes("debug_stack", debug.Stack()).
				// 		Msg("log system error")
				// 	http.Error(ww, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				// }

				log.Debug().
					// Timestamp().
					Fields(map[string]interface{}{
						"remote_ip":  r.RemoteAddr,
						"url":        r.URL.Path,
						"proto":      r.Proto,
						"method":     r.Method,
						"status":     ww.Status(),
						"latency_ms": float64(t2.Sub(t1).Nanoseconds()) / 1000000.0,
						"bytes_in":   r.Header.Get("Content-Length"),
						"bytes_out":  ww.BytesWritten(),
					}).
					Msg("request")
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

// func HandlerDummyAPI() gin.HandlerFunc {

// 	jens := "JENS"

// 	return func(ctx *gin.Context) {
// 		ctx.Writer.Write([]byte("DUMMY!!!"))
// 		ctx.Writer.Write([]byte(jens))
// 	}
// }

// func (s *Server) SetRouteManager(rm RouteManager) error {

// 	// gob.Register(map[string]interface{}{})
// 	// gob.Register(oauth2.Token{})
// 	// gob.Register(User{})

// 	// var store = sessions.NewCookieStore([]byte(s.conf.OIDC.CookieSecret))
// 	// store.Op

// 	// s.conf.OIDC.
// 	// store := cookie.NewStore([]byte(rc.Conf.Server.CookieSecret))
// 	// r.Use(sessions.SessionsMany([]string{OAuthTokenSession,
// 	// 	TokenSession, ProfileSession}, store))

// 	// store := cookie.NewStore([]byte("secret"))
// 	// store.Options(sessions.Options{MaxAge: 60 * 60 * 24}) // expire in a day
// 	// s.gin.Use(sessions.Sessions("mysession", store))

// 	// add ui routes

// 	// add api routes

// 	// apiRoutes := s.gin.Group("/api")
// 	// apiRoutes.Any("/:apiPath", rc.reverseProxy)
// 	// apiRoutes.Any("/dummy", HandlerDummyAPI())

// 	// s.routeManager = rm
// 	// err := rm.AddExtraRoutes(s.gin, s.conf)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	// return nil
// }

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		log.Fatal().
			Err(err).
			Msgf("could not shutdown server")
	}
}

func (s *Server) Start() {

	if s.routeManager == nil {
		log.Panic().Msg("route manager not set")
	}

	go func() {

		var err error

		if len(s.conf.Server.TLSCert) > 0 {
			log.Info().Msg("starting with TLS")
			err = s.srv.ListenAndServeTLS(s.conf.Server.TLSCert, s.conf.Server.TLSKey)
		} else {
			log.Info().Msg("starting without TLS")
			err = s.srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatal().
				Err(err).
				Msgf("could not start server")
		}

	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Info().Msg("shutting down server")
	s.Stop()

}
