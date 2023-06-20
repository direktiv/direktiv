package cluster

import (
	"fmt"
	"testing"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBusConfig(t *testing.T) {

	config := DefaultConfig()

	// setting data dir with temp folder
	b, err := newBus(config)
	require.NoError(t, err)
	require.NotEmpty(t, b.dataDir)
	defer b.stop()

	config.NSQDataDir = "/tmp/test"
	config.NSQDPort = 4250
	config.NSQDListenHTTPPort = 4251
	config.NSQLookupPort = 4252
	config.NSQLookupListenHTTPPort = 4253

	b2, err := newBus(config)
	require.NoError(t, err)
	defer b2.stop()

	require.NotNil(t, b2)
	assert.Equal(t, "/tmp/test", b2.dataDir)

}

func TestBusFunctions(t *testing.T) {

	config := DefaultConfig()

	// setting data dir with temp folder
	b, err := newBus(config)
	require.NoError(t, err)
	defer b.stop()

	go b.start()
	err = b.waitTillConnected()
	require.NoError(t, err)

	err = b.createTopic("topic1")
	assert.NoError(t, err)

	err = b.createTopic("^%&^%&!")
	assert.Error(t, err)

	err = b.createDeleteChannel("topic1", "channel1", true)
	assert.NoError(t, err)

	err = b.createDeleteChannel("topic1", "&^&*^%&^%&^%", true)
	assert.Error(t, err)

	err = b.createDeleteChannel("unknown", "channel1", true)
	assert.Error(t, err)

	err = b.updateBusNodes([]string{"server1:5555"})
	assert.NoError(t, err)

}

func TestBusCluster(t *testing.T) {

	config := DefaultConfig()

	// setting data dir with temp folder
	b, err := newBus(config)
	require.NoError(t, err)
	defer b.stop()
	go b.start()
	b.waitTillConnected()

	config.NSQDPort = 4250
	config.NSQDListenHTTPPort = 4251
	config.NSQLookupPort = 4252
	config.NSQLookupListenHTTPPort = 4253

	b2, err := newBus(config)
	require.NoError(t, err)
	defer b2.stop()
	go b2.start()
	b.waitTillConnected()

	// update cluster
	err = b.updateBusNodes([]string{
		"127.0.0.1:4151",
		"127.0.0.1:4252",
	})
	require.NoError(t, err)
	err = b2.updateBusNodes([]string{
		"127.0.0.1:4151",
		"127.0.0.1:4252",
	})
	require.NoError(t, err)

	// both instances should have 2 nodes
	require.Eventually(t, func() bool {
		newNodes, err := b.nodes()
		if err != nil {
			return false
		}
		newNodes2, err := b2.nodes()
		if err != nil {
			return false
		}
		if len(newNodes.Producers) == 2 &&
			len(newNodes2.Producers) == 2 {
			return true
		}
		return false
	}, 60*time.Second, time.Second, "all nodes test failed")

	// add topcs to both busses
	addTopics := func(bin *bus) {
		bin.createTopic("topic1")
		bin.createTopic("topic2")
		bin.createDeleteChannel("topic1", "ch1", true)
		bin.createDeleteChannel("topic1", "ch2", true)
		bin.createDeleteChannel("topic2", "ch3", true)
	}
	addTopics(b)
	addTopics(b2)

	clientConfig := nsq.NewConfig()

	createConsumer := func(topic, channel, connect string, mh *messageHandler) {
		consumer, _ := nsq.NewConsumer(topic, channel, clientConfig)
		consumer.AddHandler(mh)
		consumer.ConnectToNSQLookupd(connect)
	}

	mh1 := &messageHandler{
		bus: "mh1",
	}
	createConsumer("topic1", "ch1", "localhost:4153", mh1)
	mh2 := &messageHandler{
		bus: "mh2",
	}
	createConsumer("topic1", "ch2", "localhost:4253", mh2)
	mh3 := &messageHandler{
		bus: "mh3",
	}
	createConsumer("topic2", "ch3", "localhost:4253", mh3)
	mh4 := &messageHandler{
		bus: "mh4",
	}
	createConsumer("topic2", "ch3", "localhost:4253", mh4)

	// send messages
	producer, err := nsq.NewProducer("127.0.0.1:4150", clientConfig)
	require.NoError(t, err)
	defer producer.Stop()

	err = producer.Publish("topic1", []byte("msg1"))
	assert.NoError(t, err)
	err = producer.Publish("topic1", []byte("msg2"))
	assert.NoError(t, err)
	err = producer.Publish("topic2", []byte("msg3"))
	assert.NoError(t, err)

	require.Eventually(t, func() bool {
		status := true

		// both message handler should have gotten the two messages
		if mh1.counter != 2 || mh2.counter != 2 {
			status = false
		}

		// on the same channel. only one should get it
		if (mh3.counter == 1 && mh4.counter == 1) ||
			(mh3.counter == 0 && mh4.counter == 0) {
			status = false
		}

		return status
	}, 60*time.Second, time.Second, "one chanel does not work")

}

type messageHandler struct {
	bus     string
	counter int
}

func (h *messageHandler) HandleMessage(m *nsq.Message) error {
	h.counter += 1
	fmt.Printf("%s: %d\n", h.bus, h.counter)
	return nil
}
