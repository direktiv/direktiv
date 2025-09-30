package extensions

import (
	"net/http"

	"github.com/direktiv/direktiv/internal/cluster/pubsub"
	"github.com/direktiv/direktiv/internal/core"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// AdditionalSchema for hooking additional sql schema provisioning scripts. This helps build new plugins and
// extensions for Direktiv.
var AdditionalSchema string

var IsEnterprise = false

var Initialize func(db *gorm.DB, bus pubsub.EventBus, config *core.Config) error

var AdditionalAPIRoutes map[string]func(r chi.Router)

var CheckOidcMiddleware func(http.Handler) http.Handler

var CheckAPITokenMiddleware func(http.Handler) http.Handler

var CheckAPIKeyMiddleware func(http.Handler) http.Handler
