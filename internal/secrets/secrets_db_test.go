package secrets_test

import (
	"context"
	"testing"
	"time"

	"github.com/direktiv/direktiv/internal/cache"
	"github.com/direktiv/direktiv/internal/cluster/pubsub"
	natspubsub "github.com/direktiv/direktiv/internal/cluster/pubsub/nats"
	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/secrets"
	database2 "github.com/direktiv/direktiv/pkg/database"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	natsTestContainer "github.com/testcontainers/testcontainers-go/modules/nats"
	"gorm.io/gorm"
)

func TestDBSecrets(t *testing.T) {

	// create database
	ns := uuid.NewString()
	conn, err := database2.NewTestDBWithNamespace(t, ns)
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
	buss := natspubsub.New(nc, nil)

	sh1, cache1 := buildSecrets(conn, buss, "host1")
	defer cache1.Close()
	sh2, cache2 := buildSecrets(conn, buss, "host2")
	defer cache2.Close()

	sec1, _ := sh1.SecretsForNamespace(ctx, ns)
	sec2, _ := sh2.SecretsForNamespace(ctx, ns)

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

func buildSecrets(db *gorm.DB, bus pubsub.EventBus, host string) (core.SecretsManager, core.Cache) {
	cache, _ := cache.New(bus, host, true, nil)

	return secrets.NewManager(db, cache), cache
}
