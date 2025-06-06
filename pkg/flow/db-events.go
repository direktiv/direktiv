package flow

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/model"
	"github.com/direktiv/direktiv/pkg/telemetry"
	"github.com/google/uuid"
)

func (events *events) addEvent(ctx context.Context, eventin *cloudevents.Event, ns *datastore.Namespace) error {
	telemetry.LogNamespace(telemetry.LogLevelDebug, ns.Name, "event-bus registering event")

	li := make([]*datastore.Event, 0)
	if eventin.ID() == "" {
		eventin.SetID(uuid.NewString())
	}
	li = append(li, &datastore.Event{
		Event:       eventin,
		NamespaceID: ns.ID,
		Namespace:   ns.Name,
		ReceivedAt:  time.Now().UTC(),
	})
	err := events.runSQLTx(ctx, func(tx *database.DB) error {
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
	err := events.runSQLTx(ctx, func(tx *database.DB) error {
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

	return nil
}

func (events *events) deleteInstanceEventListeners(ctx context.Context, im *instanceMemory) error {
	err := events.runSQLTx(ctx, func(tx *database.DB) error {
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

	return nil
}

func RenderAllStartEventListeners(ctx context.Context, tx *database.DB) error {
	nsList, err := tx.DataStore().Namespaces().GetAll(ctx)
	if err != nil {
		return err
	}
	for _, ns := range nsList {
		files, err := tx.FileStore().ForNamespace(ns.Name).ListDirektivFilesWithData(ctx)
		if err != nil {
			return err
		}
		for _, file := range files {
			ms, err := validateRouter(ctx, tx, file)
			if err != nil {
				slog.Debug("render event-listeners", "error", err)
				continue
			}

			err = renderStartEventListener(ctx, ns.ID, ns.Name, file, ms, tx)
			if err != nil {
				slog.Debug("render event-listeners", "error", err)
				continue
			}
		}
	}

	return nil
}

func renderStartEventListener(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File, ms *muxStart, tx *database.DB) error {
	_, err := tx.DataStore().EventListener().DeleteAllForWorkflow(ctx, file.ID)
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
		err := appendEventListenersToDB(ctx, nsID, nsName, file, lifespan, ms, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func appendEventListenersToDB(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File, lifespan time.Duration, ms *muxStart, tx *database.DB) error {
	fEv := &datastore.EventListener{
		ID:                       uuid.New(),
		CreatedAt:                time.Now().UTC(),
		UpdatedAt:                time.Now().UTC(),
		Deleted:                  false,
		NamespaceID:              nsID,
		Namespace:                nsName,
		TriggerType:              datastore.StartSimple,
		ListeningForEventTypes:   []string{},
		TriggerWorkflow:          file.ID.String(),
		Metadata:                 file.Path,
		LifespanOfReceivedEvents: int(lifespan.Milliseconds()),
		EventContextFilters:      []datastore.EventContextFilter{},
	}
	switch ms.Type {
	case "default":
		fEv.TriggerType = datastore.StartSimple
	case "event":
		fEv.TriggerType = datastore.StartSimple
	case "eventsXor":
		fEv.TriggerType = datastore.StartOR
	case "eventsAnd":
		fEv.TriggerType = datastore.StartAnd
	}
	contextFilters := make([]string, 0, len(ms.Events))
	eventTypesRemovedDuplicates := map[string]any{}
	for _, sed := range ms.Events {
		eventTypesRemovedDuplicates[sed.Type] = nil
		databaseNoDupCheck := ""
		filterContext := make(map[string]string)
		for k, v := range sed.Context {
			filterContext[k] = fmt.Sprintf("%v", v)
		}
		fEv.EventContextFilters = append(fEv.EventContextFilters, datastore.EventContextFilter{
			Type:    sed.Type,
			Context: filterContext,
		})
		for k, v := range sed.Context {
			databaseNoDupCheck += fmt.Sprintf("%v %v %v", sed.Type, k, v)
		}
		contextFilters = append(contextFilters, databaseNoDupCheck)
	}
	fEv.ListeningForEventTypes = make([]string, 0, len(eventTypesRemovedDuplicates))
	for k := range eventTypesRemovedDuplicates {
		fEv.ListeningForEventTypes = append(fEv.ListeningForEventTypes, k)
	}
	for i, j := 0, len(fEv.EventContextFilters)-1; i < j; i, j = i+1, j-1 {
		fEv.EventContextFilters[i], fEv.EventContextFilters[j] = fEv.EventContextFilters[j], fEv.EventContextFilters[i]
	}
	tx, err := tx.BeginTx(ctx)
	if err != nil {
		return err
	}
	err = tx.DataStore().EventListener().Append(ctx, fEv)
	if err != nil {
		return err
	}
	for i, t := range fEv.ListeningForEventTypes {
		err = tx.DataStore().EventListenerTopics().Append(ctx, nsID, nsName, fEv.ID, nsID.String()+"-"+t, contextFilters[i])
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// called from workflow instances to create event listeners.
func (events *events) addInstanceEventListener(ctx context.Context, namespace uuid.UUID, nsName string, instance uuid.UUID, sevents []*model.ConsumeEventDefinition, all bool) error {
	// var ev []map[string]interface{}

	fEv := &datastore.EventListener{
		ID:                     uuid.New(),
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
		Deleted:                false,
		NamespaceID:            namespace,
		Namespace:              nsName,
		TriggerType:            datastore.WaitSimple,
		ListeningForEventTypes: []string{},
		TriggerInstance:        instance.String(),
		// LifespanOfReceivedEvents: , TODO?
		EventContextFilters: []datastore.EventContextFilter{},
	}
	contextFilters := make([]string, 0, len(sevents))
	eventTypesRemovedDuplicates := map[string]any{}
	for _, ced := range sevents {
		eventTypesRemovedDuplicates[ced.Type] = nil
		filterContext := make(map[string]string)
		for k, v := range ced.Context {
			filterContext[k] = fmt.Sprintf("%v", v)
		}
		fEv.EventContextFilters = append(fEv.EventContextFilters, datastore.EventContextFilter{
			Type:    ced.Type,
			Context: filterContext,
		})
		for i, j := 0, len(fEv.EventContextFilters)-1; i < j; i, j = i+1, j-1 {
			fEv.EventContextFilters[i], fEv.EventContextFilters[j] = fEv.EventContextFilters[j], fEv.EventContextFilters[i]
		}
		databaseNoDupCheck := ""
		for k, v := range ced.Context {
			databaseNoDupCheck += fmt.Sprintf("%v %v %v", ced.Type, k, v)
		}
		contextFilters = append(contextFilters, databaseNoDupCheck)
	}
	fEv.ListeningForEventTypes = make([]string, 0, len(eventTypesRemovedDuplicates))
	for k := range eventTypesRemovedDuplicates {
		fEv.ListeningForEventTypes = append(fEv.ListeningForEventTypes, k)
	}
	if all {
		fEv.TriggerType = datastore.WaitAnd
	}
	if !all && len(fEv.ListeningForEventTypes) > 1 {
		fEv.TriggerType = datastore.WaitOR
	}

	err := events.runSQLTx(ctx, func(tx *database.DB) error {
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
