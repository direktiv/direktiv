package secrets_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/direktiv/direktiv/internal/testutils"
	"github.com/direktiv/direktiv/pkg/cache"
	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/pubsub"
	"github.com/direktiv/direktiv/pkg/secrets"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	natsTestContainer "github.com/testcontainers/testcontainers-go/modules/nats"
)

func TestDBSecrets(t *testing.T) {

	// create database
	db, ns, err := testutils.NewTestDBWithNamespace(t, uuid.NewString())
	if err != nil {
		t.Fatalf("unepxected NewTestDBWithNamespace() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// create bus
	natsContainer, err := natsTestContainer.Run(context.Background(), "nats:2.9")
	if err != nil {
		t.Fatal(err)
	}
	cs, _ := natsContainer.ConnectionString(context.Background())
	nc, err := nats.Connect(cs)
	require.NoError(t, err)
	buss := pubsub.NewPubSub(nc)

	sh1, cache1 := buildSecrets(ctx, db, buss, "host1")
	sh2, cache2 := buildSecrets(ctx, db, buss, "host2")

	sec1, _ := sh1.SecretsForNamespace(ctx, ns.Name)
	sec2, _ := sh2.SecretsForNamespace(ctx, ns.Name)

	// set on one
	sec1.Set(ctx, &core.Secret{
		Name: "hello",
		Data: []byte("world"),
	})

	for i := 0; i < 5; i++ {
		sec1.Get(context.Background(), "hello")
	}
	require.Equal(t, uint64(5), cache1.Hits())

	sec2.Get(context.Background(), "hello")
	sec2.Get(context.Background(), "hello")

	require.Equal(t, uint64(1), cache2.Hits())
	require.Equal(t, uint64(1), cache2.Misses())

	// set on one
	sec1.Update(ctx, &core.Secret{
		Name: "hello",
		Data: []byte("world2"),
	})

	v, _ := sec1.Get(context.Background(), "hello")
	require.Equal(t, string(v.Data), "world2")

	require.Eventually(t, func() bool {
		v, _ := sec2.Get(context.Background(), "hello")
		return string(v.Data) == "world2"
	}, 3*time.Second, 100*time.Millisecond, "value not received")

	require.GreaterOrEqual(t, uint64(2), cache2.Misses())

	list, _ := sec1.GetAll(ctx)
	require.Equal(t, 1, len(list))

	// deleting on one should remove it from the two cache
	sec1.Delete(ctx, "hello")

	del1, _ := sec1.Get(context.Background(), "hello")
	require.Nil(t, del1)

	// cache shoule be emptied after a while
	require.Eventually(t, func() bool {
		del2, _ := sec2.Get(context.Background(), "hello")
		return del2 == nil
	}, 3*time.Second, 100*time.Millisecond, "value not received")

	cancel()

}

func buildSecrets(ctx context.Context, db *database.DB, bus core.PubSub, host string) (core.SecretsManager, core.Cache) {
	circuit := core.NewCircuit(ctx, os.Interrupt)
	cache, _ := cache.NewCache(bus, host, true)
	go cache.Run(circuit)

	return secrets.NewManager(db, cache), cache
}
