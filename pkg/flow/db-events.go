package flow

import (
	"context"
	"fmt"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore"
	pkgevents "github.com/direktiv/direktiv/pkg/refactor/events"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
)

func (events *events) addEvent(ctx context.Context, eventin *cloudevents.Event, ns *datastore.Namespace) error {
	ctx, end := traceAddtoEventlog(ctx)
	defer end()
	li := make([]*pkgevents.Event, 0)
	if eventin.ID() == "" {
		eventin.SetID(uuid.NewString())
	}
	li = append(li, &pkgevents.Event{
		Event:         eventin,
		Namespace:     ns.ID,
		NamespaceName: ns.Name,
		ReceivedAt:    time.Now().UTC(),
	})
	err := events.runSQLTx(ctx, func(tx *database.SQLStore) error {
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

func (events *events) deleteWorkflowEventListeners(ctx context.Context, nsID uuid.UUID, fileID uuid.UUID) error {
	err := events.runSQLTx(ctx, func(tx *database.SQLStore) error {
		ids, err := tx.DataStore().EventListener().DeleteAllForWorkflow(ctx, fileID)
		if err != nil {
			return err
		}

		for _, id := range ids {
			err = tx.DataStore().EventListenerTopics().Delete(ctx, *id)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	events.pubsub.NotifyEventListeners(nsID)

	return nil
}

func (events *events) deleteInstanceEventListeners(ctx context.Context, im *instanceMemory) error {
	err := events.runSQLTx(ctx, func(tx *database.SQLStore) error {
		ids, err := tx.DataStore().EventListener().DeleteAllForWorkflow(ctx, im.instance.Instance.ID)
		if err != nil {
			return err
		}

		for _, id := range ids {
			err = tx.DataStore().EventListenerTopics().Delete(ctx, *id)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	events.pubsub.NotifyEventListeners(im.instance.Instance.NamespaceID)

	return nil
}

func (events *events) processWorkflowEvents(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File, ms *muxStart) error {
	err := events.deleteWorkflowEventListeners(ctx, nsID, file.ID)
	if err != nil {
		return err
	}
	var lifespan time.Duration
	if ms.Lifespan != "" {
		p, err := convertToParseDurationFormat(ms.Lifespan)
		if err != nil {
			return err
		}
		lifespan, err = time.ParseDuration(p)
		// lifespan, err := duration.ParseISO8601(ms.Lifespan)
		if err != nil {
			return err
		}
	}

	if len(ms.Events) > 0 {
		fEv := &pkgevents.EventListener{
			ID:                       uuid.New(),
			CreatedAt:                time.Now().UTC(),
			UpdatedAt:                time.Now().UTC(),
			Deleted:                  false,
			NamespaceID:              nsID,
			TriggerType:              pkgevents.StartSimple,
			ListeningForEventTypes:   []string{},
			TriggerWorkflow:          file.ID.String(),
			Metadata:                 file.Path,
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
		contextFilters := make([]string, 0, len(ms.Events))
		for _, sed := range ms.Events {
			fEv.ListeningForEventTypes = append(fEv.ListeningForEventTypes, sed.Type)
			databaseNoDupCheck := ""
			for k, v := range sed.Context {
				databaseNoDupCheck += fmt.Sprintf("%v %v %v", sed.Type, k, v)
				fEv.GlobGatekeepers[sed.Type+"-"+k] = fmt.Sprintf("%v", v)
			}
			contextFilters = append(contextFilters, databaseNoDupCheck)
		}

		err := events.runSQLTx(ctx, func(tx *database.SQLStore) error {
			err := tx.DataStore().EventListener().Append(ctx, fEv)
			if err != nil {
				return err
			}

			for i, t := range fEv.ListeningForEventTypes {
				err = tx.DataStore().EventListenerTopics().Append(ctx, nsID, nsName, fEv.ID, nsID.String()+"-"+t, contextFilters[i])
				if err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	events.pubsub.NotifyEventListeners(nsID)

	return nil
}

// called from workflow instances to create event listeners.
func (events *events) addInstanceEventListener(ctx context.Context, namespace uuid.UUID, nsName string, instance uuid.UUID, sevents []*model.ConsumeEventDefinition, step int, all bool) error {
	// var ev []map[string]interface{}

	fEv := &pkgevents.EventListener{
		ID:                     uuid.New(),
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
		Deleted:                false,
		NamespaceID:            namespace,
		TriggerType:            pkgevents.WaitSimple,
		ListeningForEventTypes: []string{},
		TriggerInstance:        instance.String(),
		TriggerInstanceStep:    step,
		// LifespanOfReceivedEvents: , TODO?
		GlobGatekeepers: make(map[string]string),
	}
	contextFilters := make([]string, 0, len(sevents))

	for _, ced := range sevents {
		fEv.ListeningForEventTypes = append(fEv.ListeningForEventTypes, ced.Type)
		databaseNoDupCheck := ""
		for k, v := range ced.Context {
			databaseNoDupCheck += fmt.Sprintf("%v %v %v", ced.Type, k, v)
			fEv.GlobGatekeepers[ced.Type+"-"+k] = fmt.Sprintf("%v", v)
		}
		contextFilters = append(contextFilters, databaseNoDupCheck)
	}
	if all {
		fEv.TriggerType = pkgevents.WaitAnd
	}
	if !all && len(fEv.ListeningForEventTypes) > 1 {
		fEv.TriggerType = pkgevents.WaitOR
	}

	err := events.runSQLTx(ctx, func(tx *database.SQLStore) error {
		err := tx.DataStore().EventListener().Append(ctx, fEv)
		if err != nil {
			return err
		}
		for i, t := range fEv.ListeningForEventTypes {
			err = tx.DataStore().EventListenerTopics().Append(ctx, namespace, nsName, fEv.ID, namespace.String()+"-"+t, contextFilters[i])
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
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
