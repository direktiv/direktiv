package nats

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	natsContainer "github.com/testcontainers/testcontainers-go/modules/nats"
)

var StreamInstanceStatus = StreamDescriptor("instance.status")
var StreamInstanceHistory = StreamDescriptor("instance.history")
var StreamSchedRule = StreamDescriptor("sched.rule")
var StreamSchedTask = StreamDescriptor("sched.task")
var StreamEngineQueue = StreamDescriptor("engine.queue")

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

	err = resetStreams(ctx, js)
	if err != nil {
		return nil, fmt.Errorf("nats reset streams: %w", err)
	}

	// 1- ensure streams
	ensureStreams := []*nats.StreamConfig{
		{
			Name: StreamInstanceHistory.String(),
			Subjects: []string{
				StreamInstanceHistory.Subject("*", "*"),
			},
			Storage:   nats.FileStorage,
			Retention: nats.LimitsPolicy,
			MaxAge:    90 * 24 * time.Hour,
			Discard:   nats.DiscardOld,
			// set dupe window to protect idempotent publishes
			Duplicates: 48 * time.Hour,
		},
		{
			Name: StreamInstanceStatus.String(),
			Subjects: []string{
				StreamInstanceStatus.Subject("*", "*"),
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
			Name: StreamSchedRule.String(),
			Subjects: []string{
				StreamSchedRule.Subject("*", "*"),
			},
			Storage:   nats.FileStorage,
			Retention: nats.LimitsPolicy,
			// MaxAge:     90 * 24 * time.Hour,
			Discard:    nats.DiscardOld,
			Duplicates: 48 * time.Hour,
			// important: keep only 1 message per subject (latest rule)
			MaxMsgsPerSubject: 1,
		},
		{
			Name: StreamSchedTask.String(),
			Subjects: []string{
				StreamSchedTask.Subject("*", "*"),
			},
			Storage:    nats.FileStorage,
			Retention:  nats.WorkQueuePolicy,
			Duplicates: 1 * time.Hour,
		},
		{
			Name: StreamEngineQueue.String(),
			Subjects: []string{
				StreamEngineQueue.Subject("*", "*"),
			},
			Storage:    nats.FileStorage,
			Retention:  nats.WorkQueuePolicy,
			Duplicates: 1 * time.Hour,
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
			Durable:           StreamInstanceHistory.String(),
			FilterSubject:     StreamInstanceHistory.Subject("*", "*"),
			AckPolicy:         nats.AckExplicitPolicy,
			AckWait:           30 * time.Second,
			MaxDeliver:        10,
			DeliverPolicy:     nats.DeliverAllPolicy,
			ReplayPolicy:      nats.ReplayInstantPolicy,
			InactiveThreshold: 72 * time.Hour,
		},
		{
			Durable:           StreamEngineQueue.String(),
			FilterSubject:     StreamEngineQueue.Subject("*", "*"),
			AckPolicy:         nats.AckExplicitPolicy,
			AckWait:           5 * time.Minute,
			MaxDeliver:        10,
			DeliverPolicy:     nats.DeliverAllPolicy,
			ReplayPolicy:      nats.ReplayInstantPolicy,
			InactiveThreshold: 0, // means never auto-delete
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
	_, err := js.ConsumerInfo(cfg.Durable, cfg.Durable)
	if err == nil {
		return nil
	}
	if !errors.Is(err, nats.ErrConsumerNotFound) {
		return fmt.Errorf("nats info consumer %s: %w", cfg.Durable, err)
	}
	_, err = js.AddConsumer(cfg.Durable, cfg)
	if err != nil {
		return fmt.Errorf("nats add consumer %s: %w", cfg.Durable, err)
	}

	return nil
}

func NewTestNats(t *testing.T) (string, error) {
	t.Helper()
	ctx := context.Background()

	ctr, err := natsContainer.Run(
		ctx,
		"nats:2.10-alpine",
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := ctr.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// Get nats://<host>:<port>
	uri, err := ctr.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	return uri, nil
}
