package cache_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/cache"
	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	natsTestContainer "github.com/testcontainers/testcontainers-go/modules/nats"
)

func TestCache(t *testing.T) {
	natsContainer, err := natsTestContainer.Run(context.Background(), "nats:2.9")
	if err != nil {
		t.Fatal(err)
	}

	cs, _ := natsContainer.ConnectionString(context.Background())
	nc, err := nats.Connect(cs)
	require.NoError(t, err)
	bus := pubsub.NewBus(nc)

	ctx, cancel := context.WithCancel(context.Background())

	circuit := core.NewCircuit(ctx, os.Interrupt)

	cache1, _ := cache.NewCache(bus)
	go cache1.Run(circuit)
	cache2, _ := cache.NewCache(bus)
	go cache2.Run(circuit)

	cache1.Set("hello", "world")
	cache2.Set("hello", "world2")

	// cache 1 needs to be unset
	require.Eventually(t, func() bool {
		_, b := cache1.Get("hello")
		return !b
	}, 3*time.Second, 100*time.Millisecond, "sync not received")
	cancel()

	// test shutdown of cache if context cancelled
	require.Eventually(t, func() bool {
		_, b := cache2.Get("hello")
		return !b
	}, 3*time.Second, 100*time.Millisecond, "shutdown not received")
}
