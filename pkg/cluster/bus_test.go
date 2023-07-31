package cluster

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
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
	os.Mkdir("/tmp/test", 0o700)
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
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   time.Minute,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ResponseHeaderTimeout: time.Minute,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}
	// setting data dir with temp folder
	logger := zap.NewNop().Sugar()
	b, err := newBus(config, logger)
	require.NoError(t, err)
	defer b.stop()

	go b.start(context.Background(), 10*time.Second)
	err = b.waitTillConnected(context.Background(), client, 10*time.Millisecond, 10*time.Millisecond)
	require.NoError(t, err)

	err = b.createTopic(context.Background(), "topic1", client)
	assert.NoError(t, err)

	err = b.createTopic(context.Background(), "^%&^%&!", client)
	assert.Error(t, err)

	err = b.createDeleteChannel(context.Background(), client, "topic1", "channel1", true)
	assert.NoError(t, err)

	err = b.createDeleteChannel(context.Background(), client, "topic1", "&^&*^%&^%&^%", true)
	assert.Error(t, err)

	err = b.createDeleteChannel(context.Background(), client, "unknown", "channel1", true)
	assert.Error(t, err)

	err = b.updateBusNodes(context.TODO(), []string{"server1:5555"}, client)
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
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   time.Minute,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ResponseHeaderTimeout: time.Minute,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}
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
	go b.start(context.Background(), 10*time.Second)
	b.waitTillConnected(context.Background(), client, 100, 100)

	ports2 := getPorts(t)
	closePorts(ports2)

	config.NSQDPort = ports2[0].port
	config.NSQDListenHTTPPort = ports2[1].port
	config.NSQLookupPort = ports2[2].port
	config.NSQLookupListenHTTPPort = ports2[3].port

	b2, err := newBus(config, zap.NewNop().Sugar())
	require.NoError(t, err)
	defer b2.stop()
	go b2.start(context.Background(), 10*time.Second)
	b.waitTillConnected(context.Background(), client, 100, 100)

	// update cluster
	err = b.updateBusNodes(
		context.TODO(), []string{
			fmt.Sprintf("127.0.0.1:%d", ports1[2].port),
			fmt.Sprintf("127.0.0.1:%d", ports2[2].port),
		}, client)
	require.NoError(t, err)
	err = b2.updateBusNodes(context.TODO(), []string{
		fmt.Sprintf("127.0.0.1:%d", ports1[2].port),
		fmt.Sprintf("127.0.0.1:%d", ports2[2].port),
	}, client)
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
		bin.createTopic(context.Background(), "topic1", client)
		bin.createTopic(context.Background(), "topic2", client)
		bin.createDeleteChannel(context.Background(), client, "topic1", "ch1", true)
		bin.createDeleteChannel(context.Background(), client, "topic1", "ch2", true)
		bin.createDeleteChannel(context.Background(), client, "topic2", "ch3", true)
	}
	addTopics(b)
	addTopics(b2)

	clientConfig := nsq.NewConfig()
	// clientConfig.LookupdPollInterval = time.Millisecond * 100

	createConsumer := func(topic, channel, connect string, mh *messageHandler) {
		consumer, err := nsq.NewConsumer(topic, channel, clientConfig)
		if err != nil {
			t.Errorf("creating client failed %s", err)
		}
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

	status := make(chan bool)

	go func() {
		// Both message handlers should have gotten the two messages
		timeout := time.After(3 * time.Minute)

		for {
			if !(mh1.counter != 2 || mh2.counter != 2) {
				// On the same channel. Only one should get it
				if !((mh3.counter == 1 && mh4.counter == 1) ||
					(mh3.counter == 0 && mh4.counter == 0)) {
					status <- true
				}
			}

			select {
			case <-timeout:
				// Handle the timeout case
				// Add any necessary code here to handle the timeout condition
				status <- false
				return // Return from the goroutine to stop it
			case <-time.After(time.Second):
				// Wait for 1 second before checking again
			}
		}
	}()

	// Read the value from the status channel
	if <-status == false {
		t.Errorf("time out waiting for messages from bus")
	}
}

func waitForMessageHandlers(t *testing.T, mh1, mh2, mh3, mh4 *messageHandler, expectedMsgCount int) {
	// Create a channel to signal when the conditions are met or timeout occurs
	status := make(chan bool)

	// Start a goroutine to check the conditions
	go func() {
		timeout := 60 * time.Second
		interval := 100 * time.Millisecond
		maxAttempts := int(timeout / interval)

		for attempt := 0; attempt < maxAttempts; attempt++ {
			// Check the conditions
			if mh1.counter == expectedMsgCount &&
				mh2.counter == expectedMsgCount &&
				((mh3.counter == 1 && mh4.counter == 1) || (mh3.counter == 0 && mh4.counter == 0)) {
				status <- true // Signal that conditions are met
				return
			}

			// Sleep for the specified interval before the next attempt
			time.Sleep(interval)
		}

		// If the loop completes without meeting the expected conditions, signal failure
		status <- false
	}()

	// Wait for the goroutine to signal completion or timeout
	select {
	case success := <-status:
		if success {
			// Test successful, conditions are met
			return
		}
		// If status is false, it means the test failed due to the timeout
		t.Fatalf("One channel does not work: mh1.counter=%d, mh2.counter=%d, mh3.counter=%d, mh4.counter=%d",
			mh1.counter, mh2.counter, mh3.counter, mh4.counter)

	case <-time.After(time.Minute * 2):
		// Timeout occurred, fail the test
		t.Fatalf("Timeout: One channel does not work: mh1.counter=%d, mh2.counter=%d, mh3.counter=%d, mh4.counter=%d",
			mh1.counter, mh2.counter, mh3.counter, mh4.counter)
	}
}

type messageHandler struct {
	bus     string
	counter int
}

func (h *messageHandler) HandleMessage(m *nsq.Message) error {
	h.counter += 1

	return nil
}
