package api

import (
	"net/http"
	"slices"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type eventsController struct {
	store datastore.Store
}

func (c *eventsController) mountEventHistoryRouter(r chi.Router) {
	r.Get("/", c.listEvents)        // Retrieve a list of events
	r.Get("/{eventID}", c.getEvent) // Get details of a single event
}

func (c *eventsController) mountEventListenerRouter(r chi.Router) {
	r.Get("/", c.listEventListeners)                // Retrieve a list of event-listeners
	r.Get("/{eventListenerID}", c.getEventListener) // Get details of a single event-listener
}

func (c *eventsController) listEvents(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)
	starting := ""
	if v := r.URL.Query().Get("before"); v != "" {
		starting = v
	}
	t := time.Now().UTC()
	if starting != "" {
		co, err := time.Parse(time.RFC3339Nano, starting)
		if err != nil {
			writeInternalError(w, err)

			return
		}
		t = co
	}

	data, err := c.store.EventHistory().GetOld(r.Context(), ns.Name, t)
	if err != nil {
		writeInternalError(w, err)

		return
	}

	metaInfo := map[string]any{
		"previousPage": nil, // setting them to nil make ensure matching the specicied types for the clients
		"startingFrom": t,
	}

	if len(data) == 0 {
		writeJSONWithMeta(w, []logEntry{}, metaInfo)

		return
	}

	slices.Reverse(data)
	var previousPage interface{} = data[0].ReceivedAt.UTC().Format(time.RFC3339Nano)

	metaInfo = map[string]any{
		"previousPage": previousPage,
		"startingFrom": starting,
	}

	writeJSONWithMeta(w, data, metaInfo)
}

func (c *eventsController) getEvent(w http.ResponseWriter, r *http.Request) {
	eventID := ""
	if v := chi.URLParam(r, "eventID"); v != "" {
		eventID = v
	}
	d, err := c.store.EventHistory().GetByID(r.Context(), eventID)
	if err != nil {
		writeInternalError(w, err)

		return
	}
	writeJSON(w, d)
}

func (c *eventsController) getEventListener(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "eventListenerID")

	id, err := uuid.Parse(eventID)
	if err != nil {
		writeInternalError(w, err)

		return
	}
	d, err := c.store.EventListener().GetByID(r.Context(), id)
	if err != nil {
		writeInternalError(w, err)

		return
	}
	writeJSON(w, d)
}

func (c *eventsController) listEventListeners(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)
	starting := r.URL.Query().Get("before")

	t := time.Now().UTC()
	if starting != "" {
		co, err := time.Parse(time.RFC3339Nano, starting)
		if err != nil {
			writeInternalError(w, err)

			return
		}
		t = co
	}
	data, err := c.store.EventListener().GetNew(r.Context(), ns.Name, t)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	metaInfo := map[string]any{
		"previousPage": nil, // setting them to nil make ensure matching the specicied types for the clients
		"startingFrom": nil,
	}
	if len(data) == 0 {
		writeJSONWithMeta(w, []logEntry{}, metaInfo)

		return
	}

	slices.Reverse(data)
	var previousPage interface{} = data[0].CreatedAt.UTC().Format(time.RFC3339Nano)

	metaInfo = map[string]any{
		"previousPage": previousPage,
		"startingFrom": starting,
	}

	writeJSONWithMeta(w, data, metaInfo)
}
