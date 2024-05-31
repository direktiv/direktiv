package flow

import (
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	pkgevents "github.com/direktiv/direktiv/pkg/events"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"go.opentelemetry.io/otel/trace"
)

//nolint:gochecknoinits
func init() {
	gob.Register(new(event.EventContextV1))
	gob.Register(new(event.EventContextV03))
}

type events struct {
	*server
	appendStagingEvent func(ctx context.Context, events ...*datastore.StagingEvent) ([]*datastore.StagingEvent, []error)
}

func initEvents(srv *server, appendStagingEvent func(ctx context.Context, events ...*datastore.StagingEvent) ([]*datastore.StagingEvent, []error)) *events {
	events := new(events)
	events.server = srv
	events.appendStagingEvent = appendStagingEvent

	return events
}

func (events *events) handleEvent(ctx context.Context, ns uuid.UUID, nsName string, ce *cloudevents.Event) error {
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID()
	spanID := span.SpanContext().SpanID()
	slog := *slog.With("trace", traceID, "span", spanID, "namespace", nsName)
	e := pkgevents.EventEngine{
		WorkflowStart: func(workflowID uuid.UUID, ev ...*cloudevents.Event) {
			slog.Debug("Starting workflow via CloudEvent.")
			_, end := traceMessageTrigger(ctx, "wf: "+workflowID.String())
			defer end()
			events.engine.EventsInvoke(workflowID, ev...) //nolint:contextcheck
		},
		WakeInstance: func(instanceID uuid.UUID, ev []*cloudevents.Event) {
			slog.Debug("invoking instance via cloudevent", "instance", instanceID)
			_, end := traceMessageTrigger(ctx, "ins: "+instanceID.String())
			defer end()
			events.engine.WakeEventsWaiter(instanceID, ev) //nolint:contextcheck
		},
		GetListenersByTopic: func(ctx context.Context, s string) ([]*datastore.EventListener, error) {
			ctx, end := traceGetListenersByTopic(ctx, s)
			defer end()
			res := make([]*datastore.EventListener, 0)
			err := events.runSQLTx(ctx, func(tx *database.SQLStore) error {
				r, err := tx.DataStore().EventListenerTopics().GetListeners(ctx, s)
				if err != nil {
					slog.Error("Error fetching event-listener-topics.", "error", err)
					return err
				}
				res = r

				return nil
			})
			if err != nil {
				return nil, err
			}

			return res, nil
		},
		UpdateListeners: func(ctx context.Context, listener []*datastore.EventListener) []error {
			slog.Debug("Updating listeners starting.")
			err := events.runSQLTx(ctx, func(tx *database.SQLStore) error {
				errs := tx.DataStore().EventListener().UpdateOrDelete(ctx, listener)
				for _, err2 := range errs {
					if err2 != nil {
						slog.Debug("Error updating listeners.", "error", err2)

						return err2
					}
				}

				return nil
			})
			if err != nil {
				return []error{fmt.Errorf("%w", err)}
			}
			slog.Debug("Updating listeners complete.")

			return nil
		},
	}
	ctx, end := traceProcessingMessage(ctx)
	defer end()

	e.ProcessEvents(ctx, ns, []event.Event{*ce}, func(template string, args ...interface{}) {
		slog.Error(fmt.Sprintf(template, args...))
	})

	return nil
}

func (events *events) BroadcastCloudevent(ctx context.Context, ns *datastore.Namespace, event *cloudevents.Event, timer int64) error {
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID()
	spanID := span.SpanContext().SpanID()

	slog.With("trace", traceID, "span", spanID, "namespace", "event", event.ID(), "event_type", event.Type(), "event_source", event.Source())

	ctx, end := traceBrokerMessage(ctx, *event)
	defer end()

	err := events.addEvent(ctx, event, ns)
	if err != nil {
		return err
	}

	// handle event
	if timer == 0 {
		err = events.handleEvent(ctx, ns.ID, ns.Name, event)
		if err != nil {
			return err
		}
	} else {
		_, errs := events.appendStagingEvent(ctx, &datastore.StagingEvent{
			Event: &datastore.Event{
				NamespaceID: ns.ID,
				Event:       event,
				ReceivedAt:  time.Now().UTC(),
				Namespace:   ns.Name,
			},
			DatabaseID:   uuid.New(),
			DelayedUntil: time.Unix(timer, 0),
		})
		for _, err2 := range errs {
			if err2 != nil {
				slog.Error("failed to create delayed event", "error", err2)
			}
		}
	}

	return nil
}

func (events *events) listenForEvents(ctx context.Context, im *instanceMemory, ceds []*model.ConsumeEventDefinition, all bool) error {
	var transformedEvents []*model.ConsumeEventDefinition
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID()
	spanID := span.SpanContext().SpanID()

	slog.With("trace", traceID, "span", spanID, "namespace", im.Namespace(), "track", "namespace."+im.Namespace().Name, "instance", im.ID())

	for i := range ceds {
		ev := new(model.ConsumeEventDefinition)
		ev.Context = make(map[string]interface{})

		err := copier.Copy(ev, ceds[i])
		if err != nil {
			return err
		}

		for k, v := range ceds[i].Context {
			ev.Context[k], err = jqOne(im.data, v) //nolint:contextcheck
			if err != nil {
				return fmt.Errorf("failed to execute jq query for key '%s' on event definition %d: %w", k, i, err)
			}
		}

		transformedEvents = append(transformedEvents, ev)
	}

	err := events.addInstanceEventListener(ctx, im.Namespace().ID, im.Namespace().Name, im.GetInstanceID(), transformedEvents, all)
	if err != nil {
		return err
	}

	slog.Debug("Registered to receive events.")

	return nil
}
