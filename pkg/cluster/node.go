// Package cluster manages serf nodes as a cluster in Direktiv.
// In particular it manages the underlying nsq cluster and is used
// to add and remove nodes dynamically in particular in Kubernetes environments.
package cluster

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/dlog"
	"github.com/hashicorp/serf/serf"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/nsqio/go-nsq"
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
	defaultNSQInactiveTimeout      = 300 * time.Second
	defaultNSQTombstoneTimeout     = 45 * time.Second

	nsqLookupAddress = "NSQD_LOOKUP_ADDRESS"

	sharedChannel = "shared"

	concurrencyHandlers = 50
	writeTimeout        = 5 * time.Second
)

var (
	busClientLogEnabled = true
	busLogEnabled       = false
	serfLogEnabled      = false
)

// Config is the configuration structure for one node.
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

	// timeouts
	NSQInactiveTimeout  time.Duration
	NSQTombstoneTimeout time.Duration

	// Topics to handle. They are getting created on startup.
	NSQTopics []string

	NSQDataDir string
}

// Node is configured per host in the cluster.
type Node struct {
	logger *zap.SugaredLogger

	// serf settings
	serfServer *serf.Serf
	events     chan serf.Event

	nodefinder Nodefinder

	upCh chan bool

	// nsq settings
	bus            *bus
	BusChannelName string
	producer       *nsq.Producer
}

// Nodefinders have to return all ips/addr of the serf nodes available on startup
// the minimum is to return one ip/addr for serf to form a cluster.
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

	// starting the underlying bus. if it does not start, we can panic
	node.bus, err = newBus(config)
	if err != nil {
		return nil, err
	}
	go func() {
		e := node.bus.start()
		if e != nil {
			panic("can not start nsq bus")
		}
	}()

	err = node.bus.waitTillConnected()
	if err != nil {
		panic("can not start nsq bus")
	}

	producerConfig := nsq.NewConfig()
	node.producer, err = nsq.NewProducer(fmt.Sprintf("127.0.0.1:%d", config.NSQDPort), producerConfig)
	if err != nil {
		return nil, err
	}

	bl := &busLogger{
		logger: node.logger,
		debug:  busClientLogEnabled,
	}

	node.producer.SetLogger(bl, 1)

	serfConfig := serf.DefaultConfig()
	serfConfig.Init()

	serfConfig.Tags = make(map[string]string)

	addr, err := config.Nodefinder.GetAddr()
	if err != nil {
		return nil, err
	}
	serfConfig.NodeName = net.JoinHostPort(addr, fmt.Sprintf("%d", config.SerfPort))

	serfConfig.Tags = make(map[string]string)
	serfConfig.Tags[nsqLookupAddress] = net.JoinHostPort(addr,
		fmt.Sprintf("%d", config.NSQLookupPort))

	hash, err := hashstructure.Hash(serfConfig.NodeName, hashstructure.FormatV2, nil)
	if err != nil {
		return nil, err
	}
	node.BusChannelName = fmt.Sprintf("%d", hash)

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
		NSQInactiveTimeout:      defaultNSQInactiveTimeout,
		NSQTombstoneTimeout:     defaultNSQTombstoneTimeout,

		Nodefinder: NewNodefinderStatic(nil),
	}
}

func (node *Node) Stop() error {

	node.logger.Infof("stopping node: %v", node.serfServer.LocalMember().Addr)

	if node.bus != nil {
		node.bus.stop()
	}

	if node.producer != nil {
		node.producer.Stop()
	}

	// stop serf
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

// prepareBus creates the configured topics and their channels.
// a shared channel for one receiver and the other one to be used to share the message
// across the cluster.
func (node *Node) prepareBus() error {
	for i := range node.bus.config.NSQTopics {
		topic := node.bus.config.NSQTopics[i]
		err := node.bus.createTopic(topic)
		if err != nil {
			node.logger.Errorf("can not create topic %s: %s", topic, err.Error())

			return err
		}
		err = node.bus.createDeleteChannel(topic, sharedChannel, true)
		if err != nil {
			node.logger.Errorf("can not create channel shared: %s", err.Error())

			return err
		}
		err = node.bus.createDeleteChannel(topic, node.BusChannelName, true)
		if err != nil {
			node.logger.Errorf("can not create channel individual: %s", err.Error())

			return err
		}
	}

	return nil
}

func (node *Node) updateBusMember() error {
	members := node.serfServer.Members()
	updateBusMember := make([]string, 0)
	for i := range members {
		m := members[i]
		updateBusMember = append(updateBusMember, m.Tags[nsqLookupAddress])
	}

	node.logger.Debugf("updating bus members: %s", strings.Join(updateBusMember, ","))

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
				err := node.prepareBus()
				if err != nil {
					node.logger.Errorf("can not prepare bus: %s", err.Error())
					panic("can not handle member join (prepare)")
				}
				node.upCh <- true
			}

			continue
		}

		node.logger.Infof("member %s: %s", member.Name, memberEvent.String())
		err := node.updateBusMember()
		if err != nil {
			node.logger.Errorf("can not prepare bus: %s", err.Error())
			panic("can not handle member join (update)")
		}
	}
}

func (node *Node) eventHandler() {
	for e := range node.events {
		switch e.EventType() {
		default:
		case serf.EventMemberUpdate, serf.EventUser, serf.EventQuery,
			serf.EventMemberLeave, serf.EventMemberFailed:
			if memberEvent, ok := e.(serf.MemberEvent); ok {
				node.logger.Infof("member event: %s", memberEvent.String())
			}
		case serf.EventMemberJoin:
			if memberEvent, ok := e.(serf.MemberEvent); ok {
				node.handleMember(memberEvent, true)
			}
		case serf.EventMemberReap:
			if memberEvent, ok := e.(serf.MemberEvent); ok {
				node.handleMember(memberEvent, false)
			}
		}
	}
}

func (node *Node) doSubscribe(topic, channel string,
	handler func(m []byte) error,
) (*MessageConsumer, error) {
	config := nsq.NewConfig()
	config.MaxInFlight = 100
	config.MsgTimeout = time.Minute
	config.OutputBufferTimeout = time.Second
	config.WriteTimeout = writeTimeout

	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return nil, err
	}

	bl := &busLogger{
		logger: node.logger,
		debug:  busClientLogEnabled,
	}
	consumer.SetLogger(bl, 1)

	mh := &MessageConsumer{
		topic:    topic,
		consumer: consumer,
		executor: handler,
	}

	consumer.AddConcurrentHandlers(mh, concurrencyHandlers)

	err = consumer.ConnectToNSQLookupd(fmt.Sprintf("127.0.0.1:%d",
		node.bus.config.NSQLookupListenHTTPPort))
	if err != nil {
		return nil, err
	}

	return mh, nil
}

func (node *Node) Subscribe(topic string, handler func(m []byte) error) (*MessageConsumer, error) {
	return node.doSubscribe(topic, node.BusChannelName, handler)
}

func (node *Node) SubscribeOnce(topic string, handler func(m []byte) error) (*MessageConsumer, error) {
	return node.doSubscribe(topic, sharedChannel, handler)
}

func (node *Node) Unsubscribe(messageConsumer *MessageConsumer) {
	messageConsumer.consumer.Stop()
}

func (node *Node) Publish(topic string, message []byte) error {
	return node.producer.Publish(topic, message)
}

type MessageConsumer struct {
	topic    string
	executor func(m []byte) error
	consumer *nsq.Consumer
}

func (h *MessageConsumer) HandleMessage(m *nsq.Message) error {
	return h.executor(m.Body)
}
