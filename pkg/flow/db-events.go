package flow

import (
	"context"
	"fmt"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/flow/database"
	"github.com/direktiv/direktiv/pkg/model"
	pkgevents "github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
)

func (events *events) addEvent(ctx context.Context, eventin *cloudevents.Event, ns *database.Namespace, delay int64) error {
	// t := time.Now().Unix() + delay

	// processed := delay == 0 //TODO:

	li := make([]*pkgevents.Event, 0)
	_, err := uuid.Parse(eventin.ID())
	if err != nil {
		eventin.SetID(uuid.NewString())
	}
	li = append(li, &pkgevents.Event{
		Event:      eventin,
		Namespace:  ns.ID,
		ReceivedAt: time.Now(),
	})
	err = events.runSqlTx(ctx, func(tx *sqlTx) error {
		_, errs := tx.DataStore().EventHistory().Append(ctx, li)
		for _, err2 := range errs {
			if err2 != nil {
				return err2
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (events *events) deleteWorkflowEventListeners(ctx context.Context, nsID uuid.UUID, file *filestore.File) error {
	err := events.runSqlTx(ctx, func(tx *sqlTx) error {
		return tx.DataStore().EventListener().DeleteAllForWorkflow(ctx, file.ID)
	})
	if err != nil {
		return err
	}
	events.pubsub.NotifyEventListeners(nsID)

	return nil
}

func (events *events) deleteInstanceEventListeners(ctx context.Context, im *instanceMemory) error {
	err := events.runSqlTx(ctx, func(tx *sqlTx) error {
		return tx.DataStore().EventListener().DeleteAllForWorkflow(ctx, im.instance.Instance.ID)
	})
	if err != nil {
		return err
	}

	events.pubsub.NotifyEventListeners(im.instance.Instance.NamespaceID)

	return nil
}

func (events *events) processWorkflowEvents(ctx context.Context, nsID uuid.UUID, file *filestore.File, ms *muxStart) error {
	err := events.deleteWorkflowEventListeners(ctx, nsID, file)
	if err != nil {
		return err
	}

	if len(ms.Events) > 0 && ms.Enabled {
		// var ev []map[string]interface{}
		for _, e := range ms.Events {
			em := make(map[string]interface{})
			em[eventTypeString] = e.Type

			for kf, vf := range e.Context {
				em[fmt.Sprintf("%s%s", filterPrefix, strings.ToLower(kf))] = vf
			}

			// // these value are set when a matching event comes in
			// em["time"] = 0
			// em["value"] = ""
			// em["idx"] = i

			// ev = append(ev, em)
		}

		fEv := &pkgevents.EventListener{
			ID:                     uuid.New(),
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
			Deleted:                false,
			NamespaceID:            nsID,
			TriggerType:            pkgevents.StartSimple,
			ListeningForEventTypes: []string{},
			TriggerWorkflow:        file.ID,
			// LifespanOfReceivedEvents: , TODO?
			// GlobGatekeepers: , TODO
		}
		switch ms.Type {
		case "default":
			fEv.TriggerType = pkgevents.StartSimple
		case "event":
			fEv.TriggerType = pkgevents.StartSimple // TODO: is this correct?
		case "eventsXor":
			fEv.TriggerType = pkgevents.StartOR // TODO: is this correct?
		case "eventsAnd":
			fEv.TriggerType = pkgevents.StartAnd // TODO: is this correct?
		}
		for _, sed := range ms.Events {
			fEv.ListeningForEventTypes = append(fEv.ListeningForEventTypes, sed.Type)
		}

		err := events.runSqlTx(ctx, func(tx *sqlTx) error {
			return tx.DataStore().EventListener().Append(ctx, fEv)
		})
		if err != nil {
			return err
		}
	}

	events.pubsub.NotifyEventListeners(nsID)

	return nil
}

// called from workflow instances to create event listeners.
func (events *events) addInstanceEventListener(ctx context.Context, namespace, instance uuid.UUID, sevents []*model.ConsumeEventDefinition, step int, all bool) error {
	// var ev []map[string]interface{}
	for _, e := range sevents {
		em := make(map[string]interface{})
		em[eventTypeString] = e.Type

		for kf, vf := range e.Context {
			em[fmt.Sprintf("%s%s", filterPrefix, strings.ToLower(kf))] = vf
		}

		// // these value are set when a matching event comes in
		// em["time"] = 0
		// em["value"] = ""
		// em["idx"] = i

		// ev = append(ev, em)
	}

	fEv := &pkgevents.EventListener{
		ID:                     uuid.New(),
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		Deleted:                false,
		NamespaceID:            namespace,
		TriggerType:            pkgevents.WaitSimple,
		ListeningForEventTypes: []string{},
		TriggerInstance:        instance,
		TriggerInstanceStep:    step,
		// LifespanOfReceivedEvents: , TODO?
		// GlobGatekeepers: , TODO
	}
	for _, ced := range sevents {
		fEv.ListeningForEventTypes = append(fEv.ListeningForEventTypes, ced.Type)
	}
	if all {
		fEv.TriggerType = pkgevents.WaitAnd
	}
	if !all && len(fEv.ListeningForEventTypes) > 1 {
		fEv.TriggerType = pkgevents.WaitOR
	}

	err := events.runSqlTx(ctx, func(tx *sqlTx) error {
		return tx.DataStore().EventListener().Append(ctx, fEv)
	})
	if err != nil {
		return err
	}

	events.pubsub.NotifyEventListeners(namespace)

	return nil
}
