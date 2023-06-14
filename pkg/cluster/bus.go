package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/direktiv/direktiv/pkg/dlog"
	"github.com/nsqio/nsq/nsqd"
	"github.com/nsqio/nsq/nsqlookupd"
	"go.uber.org/zap"
)

type bus struct {
	logger *zap.SugaredLogger

	nsqd   *nsqd.NSQD
	lookup *nsqlookupd.NSQLookupd

	dataDir string

	config Config
}

func newBus(config Config) (*bus, error) {

	logger, err := dlog.ApplicationLogger("bus")
	if err != nil {
		return nil, err
	}

	// create data dir if it does not exist
	// if not set we use a tmp folder
	dataDir := config.DataDir
	if dataDir == "" {
		dir, err := os.MkdirTemp(os.TempDir(), "nsq")
		if err != nil {
			return nil, err
		}
		dataDir = dir
	}

	logger.Infof("using %s as nsq data dir", dataDir)

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		err := os.MkdirAll(dataDir, 0700)
		if err != nil {
			return nil, err
		}
	}

	bus := &bus{
		logger:  logger,
		dataDir: dataDir,
		config:  config,
	}

	opts := nsqd.NewOptions()

	// addr, err := config.Nodefinder.GetAddr()
	// if err != nil {
	// 	return nil, err
	// }

	// h, err := hashstructure.Hash(addr, hashstructure.FormatV2, nil)
	// if err != nil {
	// 	return nil, err
	// }

	// opts.ID = int64(h)
	opts.DataPath = dataDir
	opts.MemQueueSize = 100
	opts.LogLevel = 1
	// opts.NSQLookupdTCPAddresses = nodes

	opts.TCPAddress = net.JoinHostPort(net.IPv4zero.String(),
		fmt.Sprintf("%d", config.NSQDPort))
	opts.HTTPAddress = net.JoinHostPort(net.IPv4zero.String(),
		fmt.Sprintf("%d", config.NSQDListenHTTPPort))

	bus.nsqd, err = nsqd.New(opts)
	if err != nil {
		return nil, err
	}

	lookupOptions := nsqlookupd.NewOptions()
	lookupOptions.TCPAddress = net.JoinHostPort(net.IPv4zero.String(),
		fmt.Sprintf("%d", config.NSQLookupPort))
	lookupOptions.HTTPAddress = net.JoinHostPort(net.IPv4zero.String(),
		fmt.Sprintf("%d", config.NSQLookupListenHTTPPort))

	bus.lookup, err = nsqlookupd.New(lookupOptions)
	if err != nil {
		return nil, err
	}

	return bus, nil

}

func (b *bus) Stop() {
	if b.nsqd != nil {
		b.nsqd.Exit()
	}
	if b.lookup != nil {
		b.lookup.Exit()
	}
}

type ProducerList struct {
	Producers []struct {
		RemoteAddress    string `json:"remote_address"`
		Hostname         string `json:"hostname"`
		BroadcastAddress string `json:"broadcast_address"`
		TCPPort          int    `json:"tcp_port"`
		HTTPPort         int    `json:"http_port"`
		Version          string `json:"version"`
		Tombstones       []any  `json:"tombstones"`
		Topics           []any  `json:"topics"`
	} `json:"producers"`
}

func (b *bus) Nodes() (*ProducerList, error) {

	url := fmt.Sprintf("http://127.0.0.1:%d/nodes", b.config.NSQLookupListenHTTPPort)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bo, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pl ProducerList
	err = json.Unmarshal(bo, &pl)
	if err != nil {
		return nil, err
	}

	return &pl, nil

}

func (b *bus) UpdateBusNodes(nodes []string) error {

	url := fmt.Sprintf("http://127.0.0.1:%d/config/nsqlookupd_tcp_addresses", b.config.NSQDListenHTTPPort)

	data, err := json.Marshal(nodes)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil

}

func (b *bus) Start() error {
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

func (b *bus) WaitTillConnected() error {
	startCh := make(chan bool)
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:

				url := fmt.Sprintf("http://127.0.0.1:%d/ping", b.config.NSQDListenHTTPPort)
				resp, err := http.Get(url)
				if err != nil {
					continue
				}
				defer resp.Body.Close()

				bo, err := io.ReadAll(resp.Body)
				if err != nil {
					continue
				}

				if string(bo) == "OK" {
					startCh <- true
					return
				}

			case <-time.After(60 * time.Second):
				startCh <- false
				return
			}
		}
	}()

	success := <-startCh
	if !success {
		return fmt.Errorf("could not start nsq bus")
	}
	return nil
}

func (b *bus) ModifyTopic(topic string, create bool) error {

	url := fmt.Sprintf("http://127.0.0.1:%d/topic/create?topic=%s", b.config.NSQDListenHTTPPort, topic)

	m := http.MethodPost
	if !create {
		m = http.MethodDelete
	}

	req, err := http.NewRequest(m, url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (b *bus) CreateChannel(topic, channel string) error {

	url := fmt.Sprintf("http://127.0.0.1:%d/channel/create?topic=%s&channel=%s", b.config.NSQDListenHTTPPort, topic, channel)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
