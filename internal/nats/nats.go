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

var (
	StreamSchedRule = newDescriptor("sched.rule",
		&nats.StreamConfig{
			Storage:   nats.FileStorage,
			Retention: nats.LimitsPolicy,
			// MaxAge:     90 * 24 * time.Hour,
			Discard:    nats.DiscardOld,
			Duplicates: 48 * time.Hour,
			// important: keep only 1 message per subject (latest rule)
			MaxMsgsPerSubject: 1,
		}, nil)

	StreamSchedTask = newDescriptor("sched.task",
		&nats.StreamConfig{
			Storage:    nats.FileStorage,
			Retention:  nats.WorkQueuePolicy,
			Duplicates: 1 * time.Hour,
		}, nil)

	StreamEngineHistory = newDescriptor("engine.history",
		&nats.StreamConfig{
			Storage:   nats.FileStorage,
			Retention: nats.LimitsPolicy,
			MaxAge:    90 * 24 * time.Hour,
			Discard:   nats.DiscardOld,
			// set dupe window to protect idempotent publishes
			Duplicates: 48 * time.Hour,
		}, &nats.ConsumerConfig{
			AckPolicy:         nats.AckExplicitPolicy,
			AckWait:           30 * time.Second,
			MaxDeliver:        10,
			DeliverPolicy:     nats.DeliverAllPolicy,
			ReplayPolicy:      nats.ReplayInstantPolicy,
			InactiveThreshold: 72 * time.Hour,
		})

	StreamEngineStatus = newDescriptor("instance.status",
		&nats.StreamConfig{
			Storage:    nats.FileStorage,
			Retention:  nats.LimitsPolicy,
			MaxAge:     90 * 24 * time.Hour,
			Discard:    nats.DiscardOld,
			Duplicates: 48 * time.Hour,
			// important: keep only 1 message per subject (latest status)
			MaxMsgsPerSubject: 1,
		}, nil)

	StreamEngineQueue = newDescriptor("engine.queue",
		&nats.StreamConfig{
			Storage:    nats.FileStorage,
			Retention:  nats.WorkQueuePolicy,
			Duplicates: 1 * time.Hour,
		}, &nats.ConsumerConfig{
			AckPolicy:         nats.AckExplicitPolicy,
			AckWait:           5 * time.Minute,
			MaxDeliver:        10,
			DeliverPolicy:     nats.DeliverAllPolicy,
			ReplayPolicy:      nats.ReplayInstantPolicy,
			InactiveThreshold: 0, // means never auto-delete
		})
)

var allDescriptors = []*Descriptor{
	StreamEngineHistory,
	StreamEngineStatus,
	StreamSchedRule,
	StreamSchedTask,
	StreamEngineQueue,
}

type Conn = nats.Conn

func Connect() (*nats.Conn, error) {
	// set the deployment name in dns names
	deploymentName := os.Getenv("DIREKTIV_DEPLOYMENT_NAME")

	dirs := []string{
		"/etc/direktiv-tls",
		os.TempDir() + "/generated-direktiv-tls",
	}

	var dir string
	for _, d := range dirs {
		slog.Info("looking for tls files", "dir", d)
		if _, err := os.Stat(d + "/server.crt"); err == nil {
			dir = d
			break
		}
	}
	if dir == "" {
		return nil, errors.New("tls files don't exist")
	}

	return nats.Connect(
		fmt.Sprintf("tls://%s-nats.default.svc:4222", deploymentName),
		nats.ClientTLSConfig(
			func() (tls.Certificate, error) {
				cert, err := tls.LoadX509KeyPair(dir+"/server.crt",
					dir+"/server.key")
				if err != nil {
					slog.Error("cannot create certificate pair", slog.Any("error", err))
					return tls.Certificate{}, err
				}

				return cert, nil
			},
			func() (*x509.CertPool, error) {
				caCert, err := os.ReadFile(dir + "/ca.crt")
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

	// 1- ensure streams
	for _, dp := range allDescriptors {
		err = ensureStream(ctx, js, dp.streamConfig)
		if err != nil {
			return nil, fmt.Errorf("nats ensure stream %s: %w", dp, err)
		}
	}

	// 2- ensure shared durable consumers
	for _, dp := range allDescriptors {
		if dp.consumerConfig == nil {
			continue
		}
		err = ensureConsumer(ctx, js, dp.consumerConfig)
		if err != nil {
			return nil, fmt.Errorf("nats ensure consumer %s: %w", dp, err)
		}
	}

	return js, nil
}

// TODO: remove this debug code.
func ResetStreams(ctx context.Context, js nats.JetStreamContext) error {
	streams := js.StreamNames()
	for s := range streams {
		if err := js.DeleteStream(s); err != nil {
			return fmt.Errorf("nats delete stream %s: %w", s, err)
		}
		consumers := js.ConsumerNames(s)
		for c := range consumers {
			if err := js.DeleteConsumer(s, c); err != nil {
				return fmt.Errorf("nats delete consumer %s: %w", c, err)
			}
		}
	}

	return nil
}

func ensureStream(ctx context.Context, js nats.JetStreamContext, cfg *nats.StreamConfig) error {
	_, err := js.StreamInfo(cfg.Name, nats.Context(ctx))
	if err == nil {
		return nil
	}
	if !errors.Is(err, nats.ErrStreamNotFound) {
		return fmt.Errorf("nats info stream %s: %w", cfg.Name, err)
	}

	slog.Info("creating nats stream", slog.String("name", cfg.Name), slog.Any("subjects", cfg.Subjects))
	_, err = js.AddStream(cfg, nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("nats add stream %s: %w", cfg.Name, err)
	}

	return nil
}

func ensureConsumer(ctx context.Context, js nats.JetStreamContext, cfg *nats.ConsumerConfig) error {
	_, err := js.ConsumerInfo(cfg.Durable, cfg.Durable, nats.Context(ctx))
	if err == nil {
		return nil
	}
	if !errors.Is(err, nats.ErrConsumerNotFound) {
		return fmt.Errorf("nats info consumer %s: %w", cfg.Durable, err)
	}
	_, err = js.AddConsumer(cfg.Durable, cfg, nats.Context(ctx))
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
