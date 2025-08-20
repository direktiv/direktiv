package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type Store struct {
	nc     *nats.Conn
	js     nats.JetStreamContext
	stream string
}

func NewStore(ctx context.Context, url, name string) (*Store, error) {
	nc, err := nats.Connect(url, nats.Name(name))
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		_ = nc.Drain()
		return nil, err
	}

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
		if _, infoErr := js.StreamInfo(name); infoErr != nil {
			_ = nc.Drain()
			return nil, err
		}
	}

	return &Store{nc: nc, js: js, stream: name}, nil
}

type Message struct {
	Namespace string    `json:"namespace"`
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`

	Data json.RawMessage `json:"data"`
}

func (s *Store) PutMessage(ctx context.Context, namespace, instanceID, typ string, payload any) (uuid.UUID, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return uuid.Nil, err
	}

	msgID := uuid.New()
	msg := Message{
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

	return msgID, err
}

func (s *Store) QueryByInstance(ctx context.Context, namespace, instanceID uuid.UUID, typ string) ([]Message, error) {
	subj := fmt.Sprintf("engineMessages.%s.instanceID.%s.type.%s", namespace, instanceID, typ)

	// Fetch from base (no eventType) and typed subjects, then merge by At.
	all, err := s.fetchAllFromSubject(ctx, subj)
	if err != nil {
		return nil, err
	}

	return all, nil
}

func (s *Store) fetchAllFromSubject(ctx context.Context, subj string) ([]Message, error) {
	durable := fmt.Sprintf("consumer_%d", time.Now().UnixNano())
	cfg := &nats.ConsumerConfig{
		Durable:       durable,
		FilterSubject: subj,
		AckPolicy:     nats.AckNonePolicy,
		DeliverPolicy: nats.DeliverAllPolicy,
	}

	_, err := s.js.AddConsumer(s.stream, cfg)
	if err != nil {
		return nil, err
	}
	defer func() { _ = s.js.DeleteConsumer(s.stream, durable) }()

	sub, err := s.js.PullSubscribe(subj, durable, nats.Bind(s.stream, durable))
	if err != nil {
		return nil, err
	}

	batch := 100
	wait := 400 * time.Millisecond

	var out []Message

	for {
		msgList, fetchErr := sub.Fetch(batch, nats.MaxWait(wait), nats.Context(ctx))
		if fetchErr != nil && !errors.Is(fetchErr, nats.ErrTimeout) {
			return nil, fetchErr
		}
		if len(msgList) == 0 {
			break
		}
		for _, m := range msgList {
			var msg Message
			if err := json.Unmarshal(m.Data, &msg); err != nil {
				continue // skip bad payloads
			}
			out = append(out, msg)
		}
	}

	return out, nil
}
