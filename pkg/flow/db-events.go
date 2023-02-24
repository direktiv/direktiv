package flow

import (
	"context"
	"fmt"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/flow/ent"
	entcev "github.com/direktiv/direktiv/pkg/flow/ent/cloudevents"
	entev "github.com/direktiv/direktiv/pkg/flow/ent/events"
	entinst "github.com/direktiv/direktiv/pkg/flow/ent/instance"
	entwf "github.com/direktiv/direktiv/pkg/flow/ent/workflow"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/google/uuid"
)

func (events *events) markEventAsProcessed(ctx context.Context, tx database.Transaction, eventID string) (*cloudevents.Event, error) {
	clients := events.edb.Clients(tx)

	e, err := clients.CloudEvents.Query().Where(entcev.EventId(eventID)).Only(ctx)
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

	ev := e.Event

	return &ev, nil
}

func (events *events) getEarliestEvent(ctx context.Context, tx database.Transaction) (*ent.CloudEvents, error) {
	clients := events.edb.Clients(tx)

	e, err := clients.CloudEvents.Query().
		Where(entcev.Processed(false)).
		Order(ent.Asc(entcev.FieldFire)).
		WithNamespace().
		First(ctx)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (events *events) addEvent(ctx context.Context, tx database.Transaction, eventin *cloudevents.Event, ns *database.Namespace, delay int64) error {
	t := time.Now().Unix() + delay

	processed := (delay == 0)

	ev := *eventin

	clients := events.edb.Clients(tx)

	_, err := clients.CloudEvents.
		Create().
		SetEvent(ev).
		SetNamespaceID(ns.ID).
		SetFire(time.Unix(t, 0)).
		SetProcessed(processed).
		SetEventId(eventin.ID()).
		Save(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (events *events) deleteEventListeners(ctx context.Context, tx database.Transaction, cached *database.CacheData, id uuid.UUID) error {
	clients := events.edb.Clients(tx)

	_, err := clients.Events.Delete().Where(entev.IDEQ(id)).Exec(ctx)
	if err != nil {
		return err
	}

	events.pubsub.NotifyEventListeners(cached.Namespace)

	return nil
}

func (events *events) deleteWorkflowEventListeners(ctx context.Context, tx database.Transaction, cached *database.CacheData) error {
	clients := events.edb.Clients(tx)

	_, err := clients.Events.Delete().Where(entev.HasWorkflowWith(entwf.ID(cached.Workflow.ID))).Exec(ctx)
	if err != nil {
		return err
	}

	events.pubsub.NotifyEventListeners(cached.Namespace)

	return nil
}

func (events *events) deleteInstanceEventListeners(ctx context.Context, tx database.Transaction, cached *database.CacheData) error {
	clients := events.edb.Clients(tx)

	_, err := clients.Events.
		Delete().
		Where(entev.HasInstanceWith(entinst.ID(cached.Instance.ID))).
		Exec(ctx)
	if err != nil {
		return err
	}

	events.pubsub.NotifyEventListeners(cached.Namespace)

	return nil
}

// called by add workflow, adds event listeners if required.
func (events *events) processWorkflowEvents(ctx context.Context, tx database.Transaction, cached *database.CacheData, ms *muxStart) error {
	err := events.deleteWorkflowEventListeners(ctx, tx, cached)
	if err != nil {
		return err
	}

	if len(ms.Events) > 0 && ms.Enabled {

		var ev []map[string]interface{}
		for i, e := range ms.Events {
			em := make(map[string]interface{})
			em[eventTypeString] = e.Type

			for kf, vf := range e.Context {
				em[fmt.Sprintf("%s%s", filterPrefix, strings.ToLower(kf))] = vf
			}

			// these value are set when a matching event comes in
			em["time"] = 0
			em["value"] = ""
			em["idx"] = i

			ev = append(ev, em)
		}

		correlations := []string{}
		count := 1

		if ms.Type == model.StartTypeEventsAnd.String() {
			count = len(ms.Events)
		}

		clients := events.edb.Clients(tx)

		_, err = clients.Events.Create().
			SetNamespaceID(cached.Namespace.ID).
			SetWorkflowID(cached.Workflow.ID).
			SetEvents(ev).
			SetCorrelations(correlations).
			SetCount(count).
			Save(ctx)

		if err != nil {
			return err
		}

	}

	events.pubsub.NotifyEventListeners(cached.Namespace)

	return nil
}

func (events *events) updateInstanceEventListener(ctx context.Context, tx database.Transaction, id uuid.UUID, ev []map[string]interface{}) error {
	clients := events.edb.Clients(tx)

	_, err := clients.Events.UpdateOneID(id).SetEvents(ev).Save(ctx)
	return err
}

// called from workflow instances to create event listeners.
func (events *events) addInstanceEventListener(ctx context.Context, tx database.Transaction,
	cached *database.CacheData, sevents []*model.ConsumeEventDefinition, signature []byte, all bool,
) error {
	// add
	var ev []map[string]interface{}
	for i, e := range sevents {
		em := make(map[string]interface{})
		em[eventTypeString] = e.Type

		for kf, vf := range e.Context {
			em[fmt.Sprintf("%s%s", filterPrefix, strings.ToLower(kf))] = vf
		}

		// these value are set when a matching event comes in
		em["time"] = 0
		em["value"] = ""
		em["idx"] = i

		ev = append(ev, em)
	}

	count := 1
	if all {
		count = len(sevents)
	}

	clients := events.edb.Clients(tx)

	_, err := clients.Events.Create().
		SetNamespaceID(cached.Namespace.ID).
		SetWorkflowID(cached.Workflow.ID).
		SetInstanceID(cached.Instance.ID).
		SetEvents(ev).
		SetCorrelations([]string{}).
		SetSignature(signature).
		SetCount(count).
		Save(ctx)
	if err != nil {
		return err
	}

	events.pubsub.NotifyEventListeners(cached.Namespace)

	return nil
}
