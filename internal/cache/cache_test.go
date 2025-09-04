package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/direktiv/direktiv/internal/cache"
	natspubsub "github.com/direktiv/direktiv/internal/cluster/pubsub/nats"
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
	defer nc.Drain()
	require.NoError(t, err)
	buss := natspubsub.New(nc, nil)

	cache1, _ := cache.New(buss, "host1", false, nil)
	defer cache1.Close()
	cache2, _ := cache.New(buss, "host2", false, nil)
	defer cache2.Close()

	cache1.Set("hello", "world")
	cache1.Set("foo", "bar")

	cache2.Set("hello", "world")
	cache2.Set("foo", "bar")

	for _, c := range []*cache.Cache{
		cache1, cache2,
	} {
		require.Eventually(t, func() bool {
			v, ok := c.Get("hello")
			return ok && v.(string) == "world"
		}, 3*time.Second, 100*time.Millisecond, "test get key 'hello'")

		require.Eventually(t, func() bool {
			v, ok := c.Get("foo")
			return ok && v.(string) == "bar"
		}, 3*time.Second, 100*time.Millisecond, "test get key 'foo'")
	}

	// test cluster cache invalidation.
	cache1.Delete("hello")
	cache2.Delete("foo")

	for _, c := range []*cache.Cache{
		cache1, cache2,
	} {
		require.Eventually(t, func() bool {
			_, ok := c.Get("hello")
			return !ok
		}, 3*time.Second, 100*time.Millisecond, "test delete key 'hello'")

		require.Eventually(t, func() bool {
			_, ok := c.Get("foo")
			return !ok
		}, 3*time.Second, 100*time.Millisecond, "test delete key 'foo'")
	}
}
