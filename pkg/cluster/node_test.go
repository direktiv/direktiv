package cluster

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeConfig(t *testing.T) {
	newPort := 3333
	newTimeout := time.Minute

	config := DefaultConfig()

	assert.Equal(t, config.SerfPort, defaultSerfPort)
	assert.Equal(t, config.SerfReapTimeout, defaultReapTimeout)

	assert.Equal(t, config.NSQDPort, defaultNSQDPort)
	assert.Equal(t, config.NSQLookupPort, defaultNSQLookupPort)
	assert.Equal(t, config.NSQDListenHTTPPort, defaultNSQDListenHTTPPort)
	assert.Equal(t, config.NSQLookupListenHTTPPort, defaultNSQLookupListenHTTPPort)

	config.SerfPort = newPort
	config.SerfReapTimeout = time.Minute

	assert.Equal(t, config.SerfPort, newPort)
	assert.Equal(t, config.SerfReapTimeout, newTimeout)

	// no nodefinder should fail
	config.Nodefinder = nil
	_, err := NewNode(config)
	require.Error(t, err)
}

func TestNewNode(t *testing.T) {
	config := DefaultConfig()

	node, err := NewNode(config)
	require.NoError(t, err)
	defer node.Stop()

	nodes := node.serfServer.NumNodes()
	assert.Equal(t, nodes, 1)
}

func rightNumber(nodes []*Node) bool {
	for i := 0; i < len(nodes); i++ {

		fmt.Printf("compare nodes: %d - %d\n", len(nodes), nodes[i].serfServer.NumNodes())
		if nodes[i].serfServer.NumNodes() != len(nodes) {
			return false
		}

		nn, _ := nodes[i].bus.nodes()

		fmt.Printf("compare producers: %d - %d\n", len(nn.Producers), len(nodes))
		if len(nn.Producers) != len(nodes) {
			return false
		}

	}

	return true
}

func createConfig(t *testing.T, topics []string, change bool) (Config, []randomPort) {
	config := DefaultConfig()

	if change {
		config.NSQInactiveTimeout = 10 * time.Second
		config.NSQTombstoneTimeout = 5 * time.Second

		config.SerfReapTimeout = 3 * time.Second
		config.SerfReapInterval = 1 * time.Second
		config.SerfTombstoneTimeout = 5 * time.Second
	}

	config.NSQTopics = topics

	ports := getPorts(t)
	config.SerfPort = ports[0].port
	config.NSQDPort = ports[1].port
	config.NSQDListenHTTPPort = ports[2].port
	config.NSQLookupPort = ports[3].port
	config.NSQLookupListenHTTPPort = ports[4].port

	// config.SerfPort = port11
	// config.NSQDPort = port12
	// config.NSQDListenHTTPPort = port13
	// config.NSQLookupPort = port14
	// config.NSQLookupListenHTTPPort = portSerf

	return config, ports
}

func createCluster(t *testing.T, count int, topics []string, change bool) ([]*Node, error) {
	nfNodes := make([]string, 0)
	finalNodes := make([]*Node, 0)

	configs := make([]Config, 0)
	ports := make([][]randomPort, count)

	hn, _ := os.Hostname()

	for i := 0; i < count; i++ {
		config, ports1 := createConfig(t, topics, change)
		nfNodes = append(nfNodes, fmt.Sprintf("%s:%d", hn, config.SerfPort))
		configs = append(configs, config)
		ports[i] = ports1
		t.Logf("serf port: %+v\n", config.SerfPort)
		t.Logf("nsq port: %+v\n", config.NSQDPort)
		t.Logf("nsq http port: %+v\n", config.NSQDListenHTTPPort)
		t.Logf("nsq lookup port: %+v\n", config.NSQLookupPort)
		t.Logf("nsq lookup http port: %+v\n", config.NSQLookupListenHTTPPort)
		t.Logf("----------------------")
	}

	for i := 0; i < count; i++ {
		closePorts(ports[i])
	}

	nf := NewNodefinderStatic(nfNodes)

	for i := 0; i < count; i++ {
		c := configs[i]
		c.Nodefinder = nf
		node, err := NewNode(c)
		if err != nil {
			return nil, err
		}
		finalNodes = append(finalNodes, node)
	}

	return finalNodes, nil
}

func TestCluster(t *testing.T) {
	count := 3
	nodes, err := createCluster(t, count, []string{"topic1"}, true)
	require.NoError(t, err)

	for i := 0; i < count; i++ {
		defer nodes[i].Stop()
	}

	// check three node cluster
	require.Eventually(t, func() bool {
		return rightNumber(nodes)
	}, 10*time.Second, time.Second, "node count failed")

	// // stop one node
	err = nodes[count-1].Stop()
	assert.NoError(t, err)

	nodes = append(nodes[:count-1], nodes[count-1+1:]...)

	// // there should be only two nodes now and two bus
	require.Eventually(t, func() bool {
		return rightNumber(nodes)
	}, 60*time.Second, time.Second)

	conf, ports1 := createConfig(t, []string{"topic1"}, false)
	nfNodes := []string{
		fmt.Sprintf("127.0.0.1:%d", conf.SerfPort),
	}

	for i := range nodes {
		nfNodes = append(nfNodes, fmt.Sprintf("127.0.0.1:%d",
			nodes[i].serfServer.LocalMember().Port))
	}
	// add a node again
	nf := NewNodefinderStatic(nfNodes)
	conf.Nodefinder = nf

	closePorts(ports1)

	newNode, err := NewNode(conf)
	require.NoError(t, err)
	defer newNode.Stop()
	nodes = append(nodes, newNode)

	require.Eventually(t, func() bool {
		return rightNumber(nodes)
	}, 60*time.Second, time.Second)
}

func TestClusterSubscribe(t *testing.T) {
	count := 3
	nodes, err := createCluster(t, count, []string{"topic1", "topic2"}, false)
	require.NoError(t, err)

	for i := 0; i < count; i++ {
		defer nodes[i].Stop()
	}

	// check three node cluster
	require.Eventually(t, func() bool {
		return rightNumber(nodes)
	}, 10*time.Second, time.Second)

	// they should all get the message
	counter1 := &counterHandler{}
	mc1, err := nodes[0].Subscribe("topic1", counter1.counter)
	require.NoError(t, err)
	defer nodes[0].Unsubscribe(mc1)

	counter2 := &counterHandler{}
	mc2, err := nodes[1].Subscribe("topic1", counter2.counter)
	require.NoError(t, err)
	defer nodes[1].Unsubscribe(mc2)

	counter3 := &counterHandler{}
	mc3, err := nodes[2].Subscribe("topic1", counter3.counter)
	require.NoError(t, err)
	defer nodes[2].Unsubscribe(mc3)

	err = nodes[0].Publish("topic1", []byte("msg"))
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		t.Logf("received events on nodes: %d %d %d", counter1.cc, counter2.cc, counter3.cc)
		return counter1.cc == 1 && counter2.cc == 1 && counter3.cc == 1
	}, 30*time.Second, time.Second)

	add := 10
	for i := 0; i < add; i++ {
		err = nodes[0].Publish("topic1", []byte("msg1"))
		require.NoError(t, err)
	}

	require.Eventually(t, func() bool {
		t.Logf("received events on nodes: %d %d %d", counter1.cc, counter2.cc, counter3.cc)
		return counter1.cc == add+1 && counter2.cc == add+1 && counter3.cc == add+1
	}, 30*time.Second, time.Second)

	// test single subscriber
	counter1 = &counterHandler{}
	mc1, err = nodes[0].SubscribeOnce("topic2", counter1.counter)
	require.NoError(t, err)
	defer nodes[0].Unsubscribe(mc1)

	counter2 = &counterHandler{}
	mc2, err = nodes[1].SubscribeOnce("topic2", counter2.counter)
	require.NoError(t, err)
	defer nodes[1].Unsubscribe(mc2)

	counter3 = &counterHandler{}
	mc3, err = nodes[2].SubscribeOnce("topic2", counter3.counter)
	require.NoError(t, err)
	defer nodes[2].Unsubscribe(mc3)

	t.Logf("received events on nodes2: %d %d %d", counter1.cc, counter2.cc, counter3.cc)

	for i := 0; i < 10; i++ {
		err = nodes[0].Publish("topic2", []byte("msg1"))
		require.NoError(t, err)
	}

	require.Eventually(t, func() bool {
		t.Logf("received events on nodes2: %d %d %d", counter1.cc, counter2.cc, counter3.cc)
		return counter1.cc+counter2.cc+counter3.cc == 10
	}, 30*time.Second, time.Second, "did not get recieved events")
}

type counterHandler struct {
	cc int
}

var j int

func (ch *counterHandler) counter(msg []byte) error {
	j += 1
	ch.cc += 1
	return nil
}
