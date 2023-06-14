// Package cluster manages serf nodes as a cluster in Direktiv.
// In particular it manages the underlying nsq cluster and is used
// to add and remove nodes dynamically.
package cluster

import (
	"fmt"
	"net"
	"time"

	"github.com/direktiv/direktiv/pkg/dlog"
	"github.com/hashicorp/serf/serf"
	"go.uber.org/zap"
)

const (
	defaultSerfPort         = 5222
	defaultReapTimeout      = 5 * time.Minute
	defaultReapInterval     = 10 * time.Second
	defaultTombstoneTimeout = 6 * time.Hour

	defaultNSQDPort                = 4150
	defaultNSQLookupPort           = 4151
	defaultNSQDListenHTTPPort      = 4152
	defaultNSQLookupListenHTTPPort = 4153

	serfAddress = "SERF_ADDRESS"
)

// Config is the configuration structure for one node
type Config struct {
	// Port for the serf server.
	SerfPort int

	// SerfReapTimeout: Time after other nodes are getting reaped
	// if they have been marked as failed.
	// SerfReapInterval: Interval how often the reaper runs.
	// SerfTombstoneTimeout: Time after a node is getting
	// removed after it has left the cluster.
	SerfReapTimeout, SerfReapInterval,
	SerfTombstoneTimeout time.Duration

	Nodefinder Nodefinder

	// nsq settings
	NSQDPort, NSQLookupPort,
	NSQLookupListenHTTPPort, NSQDListenHTTPPort int

	DataDir string
}

type Node struct {
	logger *zap.SugaredLogger

	// serf settings
	serfServer *serf.Serf
	events     chan serf.Event

	nodefinder Nodefinder

	upCh chan bool

	Bus *bus
}

// nodefinders have to return all ips/addr of the serf nodes available on startup
// the minimum is to return one ip/addr for serf to form a cluster
type Nodefinder interface {
	GetNodes() ([]string, error)
	GetAddr() (string, error)
}

func NewNode(config Config) (*Node, error) {

	logger, err := dlog.ApplicationLogger("node")
	if err != nil {
		return nil, err
	}

	if config.Nodefinder == nil {
		return nil, fmt.Errorf("nodefinder not set")
	}

	node := &Node{
		logger:     logger,
		nodefinder: config.Nodefinder,
		upCh:       make(chan bool),
	}

	node.Bus, err = newBus(config)
	if err != nil {
		return nil, err
	}
	go node.Bus.Start()
	node.Bus.WaitTillConnected()

	serfConfig := serf.DefaultConfig()
	serfConfig.Init()

	serfConfig.Tags = make(map[string]string)

	addr, err := config.Nodefinder.GetAddr()
	if err != nil {
		return nil, err
	}
	serfConfig.NodeName = net.JoinHostPort(addr, fmt.Sprintf("%d", config.SerfPort))

	serfConfig.MemberlistConfig.BindAddr = net.IPv4zero.String()
	serfConfig.MemberlistConfig.BindPort = config.SerfPort

	serfConfig.ReapInterval = config.SerfReapInterval
	serfConfig.ReconnectTimeout = config.SerfReapTimeout
	serfConfig.TombstoneTimeout = config.SerfTombstoneTimeout

	node.events = make(chan serf.Event)
	serfConfig.EventCh = node.events

	serfConfig.Logger = zap.NewStdLog(logger.Desugar())
	node.serfServer, err = serf.Create(serfConfig)
	if err != nil {
		return nil, err
	}

	go node.eventHandler()
	<-node.upCh

	clusterNodes, err := node.nodefinder.GetNodes()
	if err != nil {
		return nil, err
	}

	joined, err := node.serfServer.Join(clusterNodes, true)
	if err != nil {
		return nil, err
	}

	node.logger.Infof("%d servers joined", joined)

	return node, err
}

// DefaultConfig returns a configuration for a one node cluster
// and is a good starting point for additional nodes. If the nodes
// are running on different IPs or addresses then there is no modification
// required in most of the cases.
func DefaultConfig() Config {
	return Config{
		SerfPort:             defaultSerfPort,
		SerfReapTimeout:      defaultReapTimeout,
		SerfReapInterval:     defaultReapInterval,
		SerfTombstoneTimeout: defaultTombstoneTimeout,

		NSQDPort:                defaultNSQDPort,
		NSQLookupPort:           defaultNSQLookupPort,
		NSQDListenHTTPPort:      defaultNSQDListenHTTPPort,
		NSQLookupListenHTTPPort: defaultNSQLookupListenHTTPPort,

		Nodefinder: NewNodefinderStatic(nil),
	}
}

func (node *Node) Stop() error {

	// stop serf
	err := node.serfServer.Leave()
	if err != nil {
		return err
	}

	return node.serfServer.Shutdown()
}

func (node *Node) eventHandler() {

	for e := range node.events {

		switch e.EventType() {
		default:
		case serf.EventMemberJoin:
			if memberEvent, ok := e.(serf.MemberEvent); ok {
				for _, member := range memberEvent.Members {
					if node.serfServer.LocalMember().Name == member.Name {
						node.upCh <- true
						continue
					}
					node.logger.Infof("member joined")
				}
			}
		case serf.EventMemberFailed:
			for _, member := range e.(serf.MemberEvent).Members {
				if node.serfServer.LocalMember().Name == member.Name {
					continue
				}
				node.logger.Warnf("member %s failed", member.Name)
			}
		case serf.EventMemberLeave:
			for _, member := range e.(serf.MemberEvent).Members {
				if node.serfServer.LocalMember().Name == member.Name {
					continue
				}
				node.logger.Warnf("member %s left", member.Name)
			}
		case serf.EventMemberReap:
			for _, member := range e.(serf.MemberEvent).Members {
				if node.serfServer.LocalMember().Name == member.Name {
					continue
				}
				node.logger.Warnf("member %s reaped", member.Name)
			}
		}
	}

}
