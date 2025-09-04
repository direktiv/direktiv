package cache_test

import (
	"context"
	"fmt"
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
	cache2.Set("foo", "bar")

	for _, c := range []*cache.Cache{
		cache1, cache2,
	} {
		require.Eventually(t, func() bool {
			v, ok := c.Get("hello")

			fmt.Printf("v=%v  v=%T  ok:%v\n", v, v, ok)

			return ok && v.(string) == "world"
		}, 3*time.Second, 100*time.Millisecond, "test get key 'hello'")

		//require.Eventually(t, func() bool {
		//	v, ok := c.Get("foo")
		//	return ok && v.(string) == "bar"
		//}, 3*time.Second, 100*time.Millisecond, "test get key 'foo'")

	}
}
