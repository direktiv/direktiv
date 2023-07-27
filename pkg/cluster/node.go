// Package cluster manages serf nodes as a cluster in Direktiv.
// In particular it manages the underlying nsq cluster and is used
// to add and remove nodes dynamically in particular in Kubernetes environments.
package cluster

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
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
	upCh       chan bool

	// nsq settings
	bus            *bus
	busChannelName string
	producer       *nsq.Producer
	id             string
	httpClient     *http.Client
}

func NewNode(ctx context.Context,
	config Config,
	getAddr func(ctx context.Context, nodeID string) (string, error),
	getNodes func(context.Context) ([]string, error),
	timeout time.Duration,
	logger *zap.SugaredLogger,
	httpClient *http.Client,
) (*Node, error) {
	var err error

	node := &Node{
		logger:     logger,
		upCh:       make(chan bool),
		id:         uuid.NewString(),
		httpClient: httpClient,
	}

	node.bus, err = newBus(config, node.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create nsq bus: %w", err)
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

	addr, err := getAddr(ctx, node.id)
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

	startBusErrCh := make(chan error)
	go func() {
		err := node.bus.start(ctx, busStartTimeout)
		startBusErrCh <- err
	}()
	startWaitCh := make(chan struct{})

	go func() {
		err := node.bus.waitTillConnected(ctx, node.httpClient, time.Second, timeout)
		if err == nil {
			startWaitCh <- struct{}{}
		} else {
			startWaitCh <- struct{}{}
		}
	}()
	select {
	case <-startWaitCh:
		// Bus successfully started, continue execution
	case <-ctx.Done():
		return nil, fmt.Errorf("timed out waiting for nsq bus to start")
	}

	go node.eventHandler(ctx)
	<-node.upCh

	clusterNodes, err := getNodes(ctx)
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

	if node.serfServer != nil {
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
	}

	return nil
}

func (node *Node) updateBusMember(ctx context.Context) error {
	members := node.serfServer.Members()
	updateBusMember := make([]string, 0)
	for i := range members {
		m := members[i]
		updateBusMember = append(updateBusMember, m.Tags[nsqLookupAddress])
	}

	err := node.bus.updateBusNodes(ctx, updateBusMember, node.httpClient)
	if err != nil {
		return err
	}

	return nil
}

func (node *Node) handleMember(ctx context.Context, memberEvent serf.MemberEvent, join bool) error {
	for _, member := range memberEvent.Members {
		if node.serfServer.LocalMember().Name == member.Name {
			if join {
				node.upCh <- true
			}

			continue
		}

		// Skip processing if the event is not a join event
		if !join {
			continue
		}

		// Only update the bus member if it's not the local node
		if err := node.updateBusMember(ctx); err != nil {
			return fmt.Errorf("failed to update bus member: %v", err)
		}
	}
	return nil
}

func (node *Node) eventHandler(ctx context.Context) {
	for e := range node.events {
		switch e.EventType() {
		case serf.EventMemberUpdate, serf.EventUser, serf.EventQuery:
			if memberEvent, ok := e.(serf.MemberEvent); ok {
				node.logger.Debugf("member event: %s", memberEvent.String())
			}
		case serf.EventMemberJoin:
			if memberEvent, ok := e.(serf.MemberEvent); ok {
				if err := node.handleMember(ctx, memberEvent, true); err != nil {
					node.logger.Errorf("Failed to handle member join: %v", err)
					panic(err)
				}
				node.logger.Infof("A node has joined the cluster: %v.", memberEvent.Members)
			}
		case serf.EventMemberReap:
			if memberEvent, ok := e.(serf.MemberEvent); ok {
				node.logger.Infof("A node has left the cluster: %v.", memberEvent.Members)
			}
			fallthrough
		case serf.EventMemberFailed, serf.EventMemberLeave:
			if memberEvent, ok := e.(serf.MemberEvent); ok {
				node.logger.Debugf("A node has left the cluster: %v.", memberEvent.Members)
				if err := node.handleMember(ctx, memberEvent, false); err != nil {
					node.logger.Errorf("Failed to handle member leave/fail: %v", err)
					panic(err)
				}
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
	consumer.SetLookupdHttpClient(node.httpClient)
	consumer.AddConcurrentHandlers(mh, concurrencyHandlers)

	err = consumer.ConnectToNSQLookupd(fmt.Sprintf("127.0.0.1:%d",
		node.bus.config.NSQLookupListenHTTPPort))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nsqlookupd: %w", err)
	}

	return mh, nil
}
