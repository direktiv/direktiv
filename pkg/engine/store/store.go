package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/engine"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type store struct {
	nc     *nats.Conn
	js     nats.JetStreamContext
	stream string
}

const (
	natsClientName          = "engine_messages_store"
	instanceMessagesSubject = "instanceMessages.%s.instanceID.%s.type.%s"
)

func NewStore(ctx context.Context, nc *nats.Conn) (engine.Store, error) {
	js, err := nc.JetStream()
	if err != nil {
		_ = nc.Drain()
		return nil, fmt.Errorf("nats jetstream: %w", err)
	}

	// res, err := js.AccountInfo()
	// fmt.Printf("jetstream info: res: %v, err: %s\n", res, err)

	_, err = js.AddStream(&nats.StreamConfig{
		Name: natsClientName,
		// instanceMessages.<namespace>.instanceID.<instanceID>.type.<type>
		Subjects:    []string{fmt.Sprintf(instanceMessagesSubject, "*", "*", "*")},
		Storage:     nats.FileStorage,
		Retention:   nats.LimitsPolicy,
		MaxAge:      0,              // keep forever; set if you want TTL
		Duplicates:  24 * time.Hour, // enable publish de-dup
		Replicas:    1,              // bump if you're on a clustered server
		AllowDirect: true,           // speeds up direct gets (if you use them)
	})
	if err != nil {
		_ = nc.Drain()
		return nil, fmt.Errorf("nats add jetstream: %w", err)
	}

	return &store{nc: nc, js: js, stream: natsClientName}, nil
}

func (s *store) PushInstanceMessage(ctx context.Context, namespace string, instanceID uuid.UUID, typ string, payload any) (uuid.UUID, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshalling payload: %w", err)
	}

	msgID := uuid.New()
	msg := core.EngineMessage{
		Namespace: namespace,
		ID:        msgID.String(),
		Type:      typ,
		CreatedAt: time.Now(),
		Data:      payloadBytes,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return uuid.Nil, fmt.Errorf("marshalling engine message: %w", err)
	}

	msgNats := &nats.Msg{
		Subject: fmt.Sprintf(instanceMessagesSubject, namespace, instanceID, typ),
		Data:    msgBytes,
		Header:  nats.Header{},
	}
	msgNats.Header.Set("Nats-Msg-Id", fmt.Sprintf("%s:%s:%s:%s", namespace, instanceID, typ, msgID))

	_, err = s.js.PublishMsg(msgNats, nats.Context(ctx))
	if err != nil {
		return msgID, fmt.Errorf("nats publish message: %w", err)
	}

	return msgID, nil
}

func (s *store) PullInstanceMessages(ctx context.Context, namespace string, instanceID uuid.UUID, typ string) ([]core.EngineMessage, error) {
	subj := fmt.Sprintf(instanceMessagesSubject, namespace, instanceID, typ)

	all, err := s.pullFromSubject(ctx, subj)
	if err != nil {
		return nil, fmt.Errorf("pull from subject: %w", err)
	}

	return all, nil
}

func (s *store) pullFromSubject(ctx context.Context, subj string) ([]core.EngineMessage, error) {
	durable := fmt.Sprintf("consumer_%d", time.Now().UnixNano())
	cfg := &nats.ConsumerConfig{
		Durable:       durable,
		FilterSubject: subj,
		AckPolicy:     nats.AckNonePolicy,
		DeliverPolicy: nats.DeliverAllPolicy,
	}

	_, err := s.js.AddConsumer(s.stream, cfg)
	if err != nil {
		return nil, fmt.Errorf("add consumer: %w", err)
	}
	defer func() { _ = s.js.DeleteConsumer(s.stream, durable) }()

	sub, err := s.js.PullSubscribe(subj, durable, nats.Bind(s.stream, durable))
	if err != nil {
		return nil, fmt.Errorf("pull subscribe: %w", err)
	}

	batch := 100

	var out []core.EngineMessage

	for {
		msgList, fetchErr := sub.Fetch(batch, nats.MaxWait(10*time.Millisecond))
		if fetchErr != nil && !errors.Is(fetchErr, nats.ErrTimeout) {
			return nil, fmt.Errorf("subscriber fetch: %w", err)
		}
		if len(msgList) == 0 {
			break
		}
		for _, m := range msgList {
			var msg core.EngineMessage
			if err := json.Unmarshal(m.Data, &msg); err != nil {
				continue // skip bad payloads
			}
			out = append(out, msg)
		}
	}

	return out, nil
}
