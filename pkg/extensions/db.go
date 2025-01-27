package extensions

import (
	eeDStore "github.com/direktiv/direktiv/cmd/ee/datastore"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/go-chi/chi/v5"
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
	DB       *database.DB
	Bus      *pubsub.Bus
	EEDStore eeDStore.Store
}
