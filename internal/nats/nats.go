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
	StreamInstancesHistory     = "STREAM_INSTANCES_HISTORY"
	StreamInstancesStatus      = "STREAM_INSTANCES_STATUS"

	SubjInstanceStatus  = "instances.status.%s.%s"
	SubjInstanceHistory = "instances.history.%s.%s"
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

	// 1- ensure history stream streamInstancesHistory exists
	_, err = js.StreamInfo(StreamInstancesHistory)
	if err != nil && !errors.Is(err, nats.ErrStreamNotFound) {
		return nil, fmt.Errorf("info info stream %s: %w", StreamInstancesHistory, err)
	}
	if err != nil {
		_, err = js.AddStream(&nats.StreamConfig{
			Name: StreamInstancesHistory,
			Subjects: []string{
				fmt.Sprintf(SubjInstanceHistory, "*", "*"),
			},
			Storage:   nats.FileStorage,
			Retention: nats.LimitsPolicy,
			MaxAge:    90 * 24 * time.Hour,
			Discard:   nats.DiscardOld,
			// set dupe window to protect idempotent publishes
			Duplicates: 48 * time.Hour,
		})
		if err != nil {
			return nil, fmt.Errorf("nats add stream %s: %w", StreamInstancesHistory, err)
		}
	}

	// 2- ensure status stream (streamInstancesStatus) exists
	_, err = js.StreamInfo(StreamInstancesStatus)
	if err != nil && !errors.Is(err, nats.ErrStreamNotFound) {
		return nil, fmt.Errorf("nats info stream %s: %w", StreamInstancesStatus, err)
	}
	if err != nil {
		_, err = js.AddStream(&nats.StreamConfig{
			Name: StreamInstancesStatus,
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
		})
		if err != nil {
			return nil, fmt.Errorf("nats add stream %s: %w", StreamInstancesStatus, err)
		}
	}

	// 3- Ensure shared durable consumer on streamInstancesHistory
	_, err = js.ConsumerInfo(StreamInstancesHistory, ConsumerStatusMaterializer)
	if err != nil && !errors.Is(err, nats.ErrConsumerNotFound) {
		return nil, fmt.Errorf("nats consumer info %s: %w", ConsumerStatusMaterializer, err)
	}
	if err != nil {
		_, err = js.AddConsumer(StreamInstancesHistory, &nats.ConsumerConfig{
			Durable:           ConsumerStatusMaterializer,
			FilterSubject:     fmt.Sprintf(SubjInstanceHistory, "*", "*"),
			AckPolicy:         nats.AckExplicitPolicy,
			AckWait:           30 * time.Second,
			MaxDeliver:        10,
			DeliverPolicy:     nats.DeliverAllPolicy,
			ReplayPolicy:      nats.ReplayInstantPolicy,
			InactiveThreshold: 72 * time.Hour,
		})
		if err != nil {
			return nil, fmt.Errorf("nats add consumer %s: %w", ConsumerStatusMaterializer, err)
		}
	}

	err = generateRandomEntries(context.Background(), js)
	if err != nil {
		return nil, fmt.Errorf("nats gen entries: %w", err)
	}

	return js, nil
}

func generateRandomEntries(ctx context.Context, js nats.JetStreamContext) error {
	var namespaces = []string{"ns1", "ns2", "ns3", "ns4"}
	var types = []string{"running", "failed", "succeeded"}
	instIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}

	for i := 0; i < 100; i++ {
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
			nats.MsgId(fmt.Sprintf("hist::%s", evID)))
		if err != nil {
			return err
		}
		fmt.Printf("published >>>>> %s\n", data)
	}

	return nil
}
