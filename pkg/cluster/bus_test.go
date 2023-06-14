package cluster

import (
	"fmt"
	"testing"
	"time"

	"github.com/nsqio/go-nsq"
	. "github.com/stretchr/testify/assert"
)

func TestBusConfig(t *testing.T) {

	config := DefaultConfig()

	// setting data dir with temp folder
	b, err := newBus(config)
	NoError(t, err)
	NotEmpty(t, b.dataDir)
	defer b.Stop()

	config.DataDir = "/tmp/jens"

	b2, err := newBus(config)
	defer b2.Stop()
	Nil(t, err)
	NotNil(t, b)

	Equal(t, "/tmp/jens", b2.dataDir)

}

func TestBusCluster(t *testing.T) {

	config := DefaultConfig()

	// setting data dir with temp folder
	b, err := newBus(config)
	Nil(t, err)
	defer b.Stop()
	go b.Start()
	b.WaitTillConnected()

	config.NSQDPort = 4250
	config.NSQDListenHTTPPort = 4251
	config.NSQLookupPort = 4252
	config.NSQLookupListenHTTPPort = 4253

	b2, err := newBus(config)
	Nil(t, err)
	defer b2.Stop()
	go b2.Start()
	b.WaitTillConnected()

	err = b.UpdateBusNodes([]string{
		"127.0.0.1:4151",
		"127.0.0.1:4252",
	})

	err = b2.UpdateBusNodes([]string{
		"127.0.0.1:4151",
		"127.0.0.1:4252",
	})

	// both instances should have 2 nodes
	Eventually(t, func() bool {
		newNodes, err := b.Nodes()
		if err != nil {
			return false
		}
		newNodes2, err := b2.Nodes()
		if err != nil {
			return false
		}
		if len(newNodes.Producers) == 2 &&
			len(newNodes2.Producers) == 2 {
			return true
		}
		return false
	}, 10*time.Second, time.Second)

	// add topcs to both busses
	addTopics := func(bin *bus) {
		bin.ModifyTopic("topic1", true)
		bin.CreateChannel("topic1", "ch1")
		bin.CreateChannel("topic1", "ch2")
		bin.ModifyTopic("topic2", true)
		bin.CreateChannel("topic2", "ch3")
	}
	addTopics(b)
	addTopics(b2)

	time.Sleep(2 * time.Second)
	// t.FailNow()

	clientConfig := nsq.NewConfig()

	createConsumer := func(topic, channel, connect string, mh *messageHandler) {
		consumer, _ := nsq.NewConsumer(topic, channel, clientConfig)
		consumer.AddHandler(mh)
		consumer.ConnectToNSQLookupd(connect)
		// time.Sleep(5 * time.Second)
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

	// consumer, err := nsq.NewConsumer("topic1", "channel1", clientConfig)
	// Nil(t, err)
	// consumer.AddHandler(&myMessageHandler{
	// 	bus: "bus1",
	// })
	// err = consumer.ConnectToNSQLookupd("localhost:4153")
	// Nil(t, err)

	// consumer2, err := nsq.NewConsumer("topic1", "channel2", clientConfig)
	// Nil(t, err)
	// consumer2.AddHandler(&myMessageHandler{
	// 	bus: "bus2",
	// })
	// err = consumer2.ConnectToNSQLookupd("localhost:4253")
	// Nil(t, err)

	// consumer3, err := nsq.NewConsumer("topic2", "channel3", clientConfig)
	// Nil(t, err)
	// consumer3.AddHandler(&myMessageHandler{
	// 	bus: "bus3",
	// })
	// err = consumer3.ConnectToNSQLookupd("localhost:4153")
	// Nil(t, err)

	producer, err := nsq.NewProducer("127.0.0.1:4150", clientConfig)
	Nil(t, err)
	err = producer.Publish("topic1", []byte("JENS1"))
	err = producer.Publish("topic1", []byte("JENS2"))
	err = producer.Publish("topic1", []byte("JENS3"))
	err = producer.Publish("topic1", []byte("JENS4"))

	fmt.Printf("?????????????????????????????????????????????ERR %v\n", err)

	// producer2, err := nsq.NewProducer("127.0.0.1:4250", clientConfig)
	// Nil(t, err)
	err = producer.Publish("topic2", []byte("JENS1"))

	fmt.Printf("?????????????????????????????????????????????ERR %v\n", err)

	// producer.Stop()
	// producer2.Stop()
	time.Sleep(10 * time.Second)

	fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!! %d %d %d\n", mh1.counter, mh2.counter, mh3.counter)
	// pl, err := b.Nodes()
	// NoError(t, err)

	// for i := range pl.Producers {
	// 	p := pl.Producers[i]
	// 	t.Logf("topics %v: %v", p.TCPPort, p.Topics)
	// }

	// connect to first one and subscribe to second

}

type messageHandler struct {
	bus     string
	counter int
}

func (h *messageHandler) HandleMessage(m *nsq.Message) error {
	h.counter += 1
	fmt.Printf("%v <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<,\n", h.bus)
	return nil
}
