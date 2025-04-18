package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/events"
	"github.com/direktiv/direktiv/pkg/telemetry"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type eventsController struct {
	store         datastore.Store
	wakeInstance  events.WakeEventsWaiter
	startWorkflow events.WorkflowStart
}

func (c *eventsController) mountEventHistoryRouter(r chi.Router) {
	r.Get("/", c.listEvents)         // Retrieve a list of events
	r.Get("/subscribe", c.subscribe) // Retrieve a event updates via sse
	r.Get("/{eventID}", c.getEvent)  // Get details of a single event
	r.Post("/replay/{eventID}", c.replay)
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
	starting := time.Now().UTC().Format(time.RFC3339Nano)
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
	params, err := extractEventFilterParams(r)
	if err != nil {
		writeBadrequestError(w, err)

		return
	}

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
		writeJSONWithMeta(w, []*datastore.Event{}, metaInfo)

		return
	}
	var previousPage interface{} = data[0].ReceivedAt.UTC().Format(time.RFC3339Nano)

	if len(data) > 0 {
		previousPage = data[len(data)-1].ReceivedAt.UTC().Format(time.RFC3339Nano)
	}

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

func (c *eventsController) replay(w http.ResponseWriter, r *http.Request) {
	ctx := telemetry.GetContextFromRequest(r.Context(), r)
	ctx, span := telemetry.Tracer.Start(ctx, "event-replay")
	defer span.End()

	eventID := chi.URLParam(r, "eventID")
	if eventID == "" {
		writeBadrequestError(w, fmt.Errorf("no event id provided"))
		return
	}
	ns := extractContextNamespace(r)

	d, err := c.store.EventHistory().GetByID(r.Context(), eventID)
	if err != nil {
		writeInternalError(w, err)

		return
	}
	processEvents(ctx, c, ns, *d.Event)
}

func (c *eventsController) subscribe(w http.ResponseWriter, r *http.Request) {
	// cursor is set to multiple seconds before the current time to mitigate data loss
	// that may occur due to delays between submitting and processing the request, or when a sequence of client requests is necessary.
	cursor := time.Now().UTC().Add(-time.Second * 3)

	// Set the appropriate headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	params, err := extractEventFilterParams(r)
	if err != nil {
		writeInternalError(w, err)
	}

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Create a channel to send SSE messages
	messageChannel := make(chan Event)
	var getCursoredStyle sseHandle = func(ctx context.Context, cursorTime time.Time) ([]CoursoredEvent, error) {
		return sseHandlefunc(ctx, r, c, cursorTime, params)
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
	res := convertListenersForAPI(d)

	writeJSON(w, res)
}

func (c *eventsController) listEventListeners(w http.ResponseWriter, r *http.Request) {
	ns := extractContextNamespace(r)
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		writeBadrequestError(w, err)

		return
	}
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		writeBadrequestError(w, err)

		return
	}
	data, count, err := c.store.EventListener().Get(r.Context(), ns.ID, limit, offset)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	metaInfo := map[string]any{
		"total": count,
	}
	if len(data) == 0 {
		writeJSONWithMeta(w, []*datastore.Event{}, metaInfo)

		return
	}
	res := make([]eventListenerEntry, len(data))
	for i := range data {
		l := convertListenersForAPI(data[i])
		res[i] = l
	}
	slices.Reverse(res)

	writeJSONWithMeta(w, res, metaInfo)
}

func convertListenersForAPI(listener *datastore.EventListener) eventListenerEntry {
	e := eventListenerEntry{
		ID:                     listener.ID.String(),
		CreatedAt:              listener.CreatedAt,
		UpdatedAt:              listener.UpdatedAt,
		Namespace:              listener.Namespace,
		ListeningForEventTypes: listener.ListeningForEventTypes,
	}
	if len(listener.EventContextFilters) != 0 {
		e.EventContextFilters = listener.EventContextFilters
	}
	if len(listener.ReceivedEventsForAndTrigger) != 0 {
		e.ReceivedEventsForAndTrigger = listener.ReceivedEventsForAndTrigger
	}
	if len(listener.TriggerInstance) != 0 {
		e.TriggerInstance = listener.TriggerInstance
	}
	if len(listener.TriggerWorkflow) != 0 {
		e.TriggerWorkflow = listener.Metadata
	}
	e.TriggerType = fmt.Sprint(listener.TriggerType)

	return e
}

// nolint:canonicalheader
func (c *eventsController) registerCoudEvent(w http.ResponseWriter, r *http.Request) {
	ctx := telemetry.GetContextFromRequest(r.Context(), r)
	ctx, span := telemetry.Tracer.Start(ctx, "event-request")
	defer span.End()

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
	} else if strings.HasPrefix(cType, "application/json") {
		processor = extractEvent
	} else if strings.HasPrefix(cType, "application/cloudevents+json") {
		processor = extractEvent
	} else {
		// If content type is not recognized, return an error
		http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)

		return
	}
	evs, err := processor(b)
	if err != nil {
		http.Error(w, "Error parsing CloudEvent", http.StatusBadRequest)

		return
	}
	for _, ev := range evs {
		cEv := convertEvents(*ns, ev)
		_, errs := c.store.EventHistory().Append(r.Context(), cEv)
		for _, e := range errs {
			if e != nil {
				slog.Error("Failed storing CloudEvent to the event history", "error", e)
			}
		}
		for _, e := range errs {
			if e != nil && errors.Is(e, datastore.ErrDuplication) {
				http.Error(w, "Error appending CloudEvent to history", http.StatusBadRequest)

				return
			}
		}
		for _, e := range errs {
			if e != nil {
				http.Error(w, "Internal error while appending CloudEvent to history", http.StatusInternalServerError)

				return
			}
		}

		processEvents(ctx, c, ns, ev)
		// status ok here.
	}
}

func processEvents(ctx context.Context, c *eventsController, ns *datastore.Namespace, ev event.Event) {
	engine := events.EventEngine{
		WorkflowStart:       c.startWorkflow,
		WakeInstance:        c.wakeInstance,
		GetListenersByTopic: c.store.EventListenerTopics().GetListeners,
		UpdateListeners:     c.store.EventListener().UpdateOrDelete,
	}
	engine.ProcessEvents(ctx, ns.ID, []event.Event{ev}, func(template string, args ...interface{}) {
		slog.Error(fmt.Sprintf(template, args...))
	})
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

func extractEventFilterParams(r *http.Request) ([]string, error) {
	params := make([]string, 0)
	if v := chi.URLParam(r, "namespace"); v != "" {
		params = append(params, "namespace")
		params = append(params, v)
	}
	if v := r.URL.Query().Get("createdBefore"); v != "" {
		params = append(params, "created_before")
		t, err := parseTime(v)
		if err != nil {
			return nil, err
		}
		params = append(params, t)
	}
	if v := r.URL.Query().Get("createdAfter"); v != "" {
		params = append(params, "created_after")
		t, err := parseTime(v)
		if err != nil {
			return nil, err
		}
		params = append(params, t)
	}
	if v := r.URL.Query().Get("receivedBefore"); v != "" {
		params = append(params, "received_before")
		t, err := parseTime(v)
		if err != nil {
			return nil, err
		}
		params = append(params, t)
	}
	if v := r.URL.Query().Get("receivedAfter"); v != "" {
		params = append(params, "received_after")
		t, err := parseTime(v)
		if err != nil {
			return nil, err
		}
		params = append(params, t)
	}
	if v := r.URL.Query().Get("eventContains"); v != "" {
		params = append(params, "event_contains")
		params = append(params, v)
	}
	if v := r.URL.Query().Get("typeContains"); v != "" {
		params = append(params, "type_contains")
		params = append(params, v)
	}

	return params, nil
}

func parseTime(t string) (string, error) {
	e, err := time.Parse(time.RFC3339Nano, t)
	if err != nil {
		return "", err
	}

	return e.UTC().Format(time.RFC3339Nano), nil
}

func convertEvents(ns datastore.Namespace, evs ...cloudevents.Event) []*datastore.Event {
	res := make([]*datastore.Event, len(evs))
	for i := range evs {
		res[i] = &datastore.Event{
			Event:       &evs[i],
			Namespace:   ns.Name,
			NamespaceID: ns.ID,
			ReceivedAt:  time.Now().UTC(),
		}
	}

	return res
}

type eventListenerEntry struct {
	ID                          string    `json:"id,omitempty"`
	CreatedAt                   time.Time `json:"createdAt"`
	UpdatedAt                   time.Time `json:"updatedAt"`
	Namespace                   string    `json:"namespace"`
	ListeningForEventTypes      []string  `json:"listeningForEventTypes,omitempty"`
	ReceivedEventsForAndTrigger any       `json:"receivedEventsForAndTrigger,omitempty"`
	LifespanOfReceivedEvents    any       `json:"lifespanOfReceivedEvents,omitempty"`
	TriggerType                 string    `json:"triggerType"`
	TriggerWorkflow             any       `json:"triggerWorkflow,omitempty"`
	TriggerInstance             any       `json:"triggerInstance,omitempty"`
	EventContextFilters         any       `json:"eventContextFilters,omitempty"`
}

// nolint:canonicalheader
func sseHandlefunc(ctx context.Context, r *http.Request, c *eventsController, cursorTime time.Time, params []string) ([]CoursoredEvent, error) {
	ns := chi.URLParam(r, "namespace")
	if ns == "" {
		return nil, fmt.Errorf("namespace can not be empty")
	}
	events := make([]*datastore.Event, 0)
	var err error
	if lastID := r.Header.Get("Last-Event-ID"); lastID != "" {
		id, err := strconv.Atoi(lastID)
		if err != nil {
			return nil, err
		}
		lostEvents, err := c.store.EventHistory().GetStartingIDUntilTime(ctx, ns, id, cursorTime, params...)
		if err != nil {
			return nil, err
		}
		events = append(events, lostEvents...)
	}
	newEvents, err := c.store.EventHistory().GetNew(ctx, ns, cursorTime, params...)
	events = append(events, newEvents...)
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
