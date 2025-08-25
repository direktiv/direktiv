package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/engine"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type store struct {
	nc     *nats.Conn
	js     nats.JetStreamContext
	stream string
}

func NewStore(ctx context.Context, url, name string) (engine.Store, error) {
	nc, err := nats.Connect(url, nats.Name(name))
	if err != nil {
		return nil, fmt.Errorf("nats connect: %s", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		_ = nc.Drain()
		return nil, fmt.Errorf("nats jetstream: %s", err)
	}

	// res, err := js.AccountInfo()
	// fmt.Printf("jetstream info: res: %v, err: %s\n", res, err)

	_, err = js.AddStream(&nats.StreamConfig{
		Name: name,
		// engineMessages.<namespace>.instanceID.<instanceID>.type.<type>
		Subjects:    []string{"engineMessages.*.instanceID.*.type.*"},
		Storage:     nats.FileStorage,
		Retention:   nats.LimitsPolicy,
		MaxAge:      0,              // keep forever; set if you want TTL
		Duplicates:  24 * time.Hour, // enable publish de-dup
		Replicas:    1,              // bump if you're on a clustered server
		AllowDirect: true,           // speeds up direct gets (if you use them)
	})
	if err != nil {
		_ = nc.Drain()
		return nil, fmt.Errorf("nats add jetstream: %s", err)
	}

	return &store{nc: nc, js: js, stream: name}, nil
}

func (s *store) PutInstanceMessage(ctx context.Context, namespace string, instanceID uuid.UUID, typ string, payload any) (uuid.UUID, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return uuid.Nil, err
	}

	msgID := uuid.New()
	msg := engine.Message{
		Namespace: namespace,
		ID:        msgID.String(),
		Type:      typ,
		CreatedAt: time.Now(),
		Data:      payloadBytes,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return uuid.Nil, err
	}

	msgNats := &nats.Msg{
		Subject: fmt.Sprintf("engineMessages.%s.instanceID.%s.type.%s", namespace, instanceID, typ),
		Data:    msgBytes,
		Header:  nats.Header{},
	}
	msgNats.Header.Set("Nats-Msg-Id", fmt.Sprintf("%s:%s:%s:%s", namespace, instanceID, typ, msgID.String()))

	_, err = s.js.PublishMsg(msgNats, nats.Context(ctx))
	if err != nil {
		return msgID, fmt.Errorf("nats publish message: %s", err)
	}

	return msgID, nil
}

func (s *store) GetInstanceMessages(ctx context.Context, namespace string, instanceID uuid.UUID, typ string) ([]engine.Message, error) {
	subj := fmt.Sprintf("engineMessages.%s.instanceID.%s.type.%s", namespace, instanceID, typ)

	// Fetch from base (no eventType) and typed subjects, then merge by At.
	all, err := s.fetchAllFromSubject(ctx, subj)
	if err != nil {
		return nil, fmt.Errorf("fetch all from subject: %s", err)
	}

	return all, nil
}

func (s *store) fetchAllFromSubject(ctx context.Context, subj string) ([]engine.Message, error) {
	durable := fmt.Sprintf("consumer_%d", time.Now().UnixNano())
	cfg := &nats.ConsumerConfig{
		Durable:       durable,
		FilterSubject: subj,
		AckPolicy:     nats.AckNonePolicy,
		DeliverPolicy: nats.DeliverAllPolicy,
	}

	_, err := s.js.AddConsumer(s.stream, cfg)
	if err != nil {
		return nil, fmt.Errorf("add consumer: %s", err)
	}
	defer func() { _ = s.js.DeleteConsumer(s.stream, durable) }()

	sub, err := s.js.PullSubscribe(subj, durable, nats.Bind(s.stream, durable))
	if err != nil {
		return nil, fmt.Errorf("pull subscribe: %s", err)
	}

	batch := 100

	var out []engine.Message

	for {
		msgList, fetchErr := sub.Fetch(batch, nats.MaxWait(10*time.Millisecond))
		if fetchErr != nil && !errors.Is(fetchErr, nats.ErrTimeout) {
			return nil, fmt.Errorf("subscriber fetch: %s", err)
		}
		if len(msgList) == 0 {
			break
		}
		for _, m := range msgList {
			var msg engine.Message
			if err := json.Unmarshal(m.Data, &msg); err != nil {
				continue // skip bad payloads
			}
			out = append(out, msg)
		}
	}

	return out, nil
}
