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
	slog.Debug("handle cloudevent started")
	e := pkgevents.EventEngine{
		WorkflowStart: func(ctx context.Context, workflowID uuid.UUID, ev ...*cloudevents.Event) {
			slog.Debug("starting workflow via cloudevents")
			events.Engine.EventsInvoke(ctx, workflowID, ev...) //nolint:contextcheck
		},
		WakeInstance: func(instanceID uuid.UUID, ev []*cloudevents.Event) {
			slog.Debug("invoking instance via cloudevent")
			events.Engine.WakeEventsWaiter(instanceID, ev) //nolint:contextcheck
		},
		GetListenersByTopic: func(ctx context.Context, s string) ([]*datastore.EventListener, error) {
			res := make([]*datastore.EventListener, 0)
			err := events.runSQLTx(ctx, func(tx *database.DB) error {
				r, err := tx.DataStore().EventListenerTopics().GetListeners(ctx, s)
				if err != nil {
					slog.Error("failed fetching event-listener-topics")
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
			err := events.runSQLTx(ctx, func(tx *database.DB) error {
				errs := tx.DataStore().EventListener().UpdateOrDelete(ctx, listener)
				for _, err2 := range errs {
					if err2 != nil {
						slog.Debug("error updating listeners", "error", err2)

						return err2
					}
				}

				return nil
			})
			if err != nil {
				slog.Error("failed processing events", "error", err)
				return []error{fmt.Errorf("%w", err)}
			}
			slog.Debug("updating listeners complete")

			return nil
		},
	}

	e.ProcessEvents(ctx, ns.ID, []event.Event{*ce}, func(template string, args ...interface{}) {
		slog.Error(fmt.Sprintf(template, args...))
	})
	slog.Debug("cloudevent handled")

	return nil
}

func (events *events) BroadcastCloudevent(ctx context.Context, ns *datastore.Namespace, event *cloudevents.Event, timer int64) error {
	slog.Debug("received cloudevent")
	err := events.addEvent(ctx, event, ns)
	if err != nil {
		slog.Error("failed to add event", "error", err)
		return err
	}
	// handle event
	if timer == 0 {
		slog.Debug("handling event immediately")
		err = events.handleEvent(ctx, ns, event)
		if err != nil {
			slog.Error("failed to handle event", "error", err)
			return err
		}
	} else {
		slog.Debug("scheduling delayed event", "delay-until", time.Unix(timer, 0))
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
	slog.Debug("processed cloudevent")

	return nil
}

func (events *events) listenForEvents(ctx context.Context, im *instanceMemory, ceds []*model.ConsumeEventDefinition, all bool) error {
	var transformedEvents []*model.ConsumeEventDefinition
	ctx = im.Context(ctx)
	telemetry.LogInstance(ctx, telemetry.LogLevelInfo,
		fmt.Sprintf("listening for %d events", len(ceds)))

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

	telemetry.LogInstance(ctx, telemetry.LogLevelDebug,
		"registered to receive events")

	return nil
}
