package databus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/direktiv/direktiv/internal/api/filter"
	"github.com/direktiv/direktiv/internal/engine"
	intNats "github.com/direktiv/direktiv/internal/nats"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type DataBus struct {
	js           nats.JetStreamContext
	statusCache  *StatusCache
	historyCache *HistoryCache

	notifier *instanceNotifier
}

func New(js nats.JetStreamContext) *DataBus {
	return &DataBus{
		js:           js,
		statusCache:  NewStatusCache(),
		historyCache: NewHistoryCache(),
		notifier:     newInstanceNotifier(),
	}
}

var _ engine.DataBus = &DataBus{}

func (d *DataBus) Start(lc *lifecycle.Manager) error {
	err := d.startCaches(lc.Context())
	if err != nil {
		return fmt.Errorf("start caches: %w", err)
	}
	p := &projector{d.js}
	err = p.start(lc)
	if err != nil {
		return fmt.Errorf("start projector: %w", err)
	}

	return nil
}

func (d *DataBus) PublishInstanceHistoryEvent(ctx context.Context, event *engine.InstanceEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	subject := intNats.StreamEngineHistory.Subject(event.Namespace, event.InstanceID.String())

	_, err = d.js.Publish(subject, data,
		nats.Context(ctx),
		nats.MsgId(fmt.Sprintf("engine::history::%s", event.EventID)))

	return err
}

func (d *DataBus) PublishInstanceQueueEvent(ctx context.Context, event *engine.InstanceEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	subject := intNats.StreamEngineQueue.Subject(event.Namespace, event.InstanceID.String())

	_, err = d.js.Publish(subject, data,
		nats.Context(ctx),
		nats.MsgId(fmt.Sprintf("engine::queue::%s", event.EventID)))

	return err
}

func (d *DataBus) ListInstanceStatuses(ctx context.Context, filterNamespace string, filterInstanceID uuid.UUID, limit int, offset int, filters filter.Filters) ([]*engine.InstanceStatus, int) {
	return d.statusCache.SnapshotPage(filterNamespace, filterInstanceID, limit, offset)
}

func (d *DataBus) DeleteNamespace(ctx context.Context, name string) error {
	dpList := []*intNats.Descriptor{
		intNats.StreamEngineHistory,
		intNats.StreamEngineStatus,
		intNats.StreamEngineQueue,
	}

	for _, dp := range dpList {
		err := d.js.PurgeStream(
			dp.String(),
			&nats.StreamPurgeRequest{Subject: dp.Subject(name, "*")},
			nats.Context(ctx),
		)
		if err != nil {
			return fmt.Errorf("nats purge stream %s: %w", dp, err)
		}
	}
	d.statusCache.DeleteNamespace(name)
	d.historyCache.DeleteNamespace(name)

	return nil
}

func (d *DataBus) NotifyInstanceStatus(ctx context.Context, instanceID uuid.UUID, done chan<- *engine.InstanceStatus) {
	d.notifier.Add(instanceID, done)
}

func (d *DataBus) startCaches(ctx context.Context) error {
	// 1- start the status cache subscriber
	subj := intNats.StreamEngineStatus.Subject("*", "*")
	_, err := d.js.Subscribe(subj, func(msg *nats.Msg) {
		var st engine.InstanceStatus
		if err := json.Unmarshal(msg.Data, &st); err != nil {
			// best-effort; ignore bad payloads
			// TODO: log this
			return
		}

		if st.IsEndStatus() && st.Metadata[engine.LabelWithNotify] == "yes" {
			d.notifier.Notify(st.InstanceID, &st)
		}

		d.statusCache.Upsert(&st)
	}, nats.AckNone())
	if err != nil {
		return fmt.Errorf("start status cache subscriber: %w", err)
	}

	// 2- start the history cache subscriber
	subj = intNats.StreamEngineHistory.Subject("*", "*")
	_, err = d.js.Subscribe(subj, func(msg *nats.Msg) {
		var ev engine.InstanceEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			// best-effort; ignore bad payloads
			// TODO: log this
			return
		}
		metadata, err := msg.Metadata()
		if err != nil {
			// best-effort; ignore bad payloads
			// TODO: log this
			return
		}
		ev.Sequence = metadata.Sequence.Stream
		d.historyCache.Insert(&ev)
	}, nats.AckNone())
	if err != nil {
		return fmt.Errorf("start history cache subscriber: %w", err)
	}

	return nil
}

func (d *DataBus) GetInstanceHistory(ctx context.Context, namespace string, instanceID uuid.UUID) []*engine.InstanceEvent {
	return d.historyCache.Snapshot(namespace, instanceID)
}
