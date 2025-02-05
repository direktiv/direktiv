package extensions

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/go-chi/chi/v5"
)

// AdditionalSchema for hooking additional sql schema provisioning scripts. This helps build new plugins and
// extensions for Direktiv.
var AdditionalSchema string

var IsEnterprise = false

var Initialize func(db *database.DB, bus *pubsub.Bus, config *core.Config)

var AdditionalAPIRoutes map[string]func(r chi.Router)

var CheckOidcMiddlewares func(http.Handler) http.Handler
