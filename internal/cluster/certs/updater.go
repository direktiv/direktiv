package certs

import (
	"log/slog"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/direktiv/direktiv/internal/core"
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
	go func() {
		slog.Info("run certificate loop")
		// for concurrent startup we delay it by up to ten seconds
		// if nodes startup we are trying to run them with a random delay
		delay := os.Getenv("DIREKTIV_NATS_CERT_DELAY")
		d, err := strconv.Atoi(delay)
		if err != nil {
			d = 10
		}

		time.Sleep(time.Duration(rand.Intn(d)) * time.Second) //nolint:gosec

		for {
			err := c.checkAndUpdate(circuit.Context())
			if err != nil {
				panic("can not refresh certificates")
			}

			sleepMinutes := rand.Intn(maxWait-minWait) + minWait //nolint:gosec
			slog.Info("sleeping for certificates", slog.Int("duration", sleepMinutes))
			time.Sleep(time.Duration(sleepMinutes) * time.Minute)
		}
	}()
}
