package nats_test

import (
	"context"
	"testing"
	"time"

	"github.com/direktiv/direktiv/internal/cluster/pubsub"
	natspubsub "github.com/direktiv/direktiv/internal/cluster/pubsub/nats"
	"github.com/direktiv/direktiv/internal/core"
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
	nc, err := nats.Connect(cs)
	require.NoError(t, err)
	busPublish := natspubsub.New(nc)

	circuit := core.NewCircuit(context.Background())
	go busPublish.Loop(circuit)

	nc2, err := nats.Connect(cs)
	require.NoError(t, err)
	busReceive := natspubsub.New(nc2)

	dataSend := []byte("test data")
	var dataReceived []byte

	busReceive.Subscribe(pubsub.FileSystemChangeEvent, func(data []byte) {
		dataReceived = data
	})

	require.Eventually(t, func() bool {
		busPublish.Publish(pubsub.FileSystemChangeEvent, dataSend)
		return string(dataReceived) == string(dataSend)
	}, 3*time.Second, 100*time.Millisecond, "data not received")
}
