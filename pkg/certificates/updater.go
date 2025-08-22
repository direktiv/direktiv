package certificates

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/natsclient"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type CertificateUpdater struct {
	client *kubernetes.Clientset
	ns     string
}

const (
	maxWait = 120
	minWait = 45
)

func NewCertificateUpdater(ns string) (*CertificateUpdater, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &CertificateUpdater{
		client: client,
		ns:     ns,
	}, nil
}

func (c *CertificateUpdater) Start(circuit *core.Circuit) {
	go c.runCertLoop(circuit.Context())

	// waiting for nats to be available
	// this waits for certificates as well
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			slog.Info("checking nats connection")
			_, err := natsclient.NewNATSConnection()
			if err == nil {
				slog.Info("nats available")
				return
			}
			slog.Error("nats connection not available", slog.Any("error", err))
		case <-time.After(2 * time.Minute):
			// can not recover from nats not connecting
			panic("cannot connect to nats")
		}
	}
}

func (c *CertificateUpdater) runCertLoop(ctx context.Context) {
	slog.Info("run certificate loop")
	// for concurrent startup we delay it by up to ten seconds
	// if nodes startup we are trying to run them with a random delay
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second) //nolint:gosec

	for {
		err := c.checkAndUpdate(ctx)
		if err != nil {
			panic("can not refresh certificates")
		}

		sleepMinutes := rand.Intn(maxWait-minWait) + minWait //nolint:gosec
		slog.Info("sleeping for certificates", slog.Int("duration", sleepMinutes))
		time.Sleep(time.Duration(sleepMinutes) * time.Minute)
	}
}
