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
		if nodes[i].serfServer.NumNodes() != len(nodes) {
			return false
		}

		nn, _ := nodes[i].bus.nodes()

		if len(nn.Producers) != len(nodes) {
			return false
		}

	}

	return true
}

func createConfig(topics []string, i int, change bool) Config {

	config := DefaultConfig()
	config.SerfReapTimeout = 3 * time.Second
	config.SerfReapInterval = 1 * time.Second
	config.SerfTombstoneTimeout = 5 * time.Second

	if change {
		config.NSQInactiveTimeout = 10 * time.Second
		config.NSQTombstoneTimeout = 5 * time.Second
	}

	config.NSQTopics = topics

	config.SerfPort = 5223 + (100 * i)
	config.NSQDPort = 4250 + (100 * i)
	config.NSQDListenHTTPPort = 4251 + (100 * i)
	config.NSQLookupPort = 4252 + (100 * i)
	config.NSQLookupListenHTTPPort = 4253 + (100 * i)

	return config

}

func createCluster(count int, topics []string, change bool) ([]*Node, error) {

	nfNodes := make([]string, 0)
	finalNodes := make([]*Node, 0)

	hn, _ := os.Hostname()

	for i := 0; i < count; i++ {
		nfNodes = append(nfNodes, fmt.Sprintf("%s:%d", hn, 5223+(100*i)))
	}

	nf := NewNodefinderStatic(nfNodes)

	for i := 0; i < count; i++ {
		config := createConfig(topics, i, change)
		config.Nodefinder = nf

		node, err := NewNode(config)
		if err != nil {
			return nil, err
		}
		finalNodes = append(finalNodes, node)
	}

	return finalNodes, nil
}

func TestCluster(t *testing.T) {

	count := 3
	nodes, err := createCluster(count, []string{"topic1"}, true)
	require.NoError(t, err)

	for i := 0; i < count; i++ {
		defer nodes[i].Stop()
	}

	// check three node cluster
	require.Eventually(t, func() bool {

		return rightNumber(nodes)

	}, 10*time.Second, time.Second)

	// // stop one node
	err = nodes[count-1].Stop()
	assert.NoError(t, err)

	nodes = append(nodes[:count-1], nodes[count-1+1:]...)

	// // there should be only two nodes now and two bus
	require.Eventually(t, func() bool {
		return rightNumber(nodes)
	}, 60*time.Second, time.Second)

	config := DefaultConfig()
	config.SerfReapTimeout = 3 * time.Second
	config.SerfReapInterval = 1 * time.Second
	config.SerfTombstoneTimeout = 5 * time.Second
	config.NSQInactiveTimeout = 60 * time.Second
	config.NSQTombstoneTimeout = 10 * time.Second

	config.NSQTopics = []string{"topic1"}

	config.SerfPort = 7777
	config.NSQDPort = 7800
	config.NSQDListenHTTPPort = 7801
	config.NSQLookupPort = 7802
	config.NSQLookupListenHTTPPort = 7803

	nfNodes := []string{
		"127.0.0.1:5223",
		"127.0.0.1:5323",
		"127.0.0.1:7777",
	}

	// add a node again
	nf := NewNodefinderStatic(nfNodes)
	// config := createConfig([]string{"topic1"}, 2)
	config.Nodefinder = nf

	newNode, err := NewNode(config)
	require.NoError(t, err)
	nodes = append(nodes, newNode)

	require.Eventually(t, func() bool {
		return rightNumber(nodes)
	}, 60*time.Second, time.Second)

}

func TestClusterSubscribe(t *testing.T) {

	count := 3
	nodes, err := createCluster(count, []string{"topic1", "topic2"}, false)
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
	}, 30*time.Second, time.Second)

}

type counterHandler struct {
	cc int
}

var j int

func (ch *counterHandler) counter(msg []byte) error {
	j += 1
	// fmt.Printf("COUNT %v\n", j)
	ch.cc += 1
	return nil
}
