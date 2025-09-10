package nats

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/direktiv/direktiv/internal/engine"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	ConsumerStatusMaterializer = "CONSUMER_INSTANCES_STATUS_MATERIALIZER"
	StreamInstanceHistory      = "STREAM_INSTANCE_HISTORY"
	StreamInstanceStatus       = "STREAM_INSTANCE_STATUS"

	SubjInstanceStatus  = "instance.status.%s.%s"  // instance.status.<namespace>.<instanceID>
	SubjInstanceHistory = "instance.history.%s.%s" // instance.history.<namespace>.<instanceID>

	StreamSchedRule = "STREAM_SCHED_RULE"
	SubjSchedRule   = "sched.rule.%s.%s" // shed.config.<namespace>.<ruleID>
)

type Conn = nats.Conn

func Connect() (*nats.Conn, error) {
	// set the deployment name in dns names
	deploymentName := os.Getenv("DIREKTIV_DEPLOYMENT_NAME")

	return nats.Connect(
		fmt.Sprintf("tls://%s-nats.default.svc:4222", deploymentName),
		nats.ClientTLSConfig(
			func() (tls.Certificate, error) {
				cert, err := tls.LoadX509KeyPair("/etc/direktiv-tls/server.crt",
					"/etc/direktiv-tls/server.key")
				if err != nil {
					slog.Error("cannot create certificate pair", slog.Any("error", err))
					return tls.Certificate{}, err
				}

				return cert, nil
			},
			func() (*x509.CertPool, error) {
				caCert, err := os.ReadFile("/etc/direktiv-tls/ca.crt")
				if err != nil {
					return nil, err
				}
				caPool := x509.NewCertPool()
				if !caPool.AppendCertsFromPEM(caCert) {
					slog.Error("cannot create certificate pair", slog.Any("error", err))
					return nil, err
				}

				return caPool, nil
			},
		),
	)
}

func SetupJetStream(ctx context.Context, nc *nats.Conn) (nats.JetStreamContext, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("nats jetstream: %w", err)
	}

	//err = resetStreams(ctx, js)
	//if err != nil {
	//	return nil, fmt.Errorf("nats reset streams: %w", err)
	//}

	// 1- ensure streams
	ensureStreams := []*nats.StreamConfig{
		{
			Name: StreamInstanceHistory,
			Subjects: []string{
				fmt.Sprintf(SubjInstanceHistory, "*", "*"),
			},
			Storage:   nats.FileStorage,
			Retention: nats.LimitsPolicy,
			MaxAge:    90 * 24 * time.Hour,
			Discard:   nats.DiscardOld,
			// set dupe window to protect idempotent publishes
			Duplicates: 48 * time.Hour,
		},
		{
			Name: StreamInstanceStatus,
			Subjects: []string{
				fmt.Sprintf(SubjInstanceStatus, "*", "*"),
			},
			Storage:    nats.FileStorage,
			Retention:  nats.LimitsPolicy,
			MaxAge:     90 * 24 * time.Hour,
			Discard:    nats.DiscardOld,
			Duplicates: 48 * time.Hour,
			// important: keep only 1 message per subject (latest status)
			MaxMsgsPerSubject: 1,
		},
		{
			Name: StreamSchedRule,
			Subjects: []string{
				fmt.Sprintf(SubjSchedRule, "*", "*"),
			},
			Storage:   nats.FileStorage,
			Retention: nats.LimitsPolicy,
			// MaxAge:     90 * 24 * time.Hour,
			Discard:    nats.DiscardOld,
			Duplicates: 48 * time.Hour,
			// important: keep only 1 message per subject (latest config)
			MaxMsgsPerSubject: 1,
		},
	}

	for _, cfg := range ensureStreams {
		err = ensureStream(ctx, js, cfg)
		if err != nil {
			return nil, fmt.Errorf("nats ensure stream %s: %w", cfg.Name, err)
		}
	}

	// 2- ensure shared durable consumers
	ensureConsumers := []*nats.ConsumerConfig{
		{
			Durable:           ConsumerStatusMaterializer,
			FilterSubject:     fmt.Sprintf(SubjInstanceHistory, "*", "*"),
			AckPolicy:         nats.AckExplicitPolicy,
			AckWait:           30 * time.Second,
			MaxDeliver:        10,
			DeliverPolicy:     nats.DeliverAllPolicy,
			ReplayPolicy:      nats.ReplayInstantPolicy,
			InactiveThreshold: 72 * time.Hour,
			Metadata: map[string]string{
				"stream": StreamInstanceHistory,
			},
		},
	}

	for _, cfg := range ensureConsumers {
		err = ensureConsumer(ctx, js, cfg)
		if err != nil {
			return nil, fmt.Errorf("nats ensure consumer %s: %w", cfg.Durable, err)
		}
	}

	return js, nil
}

func generateRandomEntries(ctx context.Context, js nats.JetStreamContext) error {
	namespaces := []string{"ns1", "ns2", "ns3", "ns4"}
	types := []string{"running", "failed", "succeeded"}
	instIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}

	for range 100 {
		evID := uuid.New()
		typ := types[rand.Intn(len(types))]
		ns := namespaces[rand.Intn(len(namespaces))]
		instID := instIDs[rand.Intn(len(instIDs))]

		subject := fmt.Sprintf(SubjInstanceHistory, ns, instID)

		ev := engine.InstanceEvent{
			EventID:    evID,
			InstanceID: instID,
			Namespace:  ns,
			Type:       typ,
			Time:       time.Now(),
			Script:     "",
		}
		data, _ := json.Marshal(ev)

		// Publish with a dedupe Msg-Id
		_, err := js.Publish(subject, data,
			nats.MsgId(fmt.Sprintf("instance::history::%s", evID)))
		if err != nil {
			return err
		}
		fmt.Printf("published >>>>> %s\n", data)
	}

	return nil
}

// TODO: remove this debug code.
func resetStreams(ctx context.Context, js nats.JetStreamContext) error {
	// List all streams
	streams := js.StreamNames()
	for s := range streams {
		if err := js.DeleteStream(s); err != nil {
			return fmt.Errorf("nats delete stream %s: %w", s, err)
		}
	}

	return nil
}

func ensureStream(ctx context.Context, js nats.JetStreamContext, cfg *nats.StreamConfig) error {
	_, err := js.StreamInfo(cfg.Name)
	if err == nil {
		return nil
	}
	if !errors.Is(err, nats.ErrStreamNotFound) {
		return fmt.Errorf("nats info stream %s: %w", cfg.Name, err)
	}
	_, err = js.AddStream(cfg)
	if err != nil {
		return fmt.Errorf("nats add stream %s: %w", cfg.Name, err)
	}

	return nil
}

func ensureConsumer(ctx context.Context, js nats.JetStreamContext, cfg *nats.ConsumerConfig) error {
	_, err := js.ConsumerInfo(cfg.Metadata["stream"], cfg.Durable)
	if err == nil {
		return nil
	}
	if !errors.Is(err, nats.ErrConsumerNotFound) {
		return fmt.Errorf("nats info consumer %s: %w", cfg.Durable, err)
	}
	_, err = js.AddConsumer(cfg.Metadata["stream"], cfg)
	if err != nil {
		return fmt.Errorf("nats add consumer %s: %w", cfg.Durable, err)
	}

	return nil
}
