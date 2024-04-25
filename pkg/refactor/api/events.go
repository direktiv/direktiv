package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	"github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type eventsController struct {
	store         datastore.Store
	wakeInstance  events.WakeEventsWaiter
	startWorkflow events.WorkflowStart
}

// func (engine *engine) StartWorkflow(ctx context.Context, namespace, path string, input []byte) (*instancestore.InstanceData, error) {

func (c *eventsController) mountEventHistoryRouter(r chi.Router) {
	r.Get("/", c.listEvents)         // Retrieve a list of events
	r.Get("/subscribe", c.subscribe) // Retrieve a event updates via sse
	r.Get("/{eventID}", c.getEvent)  // Get details of a single event
}

func (c *eventsController) mountEventListenerRouter(r chi.Router) {
	r.Get("/", c.listEventListeners)                // Retrieve a list of event-listeners
	r.Get("/{eventListenerID}", c.getEventListener) // Get details of a single event-listener
}

func (c *eventsController) mountBroadcast(r chi.Router) {
	r.Post("/", c.registerCoudEvent)
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
	params := extractEventFilterParams(r)
	data, err := c.store.EventHistory().GetOld(r.Context(), ns.Name, t, params...)
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

func (c *eventsController) subscribe(w http.ResponseWriter, r *http.Request) {
	// cursor is set to multiple seconds before the current time to mitigate data loss
	// that may occur due to delays between submitting and processing the request, or when a sequence of client requests is necessary.
	cursor := time.Now().UTC().Add(-time.Second * 3)

	// Set the appropriate headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Create a channel to send SSE messages
	messageChannel := make(chan Event)
	params := extractEventFilterParams(r)
	var getCursoredStyle sseHandle = func(ctx context.Context, cursorTime time.Time) ([]CoursoredEvent, error) {
		ns := chi.URLParam(r, "namespace")
		if ns == "" {
			return nil, fmt.Errorf("namespace can not be empty")
		}

		events, err := c.store.EventHistory().GetNew(ctx, ns, cursorTime, params...)
		if err != nil {
			return nil, err
		}
		res := make([]CoursoredEvent, len(events))
		for i, e := range events {
			b, err := json.Marshal(e)
			if err != nil {
				return nil, err
			}
			dst := &bytes.Buffer{}
			if err := json.Compact(dst, b); err != nil {
				return nil, err
			}
			res[i] = CoursoredEvent{
				Event: Event{
					ID:   e.Event.ID(),
					Type: "message",
					Data: dst.String(),
				},
				Time: e.ReceivedAt,
			}
		}

		return res, nil
	}

	worker := seeWorker{
		Get:      getCursoredStyle,
		Interval: time.Second,
		Ch:       messageChannel,
		Cursor:   cursor,
	}

	go worker.start(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case message := <-messageChannel:
			_, err := io.Copy(w, strings.NewReader(fmt.Sprintf("id: %v\nevent: %v\ndata: %v\n\n", message.ID, message.Type, message.Data)))
			if err != nil {
				slog.Error("serve to SSE", "err", err)
			}

			f, ok := w.(http.Flusher)
			if !ok {
				return
			}
			if f != nil {
				f.Flush()
			}
		}
	}
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
	data, err := c.store.EventListener().GetOld(r.Context(), ns.Name, t)
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

func (c *eventsController) registerCoudEvent(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)
	cType := r.Header.Get("Content-type")
	limit := int64(1024 * 1024 * 32)

	if r.ContentLength > 0 {
		if r.ContentLength > limit {
			http.Error(w, "request payload too large", http.StatusBadRequest)

			return
		}
	}

	var processor func(data []byte) ([]event.Event, error)
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error parsing CloudEvents batch", http.StatusBadRequest)

		return
	}
	// Check if the content type indicates a batch of CloudEvents
	if strings.HasPrefix(cType, "application/cloudevents-batch+json") {
		processor = extractBatchevent
	}

	// Check if the content type indicates a single CloudEvent
	if strings.HasPrefix(cType, "application/json") {
		s := r.Header.Get("Ce-Type")
		if s == "" {
			// some weird magic for historical reasons...
			r.Header.Set("Content-Type", "application/cloudevents+json; charset=UTF-8")
		}
		processor = extractEvent
	} else {
		// If content type is not recognized, return an error
		http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
	}
	evs, err := processor(b)
	if err != nil {
		http.Error(w, "Error parsing CloudEvent", http.StatusBadRequest)

		return
	}
	engine := events.EventEngine{
		WorkflowStart:       c.startWorkflow,
		WakeInstance:        c.wakeInstance,
		GetListenersByTopic: c.store.EventListenerTopics().GetListeners,
		UpdateListeners:     c.store.EventListener().UpdateOrDelete,
	}

	dEvs := convertEvents(*ns, evs...)
	c.store.EventHistory().Append(r.Context(), dEvs)

	engine.ProcessEvents(r.Context(), ns.ID, evs, func(template string, args ...interface{}) {
		slog.Error(fmt.Sprintf(template, args...))
	})
	// status ok here.
}

func extractBatchevent(data []byte) ([]cloudevents.Event, error) {
	var events []cloudevents.Event

	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("failed parsing CloudEvents batch")
	}

	var err error
	for i, ev := range events {
		events[i], err = validateEvent(ev)
		if err != nil {
			return nil, err
		}
	}

	return events, nil
}

func extractEvent(data []byte) ([]cloudevents.Event, error) {
	ev := cloudevents.NewEvent()
	if err := json.Unmarshal(data, &ev); err != nil {
		return nil, fmt.Errorf("failed parsing CloudEvent")
	}
	ev, err := validateEvent(ev)
	if err != nil {
		return nil, err
	}

	return append([]event.Event{}, ev), nil
}

func validateEvent(event cloudevents.Event) (cloudevents.Event, error) {
	if event.SpecVersion() == "" {
		event.SetSpecVersion("1.0")
	}

	if event.ID() == "" {
		event.SetID(uuid.NewString())
	}
	// NOTE: this validate check added to sanitize Azure's dodgy cloudevents.
	err := event.Validate()

	if err != nil && strings.Contains(err.Error(), "dataschema") {
		event.SetDataSchema("")
		err = event.Validate()
		if err != nil {
			return cloudevents.Event{}, fmt.Errorf("invalid cloudevent: %w", err)
		}
	}
	// NOTE: remarshal / unmarshal necessary to overcome issues with cloudevents library.
	data, err := json.Marshal(event)
	if err != nil {
		return cloudevents.Event{}, fmt.Errorf("invalid cloudevent: %w", err)
	}

	err = event.UnmarshalJSON(data)
	if err != nil {
		return cloudevents.Event{}, fmt.Errorf("invalid cloudevent: %w", err)
	}

	return event, nil
}

func extractEventFilterParams(r *http.Request) []string {
	params := make([]string, 0)
	if v := chi.URLParam(r, "namespace"); v != "" {
		params = append(params, "namespace")
		params = append(params, v)
	}
	if v := chi.URLParam(r, "createdBefore"); v != "" {
		params = append(params, "created_before")
		params = append(params, v)
	}
	if v := chi.URLParam(r, "createdAfter"); v != "" {
		params = append(params, "created_after")
		params = append(params, v)
	}
	if v := chi.URLParam(r, "receivedBefore"); v != "" {
		params = append(params, "received_before")
		params = append(params, v)
	}
	if v := chi.URLParam(r, "receivedAfter"); v != "" {
		params = append(params, "received_after")
		params = append(params, v)
	}
	if v := chi.URLParam(r, "eventContains"); v != "" {
		params = append(params, "event_contains")
		params = append(params, v)
	}
	if v := chi.URLParam(r, "typeContains"); v != "" {
		params = append(params, "type_contains")
		params = append(params, v)
	}

	return params
}

func convertEvents(ns datastore.Namespace, evs ...cloudevents.Event) []*events.Event {
	res := make([]*events.Event, len(evs))
	for i := range evs {
		res[i] = &events.Event{
			Event:         &evs[i],
			NamespaceName: ns.Name,
			Namespace:     ns.ID,
			ReceivedAt:    time.Now().UTC(),
		}
	}

	return res
}
