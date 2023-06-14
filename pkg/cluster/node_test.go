package cluster

import (
	"testing"
	"time"

	. "github.com/stretchr/testify/assert"
)

func TestNodeConfig(t *testing.T) {

	newPort := 3333
	newTimeout := time.Minute

	config := DefaultConfig()

	Equal(t, config.SerfPort, defaultSerfPort)
	Equal(t, config.SerfReapTimeout, defaultReapTimeout)

	Equal(t, config.NSQDPort, defaultNSQDPort)
	Equal(t, config.NSQLookupPort, defaultNSQLookupPort)
	Equal(t, config.NSQDListenHTTPPort, defaultNSQDListenHTTPPort)
	Equal(t, config.NSQLookupListenHTTPPort, defaultNSQLookupListenHTTPPort)

	config.SerfPort = newPort
	config.SerfReapTimeout = time.Minute

	Equal(t, config.SerfPort, newPort)
	Equal(t, config.SerfReapTimeout, newTimeout)

	// no nodefinder should fail
	config.Nodefinder = nil
	_, err := NewNode(config)
	NotNil(t, err)

}

func TestNewNode(t *testing.T) {

	config := DefaultConfig()

	node, err := NewNode(config)
	NotNil(t, err)

	nodes := node.serfServer.NumNodes()
	Equal(t, nodes, 1)

	node.Stop()

}

func TestCluster(t *testing.T) {

	nf := NewNodefinderStatic([]string{
		"127.0.0.1:5222",
		"127.0.0.1:5223",
		"127.0.0.1:5224",
	})

	config := DefaultConfig()
	config.Nodefinder = nf
	config.SerfReapTimeout = 3 * time.Second
	config.SerfReapInterval = 1 * time.Second
	config.SerfTombstoneTimeout = 5 * time.Second

	node1, err := NewNode(config)
	NotNil(t, err)

	config.SerfPort = 5223
	node2, err := NewNode(config)
	NotNil(t, err)

	config.SerfPort = 5224
	node3, err := NewNode(config)
	NotNil(t, err)

	// start three node cluster
	Eventually(t, func() bool {
		return node1.serfServer.NumNodes() == 3 && node2.serfServer.NumNodes() == 3 && node3.serfServer.NumNodes() == 3
	}, 10*time.Second, time.Second)

	// stop one node
	err = node3.Stop()
	NotNil(t, err)

	// there should be only two nodes now
	Eventually(t, func() bool {
		return node1.serfServer.NumNodes() == 2 && node2.serfServer.NumNodes() == 2
	}, 60*time.Second, time.Second)

	// let a new node start
	config.SerfPort = 5225
	node3, err = NewNode(config)
	NotNil(t, err)

	Eventually(t, func() bool {
		return node1.serfServer.NumNodes() == 3 && node2.serfServer.NumNodes() == 3 && node3.serfServer.NumNodes() == 3
	}, 10*time.Second, time.Second)

}
