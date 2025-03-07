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
	"github.com/direktiv/direktiv/pkg/telemetry"
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
	// ctx = tracing.WithTrack(tracing.AddNamespace(ctx, ns.Name), tracing.BuildNamespaceTrack(ns.Name))
	// ctx, end, err := tracing.NewSpan(ctx, "handling event-messages")
	// if err != nil {
	// 	slog.Debug("getListenersByTopic failed to init telemetry", "error", err)
	// }
	// defer end()
	slog.DebugContext(ctx, "handle cloudevent started")
	e := pkgevents.EventEngine{
		WorkflowStart: func(ctx context.Context, workflowID uuid.UUID, ev ...*cloudevents.Event) {
			ctx = tracing.WithTrack(tracing.AddNamespace(ctx, ns.Name), tracing.BuildNamespaceTrack(ns.Name))
			ctx, end, err := tracing.NewSpan(ctx, "starting workflow via cloudevent")
			if err != nil {
				slog.Debug("workflowStart failed to init telemetry", "error", err)
			}
			defer end()
			slog.DebugContext(ctx, "starting workflow via cloudevents")
			events.Engine.EventsInvoke(ctx, workflowID, ev...) //nolint:contextcheck
		},
		WakeInstance: func(instanceID uuid.UUID, ev []*cloudevents.Event) {
			// nolint:fatcontext
			ctx = tracing.AddLoseInstanceIDAttr(ctx, instanceID.String())
			ctx = tracing.WithTrack(tracing.AddNamespace(ctx, ns.Name), tracing.BuildNamespaceTrack(ns.Name))
			ctx, end, err := tracing.NewSpan(ctx, "waking instance via cloudevent")
			if err != nil {
				slog.Debug("wakeInstance failed to init telemetry", "error", err)
			}
			defer end()
			slog.DebugContext(ctx, "invoking instance via cloudevent")
			events.Engine.WakeEventsWaiter(instanceID, ev) //nolint:contextcheck
		},
		GetListenersByTopic: func(ctx context.Context, s string) ([]*datastore.EventListener, error) {
			ctx = tracing.WithTrack(tracing.AddNamespace(ctx, ns.Name), tracing.BuildNamespaceTrack(ns.Name))
			ctx, end, err := tracing.NewSpan(ctx, "fetching cloudevens from event bus")
			if err != nil {
				slog.Debug("getListenersByTopic failed to init telemetry", "error", err)
			}
			defer end()
			res := make([]*datastore.EventListener, 0)
			err = events.runSQLTx(ctx, func(tx *database.DB) error {
				r, err := tx.DataStore().EventListenerTopics().GetListeners(ctx, s)
				if err != nil {
					slog.ErrorContext(ctx, "failed fetching event-listener-topics")
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
			ctx = tracing.WithTrack(tracing.AddNamespace(ctx, ns.Name), tracing.BuildNamespaceTrack(ns.Name))
			ctx, end, err := tracing.NewSpan(ctx, "updating even-listeners in the event bus")
			if err != nil {
				slog.Debug("updateListeners:c failed to init telemetry", "error", err)
			}
			defer end()
			err = events.runSQLTx(ctx, func(tx *database.DB) error {
				errs := tx.DataStore().EventListener().UpdateOrDelete(ctx, listener)
				for _, err2 := range errs {
					if err2 != nil {
						slog.DebugContext(ctx, "error updating listeners", "error", err2)

						return err2
					}
				}

				return nil
			})
			if err != nil {
				slog.ErrorContext(ctx, "failed processing events", "error", err)
				return []error{fmt.Errorf("%w", err)}
			}
			slog.DebugContext(ctx, "updating listeners complete")

			return nil
		},
	}

	e.ProcessEvents(ctx, ns.ID, []event.Event{*ce}, func(template string, args ...interface{}) {
		slog.ErrorContext(ctx, fmt.Sprintf(template, args...))
	})
	slog.DebugContext(ctx, "cloudevent handled")

	return nil
}

func (events *events) BroadcastCloudevent(ctx context.Context, ns *datastore.Namespace, event *cloudevents.Event, timer int64) error {
	loggingCtx := tracing.WithTrack(tracing.AddNamespace(ctx, ns.Name), tracing.BuildNamespaceTrack(ns.Name))
	loggingCtx, cleanup, err := tracing.NewSpan(loggingCtx, "adding cloudevent to the event bus with id: "+event.ID())
	if err != nil {
		slog.Debug("failed to popupate telemetry in broadcast-cloudevent", "error", err)
	}
	defer cleanup()
	slog.DebugContext(loggingCtx, "received cloudevent")
	err = events.addEvent(ctx, event, ns)
	if err != nil {
		slog.ErrorContext(loggingCtx, "failed to add event", "error", err)
		return err
	}
	// handle event
	if timer == 0 {
		slog.DebugContext(loggingCtx, "handling event immediately")
		err = events.handleEvent(ctx, ns, event)
		if err != nil {
			slog.ErrorContext(loggingCtx, "failed to handle event", "error", err)
			return err
		}
	} else {
		slog.DebugContext(loggingCtx, "scheduling delayed event", "delay-until", time.Unix(timer, 0))
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
	slog.DebugContext(loggingCtx, "processed cloudevent")

	return nil
}

func (events *events) listenForEvents(ctx context.Context, im *instanceMemory, ceds []*model.ConsumeEventDefinition, all bool) error {
	var transformedEvents []*model.ConsumeEventDefinition
	// loggingCtx := tracing.AddNamespace(ctx, im.Namespace().Name)
	// instanceTrackCtx := tracing.WithTrack(loggingCtx, tracing.BuildInstanceTrack(im.instance))
	// instanceTrackCtx, end, err := tracing.NewSpan(instanceTrackCtx, "waiting for events")
	// if err != nil {
	// 	slog.Debug("telemetry failed", "error", err)
	// }
	// defer end()
	ctx = im.Context(ctx)
	telemetry.LogInstanceInfo(ctx, fmt.Sprintf("listening for %d events", len(ceds)))
	for i := range ceds {
		ev := new(model.ConsumeEventDefinition)
		ev.Context = make(map[string]interface{})

		err := copier.Copy(ev, ceds[i])
		if err != nil {
			telemetry.LogInstanceError(ctx, "failed to copy event definition", err)

			return err
		}

		for k, v := range ceds[i].Context {
			ev.Context[k], err = jqOne(im.data, v) //nolint:contextcheck
			if err != nil {
				err1 := fmt.Errorf("failed to execute jq query for key '%s' on event definition %d: %w", k, i, err)
				telemetry.LogInstanceError(ctx, "failed to execute jq query", err1)

				return err1
			}
		}

		transformedEvents = append(transformedEvents, ev)
	}

	err := events.addInstanceEventListener(ctx, im.Namespace().ID, im.Namespace().Name, im.GetInstanceID(), transformedEvents, all)
	if err != nil {
		telemetry.LogInstanceError(ctx, "failed to add instance event listener", err)

		return err
	}
	telemetry.LogInstanceDebug(ctx, "registered to receive events")

	return nil
}
