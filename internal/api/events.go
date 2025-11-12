package api

import (
	"net/http"

	"github.com/direktiv/direktiv/internal/datastore"
	"github.com/go-chi/chi/v5"
)

type eventsController struct {
	store         datastore.Store
	wakeInstance  any
	startWorkflow any
}

func (c *eventsController) mountEventHistoryRouter(r chi.Router) {
	r.Get("/", c.dummy)          // Retrieve a list of events
	r.Get("/subscribe", c.dummy) // Retrieve a event updates via sse
	r.Get("/{eventID}", c.dummy) // Get details of a single event
	r.Post("/replay/{eventID}", c.dummy)
}

func (c *eventsController) mountEventListenerRouter(r chi.Router) {
	r.Get("/", c.dummy)                  // Retrieve a list of event-listeners
	r.Get("/{eventListenerID}", c.dummy) // Get details of a single event-listener
}

func (c *eventsController) mountBroadcast(r chi.Router) {
	r.Post("/", c.dummy)
}

func (c *eventsController) dummy(w http.ResponseWriter, r *http.Request) {
}
