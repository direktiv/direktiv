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

type ClusterManager struct {
	client *kubernetes.Clientset
	ns     string
}

func NewClusterManager(ns string) (*ClusterManager, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &ClusterManager{
		client: client,
		ns:     ns,
	}, nil
}

func (c *ClusterManager) Start(circuit *core.Circuit) {
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

	}

	go func() {
		for {
			elector.Run(circuit.Context())
		}
	}()

	// go leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
	// 	Lock:            lock,
	// 	LeaseDuration:   15 * time.Second,
	// 	RenewDeadline:   10 * time.Second,
	// 	RetryPeriod:     2 * time.Second,
	// 	ReleaseOnCancel: true,
	// 	Callbacks: leaderelection.LeaderCallbacks{
	// 		OnStartedLeading: func(ctx context.Context) {
	// 			slog.Info("cluster leadership")
	// 			err := c.refreshCerts(ctx)
	// 			if err != nil {
	// 				slog.Error("error refreshing certificates", slog.Any("error", err))
	// 				cancel()
	// 			}
	// 			slog.Info("dropping cluster leadership")
	// 		},
	// 		OnStoppedLeading: func() {
	// 			slog.Info("cluster leadership lost")
	// 		},
	// 		OnNewLeader: func(identity string) {
	// 			slog.Info("new cluster leader", slog.String("identity", identity))
	// 		},
	// 	},
	// })
}

func (c *ClusterManager) runLeader(ctx context.Context) {
	podName := os.Getenv("POD_NAME")
	ctx, cancel := context.WithCancel(ctx)

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

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		LeaseDuration:   15 * time.Second,
		RenewDeadline:   10 * time.Second,
		RetryPeriod:     2 * time.Second,
		ReleaseOnCancel: true,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				slog.Info("cluster leadership")
				err := c.refreshCerts(ctx)
				if err != nil {
					slog.Error("error refreshing certificates", slog.Any("error", err))
					cancel()
				}
				slog.Info("dropping cluster leadership")
			},
			OnStoppedLeading: func() {
				slog.Info("cluster leadership lost")
			},
			OnNewLeader: func(identity string) {
				slog.Info("new cluster leader", slog.String("identity", identity))
			},
		},
	})
}
