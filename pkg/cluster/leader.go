package cluster

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

type CertificateManager struct {
	client     *kubernetes.Clientset
	ns         string
	certTicker *time.Ticker
}

func NewClusterManager(ns string) (*CertificateManager, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &CertificateManager{
		client: client,
		ns:     ns,
	}, nil
}

func (c *CertificateManager) Start(circuit *core.Circuit) {
	go c.runLeader(circuit.Context())

	// waiting for nats to be available
	// this waits for certificates as well
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			slog.Info("checking nats connection")
			_, err := natsConnect()
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

func (c *CertificateManager) runLeader(ctx context.Context) {
	for {
		podName := os.Getenv("POD_NAME")
		lock := &resourcelock.LeaseLock{
			LeaseMeta: metav1.ObjectMeta{
				Name:      "direktiv-leader-lock",
				Namespace: c.ns,
			},
			Client: c.client.CoordinationV1(),
			LockConfig: resourcelock.ResourceLockConfig{
				Identity: podName,
			},
		}

		ctxCancel, cancel := context.WithCancel(ctx)

		elector, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
			Lock:          lock,
			LeaseDuration: 15 * time.Second,
			RenewDeadline: 10 * time.Second,
			RetryPeriod:   2 * time.Second,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(ctx context.Context) {
					slog.Info("cluster leadership")
					err := c.refreshCerts(ctx)
					if err != nil {
						slog.Error("error refreshing certificates", slog.Any("error", err))
					}
					slog.Info("dropping cluster leadership")

					// stop leadership run
					cancel()
				},
				OnStoppedLeading: func() {
					slog.Info("cluster leadership lost")
				},
				OnNewLeader: func(identity string) {
					slog.Info("new cluster leader", slog.String("identity", identity))
				},
			},
		})
		if err != nil {
			panic("cannot run leadership config")
		}

		elector.Run(ctxCancel)
		time.Sleep(1 * time.Second)
	}
}
