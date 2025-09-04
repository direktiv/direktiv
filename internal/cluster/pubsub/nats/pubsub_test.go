package nats_test

import (
	"context"
	"testing"
	"time"

	"github.com/direktiv/direktiv/internal/cluster/pubsub"
	natspubsub "github.com/direktiv/direktiv/internal/cluster/pubsub/nats"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	natsTestContainer "github.com/testcontainers/testcontainers-go/modules/nats"
)

func TestPubSub(t *testing.T) {
	natsContainer, err := natsTestContainer.Run(context.Background(), "nats:2.9")
	if err != nil {
		t.Fatal(err)
	}

	cs, _ := natsContainer.ConnectionString(context.Background())
	busPublish, err := natspubsub.New(func() (*nats.Conn, error) {
		return nats.Connect(cs)
	}, nil)
	require.NoError(t, err)
	defer busPublish.Close()

	busReceive, err := natspubsub.New(func() (*nats.Conn, error) {
		return nats.Connect(cs)
	}, nil)
	require.NoError(t, err)
	defer busReceive.Close()

	dataSend := []byte("test data")
	var dataReceived []byte

	busReceive.Subscribe(pubsub.SubjFileSystemChange, func(data []byte) {
		dataReceived = data
	})

	require.Eventually(t, func() bool {
		busPublish.Publish(pubsub.SubjFileSystemChange, dataSend)
		return string(dataReceived) == string(dataSend)
	}, 3*time.Second, 100*time.Millisecond, "data not received")
}
