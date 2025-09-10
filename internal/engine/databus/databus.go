package databus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/direktiv/direktiv/internal/engine"
	intNats "github.com/direktiv/direktiv/internal/nats"
	"github.com/direktiv/direktiv/pkg/lifecycle"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type DataBus struct {
	js    nats.JetStreamContext
	cache *StatusCache
}

func New(js nats.JetStreamContext) *DataBus {
	return &DataBus{js: js, cache: NewStatusCache()}
}

var _ engine.DataBus = &DataBus{}

func (d *DataBus) Start(lc *lifecycle.Manager) error {
	err := d.startStatusCache(lc.Context())
	if err != nil {
		return fmt.Errorf("start status cache: %w", err)
	}

	return nil
}

func (d *DataBus) PushInstanceEvent(ctx context.Context, event *engine.InstanceEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	subject := fmt.Sprintf(intNats.SubjInstanceHistory, event.Namespace, event.InstanceID)

	_, err = d.js.Publish(subject, data,
		nats.Context(ctx),
		nats.MsgId(fmt.Sprintf("instance::history::%s", event.EventID)))

	return err
}

func (d *DataBus) QueryInstanceStatus(ctx context.Context, filterNamespace string, filterInstanceID uuid.UUID) []engine.InstanceStatus {
	return d.cache.Snapshot(filterNamespace, filterInstanceID)
}

func (d *DataBus) startStatusCache(ctx context.Context) error {
	subj := fmt.Sprintf(intNats.SubjInstanceStatus, "*", "*")
	// ephemeral, AckNone (we don't want to disturb the stream/consumers)
	_, err := d.js.Subscribe(subj, func(msg *nats.Msg) {
		var st engine.InstanceStatus
		if err := json.Unmarshal(msg.Data, &st); err != nil {
			// best-effort; ignore bad payloads
			return
		}
		d.cache.Upsert(st)
	}, nats.AckNone())
	if err != nil {
		return err
	}

	return nil
}
