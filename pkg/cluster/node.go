// Package cluster manages serf nodes as a cluster in Direktiv.
// In particular it manages the underlying nsq cluster and is used
// to add and remove nodes dynamically in particular in Kubernetes environments.
package cluster

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/hashicorp/serf/serf"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"
)

// Node is configured per host in the cluster.
type Node struct {
	logger *zap.SugaredLogger

	// serf settings
	serfServer *serf.Serf
	events     chan serf.Event
	nodefinder NodeFinder
	upCh       chan bool

	// nsq settings
	bus            *bus
	busChannelName string
	producer       *nsq.Producer
}

func NewNode(config Config, nodeFinder NodeFinder, logger *zap.SugaredLogger) (*Node, error) {
	var err error

	if logger == nil {
		logger = zap.NewNop().Sugar()
	}

	if nodeFinder == nil {
		panic(fmt.Errorf("nodefinder not set"))
	}

	node := &Node{
		logger:     logger,
		nodefinder: nodeFinder,
		upCh:       make(chan bool),
	}

	node.bus, err = newBus(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create nsq bus: %w", err)
	}
	go func() {
		e := node.bus.start()
		if e != nil {
			panic("can not start nsq bus")
		}
	}()

	err = node.bus.waitTillConnected()
	if err != nil {
		return nil, fmt.Errorf("failed to start nsq bus: %w", err)
	}

	producerConfig := nsq.NewConfig()
	node.producer, err = nsq.NewProducer(fmt.Sprintf("127.0.0.1:%d", config.NSQDPort), producerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create nsq producer client: %w", err)
	}

	bl := &busLogger{
		logger: node.logger,
		debug:  busClientLogEnabled,
	}

	node.producer.SetLogger(bl, nsq.LogLevelWarning)

	serfConfig := serf.DefaultConfig()
	serfConfig.Init()

	serfConfig.Tags = make(map[string]string)

	addr, err := nodeFinder.GetAddr()
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}
	serfConfig.NodeName = net.JoinHostPort(addr, fmt.Sprintf("%d", config.SerfPort))

	serfConfig.Tags = make(map[string]string)
	serfConfig.Tags[nsqLookupAddress] = net.JoinHostPort(addr,
		fmt.Sprintf("%d", config.NSQLookupPort))

	hash, err := hashstructure.Hash(serfConfig.NodeName, hashstructure.FormatV2, nil)
	if err != nil {
		panic(err)
	}
	node.busChannelName = fmt.Sprintf("%d", hash)

	serfConfig.MemberlistConfig.BindAddr = net.IPv4zero.String()
	serfConfig.MemberlistConfig.BindPort = config.SerfPort

	serfConfig.ReapInterval = config.SerfReapInterval
	serfConfig.ReconnectTimeout = config.SerfReapTimeout
	serfConfig.TombstoneTimeout = config.SerfTombstoneTimeout

	node.events = make(chan serf.Event)
	serfConfig.EventCh = node.events

	loggerDiscard := log.New(io.Discard, "", log.LstdFlags)

	serfConfig.MemberlistConfig.Logger = loggerDiscard
	serfConfig.Logger = loggerDiscard

	if serfLogEnabled {
		serfConfig.Logger = zap.NewStdLog(logger.Desugar())
		serfConfig.MemberlistConfig.Logger = zap.NewStdLog(logger.Desugar())
	}

	node.serfServer, err = serf.Create(serfConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create serf node: %w", err)
	}

	go node.eventHandler()
	<-node.upCh

	clusterNodes, err := node.nodefinder.GetNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to find nodes: %w", err)
	}

	joined, err := node.serfServer.Join(clusterNodes, true)
	if err != nil {
		return nil, fmt.Errorf("failed to join serf cluster: %w", err)
	}

	node.logger.Infof("Cluster with %d servers joined.", joined)

	return node, err
}

// Stop attempts to gracefully leave the cluster, notifying other nodes that
// we are leaving.
func (node *Node) Stop() error {
	node.logger.Infof("Stopping node: %v.", node.serfServer.LocalMember().Addr)

	if node.bus != nil {
		node.bus.stop()
	}

	if node.producer != nil {
		node.producer.Stop()
	}

	err := node.serfServer.Leave()
	if err != nil {
		return err
	}

	shutdownCh := node.serfServer.ShutdownCh()

	err = node.serfServer.Shutdown()
	if err != nil {
		return err
	}

	<-shutdownCh

	return nil
}

func (node *Node) updateBusMember() error {
	members := node.serfServer.Members()
	updateBusMember := make([]string, 0)
	for i := range members {
		m := members[i]
		updateBusMember = append(updateBusMember, m.Tags[nsqLookupAddress])
	}

	err := node.bus.updateBusNodes(updateBusMember)
	if err != nil {
		return err
	}

	return nil
}

func (node *Node) handleMember(memberEvent serf.MemberEvent, join bool) {
	for _, member := range memberEvent.Members {
		if node.serfServer.LocalMember().Name == member.Name {
			if join {
				node.upCh <- true
			}

			continue
		}

		err := node.updateBusMember()
		if err != nil {
			panic(fmt.Errorf("can not handle member join (update): %w", err))
		}
	}
}

func (node *Node) eventHandler() {
	for e := range node.events {
		switch e.EventType() {
		default:
		case serf.EventMemberUpdate, serf.EventUser, serf.EventQuery:
			if memberEvent, ok := e.(serf.MemberEvent); ok {
				node.logger.Debugf("member event: %s", memberEvent.String())
			}
		case serf.EventMemberJoin:
			if memberEvent, ok := e.(serf.MemberEvent); ok {
				node.handleMember(memberEvent, true)
				node.logger.Infof("A node has joined the cluster: %v.", memberEvent.Members)
			}
		case serf.EventMemberReap:
			if memberEvent, ok := e.(serf.MemberEvent); ok {
				node.logger.Infof("A node has left the cluster: %v.", memberEvent.Members)
			}

			fallthrough
		case serf.EventMemberFailed:
			fallthrough
		case serf.EventMemberLeave:
			if memberEvent, ok := e.(serf.MemberEvent); ok {
				node.logger.Debugf("A node has left the cluster: %v.", memberEvent.Members)
				node.handleMember(memberEvent, false)
			}
		}
	}
}

func (node *Node) doSubscribe(topic, channel string,
	handler func(m []byte) error,
) (*messageConsumer, error) {
	config := nsq.NewConfig()
	config.MaxInFlight = 100
	config.MsgTimeout = time.Minute
	config.OutputBufferTimeout = time.Second
	config.WriteTimeout = writeTimeout

	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create new consumer: %w", err)
	}

	bl := &busLogger{
		logger: node.logger,
		debug:  busClientLogEnabled,
	}
	consumer.SetLogger(bl, 1)

	mh := &messageConsumer{
		topic:    topic,
		consumer: consumer,
		executor: handler,
	}

	consumer.AddConcurrentHandlers(mh, concurrencyHandlers)

	err = consumer.ConnectToNSQLookupd(fmt.Sprintf("127.0.0.1:%d",
		node.bus.config.NSQLookupListenHTTPPort))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nsqlookupd: %w", err)
	}

	return mh, nil
}
