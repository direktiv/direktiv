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
	ctx, end := traceAddtoEventlog(ctx)
	defer end()
	li := make([]*pkgevents.Event, 0)
	if eventin.ID() == "" {
		eventin.SetID(uuid.NewString())
	}
	li = append(li, &pkgevents.Event{
		Event:      eventin,
		Namespace:  ns.ID,
		ReceivedAt: time.Now(),
	})
	err := events.runSqlTx(ctx, func(tx *sqlTx) error {
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
	deletedIds := []*uuid.UUID{}
	err := events.runSqlTx(ctx, func(tx *sqlTx) error {
		ids, err := tx.DataStore().EventListener().DeleteAllForWorkflow(ctx, file.ID)
		deletedIds = ids
		return err
	})
	if err != nil {
		return err
	}
	for _, id := range deletedIds {
		err := events.runSqlTx(ctx, func(tx *sqlTx) error {
			return tx.DataStore().EventListenerTopics().Delete(ctx, *id)
		})
		if err != nil {
			return err
		}
	}
	events.pubsub.NotifyEventListeners(nsID)

	return nil
}

func (events *events) deleteInstanceEventListeners(ctx context.Context, im *instanceMemory) error {
	deletedIds := []*uuid.UUID{}
	err := events.runSqlTx(ctx, func(tx *sqlTx) error {
		ids, err := tx.DataStore().EventListener().DeleteAllForWorkflow(ctx, im.instance.Instance.ID)
		deletedIds = ids
		return err
	})
	if err != nil {
		return err
	}
	for _, id := range deletedIds {
		err := events.runSqlTx(ctx, func(tx *sqlTx) error {
			return tx.DataStore().EventListenerTopics().Delete(ctx, *id)
		})
		if err != nil {
			return err
		}
	}
	events.pubsub.NotifyEventListeners(im.instance.Instance.NamespaceID)

	return nil
}

func (events *events) processWorkflowEvents(ctx context.Context, nsID uuid.UUID, file *filestore.File, ms *muxStart) error {
	err := events.deleteWorkflowEventListeners(ctx, nsID, file)
	if err != nil {
		return err
	}

	p, err := convertToParseDurationFormat(ms.Lifespan)
	if err != nil {
		return err
	}
	lifespan, err := time.ParseDuration(p)
	// lifespan, err := duration.ParseISO8601(ms.Lifespan)
	if err != nil {
		return err
	}

	if len(ms.Events) > 0 && ms.Enabled {
		fEv := &pkgevents.EventListener{
			ID:                       uuid.New(),
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
			Deleted:                  false,
			NamespaceID:              nsID,
			TriggerType:              pkgevents.StartSimple,
			ListeningForEventTypes:   []string{},
			TriggerWorkflow:          file.ID,
			Metadata:                 file.Name(),
			LifespanOfReceivedEvents: int(lifespan.Milliseconds()),
			GlobGatekeepers:          make(map[string]string),
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
			for k, v := range sed.Context {
				fEv.GlobGatekeepers[sed.Type+"-"+k] = fmt.Sprintf("%v", v)
			}
		}

		err := events.runSqlTx(ctx, func(tx *sqlTx) error {
			return tx.DataStore().EventListener().Append(ctx, fEv)
		})
		if err != nil {
			return err
		}
		for _, t := range fEv.ListeningForEventTypes {
			err := events.runSqlTx(ctx, func(tx *sqlTx) error {
				return tx.DataStore().EventListenerTopics().Append(ctx, nsID, fEv.ID, nsID.String()+"-"+t)
			})
			if err != nil {
				return err
			}
		}
	}

	events.pubsub.NotifyEventListeners(nsID)

	return nil
}

// called from workflow instances to create event listeners.
func (events *events) addInstanceEventListener(ctx context.Context, namespace, instance uuid.UUID, sevents []*model.ConsumeEventDefinition, step int, all bool) error {
	// var ev []map[string]interface{}

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
		GlobGatekeepers: make(map[string]string),
	}

	for _, ced := range sevents {
		fEv.ListeningForEventTypes = append(fEv.ListeningForEventTypes, ced.Type)
		for k, v := range ced.Context {
			fEv.GlobGatekeepers[ced.Type+"-"+k] = fmt.Sprintf("%v", v)
		}
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
	for _, t := range fEv.ListeningForEventTypes {
		err := events.runSqlTx(ctx, func(tx *sqlTx) error {
			return tx.DataStore().EventListenerTopics().Append(ctx, namespace, fEv.ID, namespace.String()+"-"+t)
		})
		if err != nil {
			return err
		}
	}
	events.pubsub.NotifyEventListeners(namespace)

	return nil
}

func convertToParseDurationFormat(iso8601Duration string) (string, error) {
	if !strings.HasPrefix(iso8601Duration, "P") {
		return "", fmt.Errorf("invalid ISO8601 duration format")
	}

	durationStr := ""

	durationComponents := strings.Split(iso8601Duration[1:], "T")

	for _, component := range durationComponents {
		timeStr := strings.ReplaceAll(component, "H", "h")
		timeStr = strings.ReplaceAll(timeStr, "M", "m")
		timeStr = strings.ReplaceAll(timeStr, "S", "s")

		durationStr += timeStr
	}

	return durationStr, nil
}
