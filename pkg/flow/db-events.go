package flow

import (
	"context"
	"fmt"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entcev "github.com/direktiv/direktiv/pkg/flow/ent/cloudevents"
	entev "github.com/direktiv/direktiv/pkg/flow/ent/events"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	"github.com/direktiv/direktiv/pkg/flow/ent/workflow"
	entwf "github.com/direktiv/direktiv/pkg/flow/ent/workflow"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/google/uuid"
)

func (events *events) markEventAsProcessed(ctx context.Context, cevc *ent.CloudEventsClient, eventID string) (*cloudevents.Event, error) {

	e, err := cevc.Query().Where(entcev.EventId(eventID)).Only(ctx)
	if err != nil {
		return nil, err
	}

	if e.Processed {
		return nil, fmt.Errorf("event already processed")
	}

	e, err = e.Update().SetProcessed(true).Save(ctx)
	if err != nil {
		return nil, err
	}

	ev := cloudevents.Event(e.Event)

	return &ev, nil

}

func (events *events) deleteExpiredEvents(ctx context.Context, cevc *ent.CloudEventsClient) error {

	_, err := cevc.Delete().
		Where(entcev.And(entcev.Processed(true), entcev.FireLT(time.Now().Add(-1*time.Hour)))).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil

}

func (events *events) getEarliestEvent(ctx context.Context, cevc *ent.CloudEventsClient) (*ent.CloudEvents, error) {

	e, err := cevc.Query().
		Where(entcev.Processed(false)).
		Order(ent.Asc(entcev.FieldFire)).
		WithNamespace().
		First(ctx)

	if err != nil {
		return nil, err
	}

	return e, nil

}

func (events *events) addEvent(ctx context.Context, cevc *ent.CloudEventsClient, eventin *cloudevents.Event, ns *ent.Namespace, delay int64) error {

	t := time.Now().Unix() + delay

	processed := (delay == 0)

	ev := event.Event(*eventin)

	_, err := cevc.
		Create().
		SetEvent(ev).
		SetNamespace(ns).
		SetFire(time.Unix(t, 0)).
		SetProcessed(processed).
		SetEventId(eventin.ID()).
		Save(ctx)
	if err != nil {
		return err
	}

	return nil

}

func (events *events) getWorkflowEventByWorkflowUID(ctx context.Context, evc *ent.EventsClient, id uuid.UUID) (*ent.Events, error) {

	evs, err := evc.Query().
		Where(entev.HasWorkflowWith(
			workflow.IDEQ(id),
		)).
		WithWorkflow().
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return evs, nil

}

func (events *events) deleteWorkflowEventListeners(ctx context.Context, evc *ent.EventsClient, wf *ent.Workflow) error {

	var err error
	ns := wf.Edges.Namespace
	if ns == nil {
		ns, err = wf.Namespace(ctx)
		if err != nil {
			return err
		}
	}

	_, err = evc.
		Delete().
		Where(entev.HasWorkflowWith(entwf.ID(wf.ID))).
		Exec(ctx)
	if err != nil {
		return err
	}

	events.pubsub.NotifyEventListeners(ns)

	return nil

}

func (events *events) deleteInstanceEventListeners(ctx context.Context, in *ent.Instance) error {

	ns := in.Edges.Namespace

	_, err := events.db.Events.
		Delete().
		Where(entev.HasInstanceWith(entinst.ID(in.ID))).
		Exec(ctx)
	if err != nil {
		return err
	}

	events.pubsub.NotifyEventListeners(ns)

	return nil

}

func (events *events) addWorkflowEventWait(ctx context.Context, ewc *ent.EventsWaitClient, ev map[string]interface{}, count int, id uuid.UUID) error {

	_, err := ewc.Create().
		SetEvents(ev).
		SetWorkfloweventID(id).
		Save(ctx)

	if err != nil {
		return err
	}

	return nil

}

// called by add workflow, adds event listeners if required
func (events *events) processWorkflowEvents(ctx context.Context, evc *ent.EventsClient,
	wf *ent.Workflow, ms *muxStart) error {

	err := events.deleteWorkflowEventListeners(ctx, evc, wf)
	if err != nil {
		return err
	}

	ns := wf.Edges.Namespace
	if ns == nil {
		ns, err = wf.Namespace(ctx)
		if err != nil {
			return err
		}
	}

	if len(ms.Events) > 0 && ms.Enabled {

		var ev []map[string]interface{}
		for _, e := range ms.Events {
			em := make(map[string]interface{})
			em[eventTypeString] = e.Type

			for kf, vf := range e.Context {
				em[fmt.Sprintf("%s%s", filterPrefix, strings.ToLower(kf))] = vf
			}
			ev = append(ev, em)
		}

		correlations := []string{}
		count := 1

		if ms.Type == model.StartTypeEventsAnd.String() {
			correlations = append(correlations, ms.Correlate...)
			count = len(ms.Events)
		}

		_, err = evc.Create().
			SetNamespace(ns).
			SetWorkflow(wf).
			SetEvents(ev).
			SetCorrelations(correlations).
			SetCount(count).
			Save(ctx)

		if err != nil {
			return err
		}

	}

	events.pubsub.NotifyEventListeners(ns)

	return nil

}

// called from workflow instances to create event listeners
func (events *events) addInstanceEventListener(ctx context.Context, evc *ent.EventsClient, wf *ent.Workflow, in *ent.Instance,
	sevents []*model.ConsumeEventDefinition, signature []byte, all bool) error {

	var ev []map[string]interface{}
	for _, e := range sevents {
		em := make(map[string]interface{})
		em[eventTypeString] = e.Type

		for kf, vf := range e.Context {
			em[fmt.Sprintf("%s%s", filterPrefix, strings.ToLower(kf))] = vf
		}
		ev = append(ev, em)
	}

	count := 1
	if all {
		count = len(sevents)
	}

	ns := wf.Edges.Namespace

	_, err := evc.Create().
		SetNamespace(ns).
		SetWorkflow(wf).
		SetInstance(in).
		SetEvents(ev).
		SetCorrelations([]string{}).
		SetSignature(signature).
		SetCount(count).
		Save(ctx)

	if err != nil {
		return err
	}

	events.pubsub.NotifyEventListeners(ns)

	return nil

}
