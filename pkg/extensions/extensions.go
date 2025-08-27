package extensions

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/go-chi/chi/v5"
)

// AdditionalSchema for hooking additional sql schema provisioning scripts. This helps build new plugins and
// extensions for Direktiv.
var AdditionalSchema string

var IsEnterprise = false

var Initialize func(db *database.DB, bus core.PubSub, config *core.Config) error

var AdditionalAPIRoutes map[string]func(r chi.Router)

var CheckOidcMiddleware func(http.Handler) http.Handler

var CheckAPITokenMiddleware func(http.Handler) http.Handler

var CheckAPIKeyMiddleware func(http.Handler) http.Handler
