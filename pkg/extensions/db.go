package extensions

import (
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// AdditionalSchema for hooking additional sql schema provisioning scripts. This helps build new plugins and
// extensions for Direktiv.
var AdditionalSchema func() string

var AdditionalAPIRoutes map[string]RouteController

type RouteController interface {
	Initialize(app App)
	MountRouter(r chi.Router)
}

type App struct {
	DB  *database.DB
	Bus *pubsub.Bus
}

var IsEnterprise = false

var AdditionalMiddlewares interface {
	CheckOidc(http.Handler) http.Handler
}
