package bus

import (
	"fmt"
	"os"
	"time"

	"github.com/nsqio/nsq/nsqd"
	"github.com/nsqio/nsq/nsqlookupd"
)

// type Nodefinder func(args ...interface{}) ([]string, error)

type Nodefinder interface {
	GetNodes() ([]string, error)
}

type Bus struct {
	nsqd   *nsqd.NSQD
	lookup *nsqlookupd.NSQLookupd
}

type Config struct {
	QueueSize  int64
	DataPath   string
	Nodefinder Nodefinder

	NSQDListen, LookupListen, LookupListenHTTP, NSQDListenHTTP string

	PREFIX string
}

func DefaultConfig() *Config {

	nfs := NewNodefinderStatic(nil)

	return &Config{
		QueueSize:        1000,
		DataPath:         "/tmp/nsqd",
		Nodefinder:       nfs,
		NSQDListen:       "0.0.0.0:4150",
		NSQDListenHTTP:   "0.0.0.0:4160",
		LookupListen:     "0.0.0.0:4170",
		LookupListenHTTP: "0.0.0.0:4180",
	}

}

func NewBus(config *Config) (*Bus, error) {

	if config.Nodefinder == nil {
		return nil, fmt.Errorf("nodefinder not set in configuration")
	}

	err := os.MkdirAll(config.DataPath, 0700)
	if err != nil {
		return nil, err
	}

	nodes, err := config.Nodefinder.GetNodes()
	fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!! %v\n", nodes)

	opts := nsqd.NewOptions()

	opts.DataPath = config.DataPath
	opts.MemQueueSize = config.QueueSize
	opts.NSQLookupdTCPAddresses = nodes
	// opts.BroadcastAddress = "127.0.0.1"

	opts.TCPAddress = config.NSQDListen
	opts.HTTPAddress = config.NSQDListenHTTP

	// opts.HTTPAddress = ""
	opts.LogLevel = 1

	nsqd, err := nsqd.New(opts)
	if err != nil {
		return nil, err
	}

	lookupOptions := nsqlookupd.NewOptions()
	lookupOptions.TCPAddress = config.LookupListen
	lookupOptions.HTTPAddress = config.LookupListenHTTP
	lookupOptions.LogPrefix = config.PREFIX
	lookupOptions.LogLevel = 1

	// lookupOptions.BroadcastAddress = "127.0.0.1"

	lookup, err := nsqlookupd.New(lookupOptions)
	if err != nil {
		return nil, err
	}

	return &Bus{
		nsqd:   nsqd,
		lookup: lookup,
	}, nil
}

func (b *Bus) Start() error {

	errChan := make(chan error, 1)

	go func() {
		err := b.nsqd.Main()
		errChan <- err
	}()

	go func() {
		err := b.lookup.Main()
		errChan <- err
	}()

	err := <-errChan
	return err

}

func (b *Bus) WaitTillConnected(startCh chan bool) {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				status := b.Status()
				b.JJ()
				fmt.Printf(">>NODES %v\n", len(status.Nodes))
				if len(status.Nodes) > 0 {
					startCh <- true
					return
				}
			case <-time.After(60 * time.Second):
				startCh <- false
				return
			}
		}
	}()
}

func (b *Bus) Stop() {

	if b.nsqd != nil {
		b.nsqd.Exit()
	}
	if b.lookup != nil {
		b.lookup.Exit()
	}
}

type Node struct {
	RemoteAddress    string
	Hostname         string
	BroadcastAddress string
	TCPPort          int
	HTTPPort         int
	Version          string
}

type Status struct {
	Status string
	Nodes  []Node
}

func (b *Bus) JJ() {
	producers := b.lookup.DB.FindProducers("client", "", "").FilterByActive(
		300*time.Second, 0)
	for i := range producers.PeerInfo() {
		p := producers.PeerInfo()[i]
		fmt.Printf(">>>> %v\n", p)
	}
}

func (b *Bus) Status() *Status {

	status := &Status{
		Status: b.nsqd.GetHealth(),
		Nodes:  make([]Node, 0),
	}

	producers := b.lookup.DB.FindProducers("client", "", "").FilterByActive(
		300*time.Second, 0)

	for i := range producers.PeerInfo() {
		p := producers.PeerInfo()[i]

		status.Nodes = append(status.Nodes, Node{
			RemoteAddress:    p.RemoteAddress,
			Hostname:         p.Hostname,
			BroadcastAddress: p.BroadcastAddress,
			TCPPort:          p.TCPPort,
			HTTPPort:         p.HTTPPort,
			Version:          p.Version,
		})
	}

	return status
}
