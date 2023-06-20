package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
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

	mtx sync.Mutex
}

type busLogger struct {
	logger *zap.SugaredLogger
	debug  bool
}

func (bl *busLogger) Output(maxdepth int, s string) error {
	if bl.debug {
		bl.logger.Infof(s)
	}
	return nil
}

func newBus(config Config) (*bus, error) {

	logger, err := dlog.ApplicationLogger("bus")
	if err != nil {
		return nil, err
	}

	// create data dir if it does not exist
	// if not set we use a tmp folder
	dataDir := config.NSQDataDir
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

	opts.DataPath = dataDir

	opts.TCPAddress = net.JoinHostPort(net.IPv4zero.String(),
		fmt.Sprintf("%d", config.NSQDPort))
	opts.HTTPAddress = net.JoinHostPort(net.IPv4zero.String(),
		fmt.Sprintf("%d", config.NSQDListenHTTPPort))

	opts.NSQLookupdTCPAddresses = []string{
		fmt.Sprintf("127.0.0.1:%d", config.NSQLookupPort),
	}

	opts.MaxRdyCount = 10000
	opts.MemQueueSize = 10000

	opts.Logger = &busLogger{
		logger: logger,
		debug:  busLogEnabled,
	}

	bus.nsqd, err = nsqd.New(opts)
	if err != nil {
		return nil, err
	}

	lookupOptions := nsqlookupd.NewOptions()
	lookupOptions.TCPAddress = net.JoinHostPort(net.IPv4zero.String(),
		fmt.Sprintf("%d", config.NSQLookupPort))
	lookupOptions.HTTPAddress = net.JoinHostPort(net.IPv4zero.String(),
		fmt.Sprintf("%d", config.NSQLookupListenHTTPPort))

	lookupOptions.InactiveProducerTimeout = config.NSQInactiveTimeout
	lookupOptions.TombstoneLifetime = config.NSQTombstoneTimeout

	lookupOptions.Logger = &busLogger{
		logger: logger,
		debug:  busLogEnabled,
	}

	bus.lookup, err = nsqlookupd.New(lookupOptions)
	if err != nil {
		return nil, err
	}

	return bus, nil

}

func (b *bus) stop() {

	b.logger.Info("stopping nsqd")
	if b.nsqd != nil {
		b.nsqd.Exit()
	}
	b.logger.Info("stopping nsqd lookup")
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

func (b *bus) nodes() (*ProducerList, error) {

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

func (b *bus) updateBusNodes(nodes []string) error {

	b.mtx.Lock()
	defer b.mtx.Unlock()

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

	if resp.StatusCode != 200 {
		return fmt.Errorf("can not set newe bus members")
	}

	return nil

}

func (b *bus) start() error {
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

func (b *bus) waitTillConnected() error {
	startCh := make(chan bool)
	ticker := time.NewTicker(1 * time.Second)

	ping := func(port int) bool {
		url := fmt.Sprintf("http://127.0.0.1:%d/ping", port)
		resp, err := http.Get(url)
		if err != nil {
			return false
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return false
		}

		bo, err := io.ReadAll(resp.Body)
		if err != nil {
			return false
		}

		if string(bo) == "OK" {
			return true
		}

		return false
	}

	go func() {
		for {
			select {
			case <-ticker.C:

				if !ping(b.config.NSQDListenHTTPPort) {
					continue
				}

				if !ping(b.config.NSQLookupListenHTTPPort) {
					continue
				}

				startCh <- true
				return

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

func (b *bus) createTopic(topic string) error {

	url := fmt.Sprintf("http://127.0.0.1:%d/topic/create?topic=%s", b.config.NSQDListenHTTPPort, topic)

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

	if resp.StatusCode != 200 {
		return fmt.Errorf("create channel failed")
	}

	return nil
}

func (b *bus) createDeleteChannel(topic, channel string, create bool) error {

	action := "delete"
	if create {
		action = "create"
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/channel/%s?topic=%s&channel=%s", b.config.NSQDListenHTTPPort, action, topic, channel)

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

	if resp.StatusCode != 200 {
		return fmt.Errorf("create channel failed")
	}

	return nil
}
