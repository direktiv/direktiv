package cluster

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBusConfig(t *testing.T) {
	config := DefaultConfig()

	// setting data dir with temp folder
	logger := zap.NewNop().Sugar()
	b, err := newBus(config, logger)
	require.NoError(t, err)
	require.NotEmpty(t, b.dataDir)
	defer b.stop()

	config.NSQDataDir = "/tmp/test"

	ports := getPorts(t)

	config.NSQDPort = ports[0].port
	config.NSQDListenHTTPPort = ports[1].port
	config.NSQLookupPort = ports[2].port
	config.NSQLookupListenHTTPPort = ports[3].port

	closePorts(ports)

	b2, err := newBus(config, zap.NewNop().Sugar())
	require.NoError(t, err)
	defer b2.stop()

	require.NotNil(t, b2)
	assert.Equal(t, "/tmp/test", b2.dataDir)
}

func TestBusFunctions(t *testing.T) {
	config := DefaultConfig()

	// setting data dir with temp folder
	logger := zap.NewNop().Sugar()
	b, err := newBus(config, logger)
	require.NoError(t, err)
	defer b.stop()

	go b.start()
	err = b.waitTillConnected(100)
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

	err = b.updateBusNodes(context.TODO(), []string{"server1:5555"})
	assert.NoError(t, err)
}

type randomPort struct {
	l    net.Listener
	port int
}

func closePorts(rp []randomPort) {
	for i := range rp {
		r := rp[i]
		r.l.Close()
	}
}

func getPorts(t *testing.T) []randomPort {
	ports := make([]randomPort, 5)

	for i := 0; i < 5; i++ {
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			t.FailNow()
		}
		ports[i] = randomPort{
			l:    l,
			port: l.Addr().(*net.TCPAddr).Port,
		}
	}

	return ports
}

func TestBusCluster(t *testing.T) {
	config := DefaultConfig()

	ports1 := getPorts(t)

	config.NSQDPort = ports1[0].port
	config.NSQDListenHTTPPort = ports1[1].port
	config.NSQLookupPort = ports1[2].port
	config.NSQLookupListenHTTPPort = ports1[3].port
	closePorts(ports1)

	// setting data dir with temp folder
	logger := zap.NewNop().Sugar()
	b, err := newBus(config, logger)
	require.NoError(t, err)
	defer b.stop()
	go b.start()
	b.waitTillConnected(100)

	ports2 := getPorts(t)
	closePorts(ports2)

	config.NSQDPort = ports2[0].port
	config.NSQDListenHTTPPort = ports2[1].port
	config.NSQLookupPort = ports2[2].port
	config.NSQLookupListenHTTPPort = ports2[3].port

	b2, err := newBus(config, zap.NewNop().Sugar())
	require.NoError(t, err)
	defer b2.stop()
	go b2.start()
	b.waitTillConnected(100)

	// update cluster
	err = b.updateBusNodes(
		context.TODO(), []string{
			fmt.Sprintf("127.0.0.1:%d", ports1[2].port),
			fmt.Sprintf("127.0.0.1:%d", ports2[2].port),
		})
	require.NoError(t, err)
	err = b2.updateBusNodes(context.TODO(), []string{
		fmt.Sprintf("127.0.0.1:%d", ports1[2].port),
		fmt.Sprintf("127.0.0.1:%d", ports2[2].port),
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
	}, 10*time.Second, 100*time.Millisecond, "all nodes test failed")

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
	// clientConfig.LookupdPollInterval = time.Millisecond * 100

	createConsumer := func(topic, channel, connect string, mh *messageHandler) {
		consumer, _ := nsq.NewConsumer(topic, channel, clientConfig)
		consumer.AddHandler(mh)
		consumer.ConnectToNSQLookupd(connect)
	}

	mh1 := &messageHandler{
		bus: "mh1",
	}
	createConsumer("topic1", "ch1", fmt.Sprintf("localhost:%d", ports1[3].port), mh1)
	mh2 := &messageHandler{
		bus: "mh2",
	}
	createConsumer("topic1", "ch2", fmt.Sprintf("localhost:%d", ports2[3].port), mh2)
	mh3 := &messageHandler{
		bus: "mh3",
	}
	createConsumer("topic2", "ch3", fmt.Sprintf("localhost:%d", ports2[3].port), mh3)
	mh4 := &messageHandler{
		bus: "mh4",
	}
	createConsumer("topic2", "ch3", fmt.Sprintf("localhost:%d", ports2[3].port), mh4)

	// send messages
	producer, err := nsq.NewProducer(fmt.Sprintf("127.0.0.1:%d", ports1[0].port), clientConfig)
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
	}, 60*time.Second, 100*time.Millisecond, "one channel does not work")
}

type messageHandler struct {
	bus     string
	counter int
}

func (h *messageHandler) HandleMessage(m *nsq.Message) error {
	h.counter += 1

	return nil
}
