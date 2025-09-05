package certs

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/direktiv/direktiv/pkg/lifecycle"
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

func (c *CertificateUpdater) Run(lc *lifecycle.Manager) error {
	slog.Info("run certificate loop")
	// for concurrent startup we delay it by up to ten seconds
	// if nodes startup we are trying to run them with a random delay
	delay := os.Getenv("DIREKTIV_NATS_CERT_DELAY")
	d, err := strconv.Atoi(delay)
	if err != nil {
		d = 10
	}
	sleepMinutes := rand.Intn(maxWait-minWait) + minWait //nolint:gosec

	initialTicker := time.NewTicker(time.Duration(d) * time.Second)
	ticker := time.NewTicker(time.Duration(sleepMinutes) * time.Minute)

	slog.Info("ticking time for certificates", slog.Int("minutes", sleepMinutes))

	for {
		select {
		case <-lc.Done():
			return nil
		case <-initialTicker.C:
			// run only once
			initialTicker.Stop()
			err := c.checkAndUpdate(lc.Context())
			if err != nil {
				return fmt.Errorf("certificate checkAndUpdate, err: %w", err)
			}
		case <-ticker.C:
			err := c.checkAndUpdate(lc.Context())
			if err != nil {
				return fmt.Errorf("certificate checkAndUpdate, err: %w", err)
			}
		}
	}
}
