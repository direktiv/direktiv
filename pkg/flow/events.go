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
	"github.com/direktiv/direktiv/pkg/tracing"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
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

func (events *events) handleEvent(ctx context.Context, ns *datastore.Namespace, ce *cloudevents.Event) error {
	loggingCtx := tracing.WithTrack(ns.WithTags(ctx), tracing.BuildNamespaceTrack(ns.Name))
	loggingCtx, end, err := tracing.NewSpan(loggingCtx, "handling event-messages")
	if err != nil {
		slog.Warn("GetListenersByTopic failed to init telemetry", "error", err)
	}
	defer end()

	slog.DebugContext(loggingCtx, "handle CloudEvent started")
	e := pkgevents.EventEngine{
		WorkflowStart: func(ctx context.Context, workflowID uuid.UUID, ev ...*cloudevents.Event) {
			slog.DebugContext(loggingCtx, "starting workflow via CloudEvent.")
			//_, end := traceMessageTrigger(ctx, "wf: "+workflowID.String())
			//defer end()
			events.engine.EventsInvoke(ctx, workflowID, ev...) //nolint:contextcheck
		},
		WakeInstance: func(instanceID uuid.UUID, ev []*cloudevents.Event) {
			loggingCtx = tracing.AddTag(loggingCtx, "instance", instanceID)
			slog.DebugContext(loggingCtx, "invoking instance via cloudevent")
			//_, end := traceMessageTrigger(ctx, "ins: "+instanceID.String())
			//defer end()
			events.engine.WakeEventsWaiter(instanceID, ev) //nolint:contextcheck
		},
		GetListenersByTopic: func(ctx context.Context, s string) ([]*datastore.EventListener, error) {
			lCtx, end, err := tracing.NewSpan(loggingCtx, "fetching event-messages")
			if err != nil {
				slog.Warn("GetListenersByTopic failed to init telemetry", "error", err)
			}
			defer end()
			res := make([]*datastore.EventListener, 0)
			err = events.runSQLTx(ctx, func(tx *database.SQLStore) error {
				r, err := tx.DataStore().EventListenerTopics().GetListeners(ctx, s)
				if err != nil {
					slog.ErrorContext(lCtx, "failed fetching event-listener-topics.")
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
			slog.DebugContext(loggingCtx, "starting updating listeners.")
			err := events.runSQLTx(ctx, func(tx *database.SQLStore) error {
				errs := tx.DataStore().EventListener().UpdateOrDelete(ctx, listener)
				for _, err2 := range errs {
					if err2 != nil {
						slog.DebugContext(loggingCtx, "Error updating listeners.", "error", err2)

						return err2
					}
				}

				return nil
			})
			if err != nil {
				slog.ErrorContext(loggingCtx, "failed processing events", "error", err)
				return []error{fmt.Errorf("%w", err)}
			}
			slog.DebugContext(loggingCtx, "updating listeners complete.")

			return nil
		},
	}
	// TODO: ctx, end := traceProcessingMessage(ctx)
	// defer end()

	e.ProcessEvents(ctx, ns.ID, []event.Event{*ce}, func(template string, args ...interface{}) {
		slog.Error(fmt.Sprintf(template, args...))
	})
	slog.DebugContext(loggingCtx, "CloudEvent handled successfully")

	return nil
}

func (events *events) BroadcastCloudevent(ctx context.Context, ns *datastore.Namespace, event *cloudevents.Event, timer int64) error {
	loggingCtx := tracing.WithTrack(ns.WithTags(ctx), tracing.BuildNamespaceTrack(ns.Name))
	slog.DebugContext(loggingCtx, "received CloudEvent")

	// TODO: ctx, end := traceBrokerMessage(ctx, *event)
	// defer end()

	err := events.addEvent(ctx, event, ns)
	if err != nil {
		slog.ErrorContext(loggingCtx, "failed to add event", "error", err)
		return err
	}

	// handle event
	if timer == 0 {
		slog.Debug("handling event immediately")
		err = events.handleEvent(ctx, ns, event)
		if err != nil {
			slog.ErrorContext(loggingCtx, "failed to handle event", "error", err)
			return err
		}
	} else {
		slog.Debug("Scheduling delayed event", "delay_until", time.Unix(timer, 0))
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
				slog.ErrorContext(loggingCtx, "failed to create delayed event", "error", err2)
			}
		}
	}
	slog.DebugContext(loggingCtx, "processed CloudEvent successfully")

	return nil
}

func (events *events) listenForEvents(ctx context.Context, im *instanceMemory, ceds []*model.ConsumeEventDefinition, all bool) error {
	var transformedEvents []*model.ConsumeEventDefinition
	loggingCtx := im.Namespace().WithTags(ctx)
	instanceTrackCtx := tracing.WithTrack(loggingCtx, tracing.BuildInstanceTrack(im.instance))
	instanceTrackCtx, end, err := tracing.NewSpan(instanceTrackCtx, "waiting for events")
	if err != nil {
		slog.Warn("telemetry failed", "error", err)
	}
	defer end()
	slog.InfoContext(instanceTrackCtx, "listening for events")
	for i := range ceds {
		ev := new(model.ConsumeEventDefinition)
		ev.Context = make(map[string]interface{})

		err := copier.Copy(ev, ceds[i])
		if err != nil {
			slog.ErrorContext(instanceTrackCtx, "failed to copy event definition", "error", err)

			return err
		}

		for k, v := range ceds[i].Context {
			ev.Context[k], err = jqOne(im.data, v) //nolint:contextcheck
			if err != nil {
				err1 := fmt.Errorf("failed to execute jq query for key '%s' on event definition %d: %w", k, i, err)
				slog.ErrorContext(instanceTrackCtx, "Failed to execute jq query", "error", err1)

				return err1
			}
		}

		transformedEvents = append(transformedEvents, ev)
	}

	err = events.addInstanceEventListener(ctx, im.Namespace().ID, im.Namespace().Name, im.GetInstanceID(), transformedEvents, all)
	if err != nil {
		slog.ErrorContext(ctx, "failed to add instance event listener", "error", err)

		return err
	}
	slog.DebugContext(ctx, "successfully registered to receive events.")

	return nil
}
