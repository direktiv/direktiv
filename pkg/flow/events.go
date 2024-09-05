package flow

import (
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/direktiv/direktiv/pkg/core"
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

	slog.Debug("handle CloudEvent started", tracing.GetSlogAttributesWithStatus(loggingCtx, core.LogRunningStatus)...)
	e := pkgevents.EventEngine{
		WorkflowStart: func(ctx context.Context, workflowID uuid.UUID, ev ...*cloudevents.Event) {
			slog.Debug("starting workflow via CloudEvent.", tracing.GetSlogAttributesWithStatus(loggingCtx, core.LogRunningStatus)...)
			//_, end := traceMessageTrigger(ctx, "wf: "+workflowID.String())
			//defer end()
			events.engine.EventsInvoke(ctx, workflowID, ev...) //nolint:contextcheck
		},
		WakeInstance: func(instanceID uuid.UUID, ev []*cloudevents.Event) {
			slog.Debug("invoking instance via cloudevent", tracing.GetSlogAttributesWithStatus(tracing.AddTag(loggingCtx, "instance", instanceID), core.LogRunningStatus)...)
			//_, end := traceMessageTrigger(ctx, "ins: "+instanceID.String())
			//defer end()
			events.engine.WakeEventsWaiter(instanceID, ev) //nolint:contextcheck
		},
		GetListenersByTopic: func(ctx context.Context, s string) ([]*datastore.EventListener, error) {
			//ctx, end := traceGetListenersByTopic(ctx, s)
			//defer end()
			res := make([]*datastore.EventListener, 0)
			err := events.runSQLTx(ctx, func(tx *database.SQLStore) error {
				r, err := tx.DataStore().EventListenerTopics().GetListeners(ctx, s)
				if err != nil {
					slog.Error("failed fetching event-listener-topics.", tracing.GetSlogAttributesWithError(loggingCtx, err)...)
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
			slog.Debug("starting updating listeners.", tracing.GetSlogAttributesWithStatus(loggingCtx, core.LogRunningStatus)...)
			err := events.runSQLTx(ctx, func(tx *database.SQLStore) error {
				errs := tx.DataStore().EventListener().UpdateOrDelete(ctx, listener)
				for _, err2 := range errs {
					if err2 != nil {
						slog.Debug("Error updating listeners.", tracing.GetSlogAttributesWithError(loggingCtx, err2)...)

						return err2
					}
				}

				return nil
			})
			if err != nil {
				slog.Error("failed processing events", tracing.GetSlogAttributesWithError(loggingCtx, err)...)
				return []error{fmt.Errorf("%w", err)}
			}
			slog.Debug("updating listeners complete.", tracing.GetSlogAttributesWithStatus(loggingCtx, core.LogRunningStatus)...)

			return nil
		},
	}
	// TODO: ctx, end := traceProcessingMessage(ctx)
	// defer end()

	e.ProcessEvents(ctx, ns.ID, []event.Event{*ce}, func(template string, args ...interface{}) {
		slog.Error(fmt.Sprintf(template, args...))
	})
	slog.Debug("CloudEvent handled successfully", tracing.GetSlogAttributesWithStatus(loggingCtx, core.LogRunningStatus)...)

	return nil
}

func (events *events) BroadcastCloudevent(ctx context.Context, ns *datastore.Namespace, event *cloudevents.Event, timer int64) error {
	loggingCtx := tracing.WithTrack(ns.WithTags(ctx), tracing.BuildNamespaceTrack(ns.Name))
	slog.Debug("received CloudEvent", tracing.GetSlogAttributesWithStatus(loggingCtx, core.LogRunningStatus)...)

	// TODO: ctx, end := traceBrokerMessage(ctx, *event)
	// defer end()

	err := events.addEvent(ctx, event, ns)
	if err != nil {
		slog.Error("failed to add event", tracing.GetSlogAttributesWithError(loggingCtx, err)...)
		return err
	}

	// handle event
	if timer == 0 {
		slog.Debug("handling event immediately")
		err = events.handleEvent(ctx, ns, event)
		if err != nil {
			slog.Error("failed to handle event", tracing.GetSlogAttributesWithError(loggingCtx, err)...)
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
				slog.Error("failed to create delayed event", tracing.GetSlogAttributesWithError(loggingCtx, err2)...)
			}
		}
	}
	slog.Debug("processed CloudEvent successfully", tracing.GetSlogAttributesWithStatus(loggingCtx, core.LogRunningStatus)...)

	return nil
}

func (events *events) listenForEvents(ctx context.Context, im *instanceMemory, ceds []*model.ConsumeEventDefinition, all bool) error {
	var transformedEvents []*model.ConsumeEventDefinition
	loggingCtx := im.Namespace().WithTags(ctx)
	instanceTrackCtx := tracing.WithTrack(loggingCtx, tracing.BuildInstanceTrack(im.instance))

	slog.Info("listening for events", tracing.GetSlogAttributesWithStatus(instanceTrackCtx, core.LogRunningStatus)...)
	for i := range ceds {
		ev := new(model.ConsumeEventDefinition)
		ev.Context = make(map[string]interface{})

		err := copier.Copy(ev, ceds[i])
		if err != nil {
			slog.Error("failed to copy event definition", tracing.GetSlogAttributesWithError(ctx, err)...)

			return err
		}

		for k, v := range ceds[i].Context {
			ev.Context[k], err = jqOne(im.data, v) //nolint:contextcheck
			if err != nil {
				err1 := fmt.Errorf("failed to execute jq query for key '%s' on event definition %d: %w", k, i, err)
				slog.Error("Failed to execute jq query", tracing.GetSlogAttributesWithError(ctx, err1)...)

				return err1
			}
		}

		transformedEvents = append(transformedEvents, ev)
	}

	err := events.addInstanceEventListener(ctx, im.Namespace().ID, im.Namespace().Name, im.GetInstanceID(), transformedEvents, all)
	if err != nil {
		slog.Error("failed to add instance event listener", tracing.GetSlogAttributesWithError(ctx, err)...)

		return err
	}
	slog.Debug("successfully registered to receive events.", tracing.GetSlogAttributesWithStatus(ctx, core.LogRunningStatus)...)

	return nil
}
