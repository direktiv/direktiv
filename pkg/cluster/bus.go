package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nsqio/nsq/nsqd"
	"github.com/nsqio/nsq/nsqlookupd"
	"go.uber.org/zap"
)

var busStartTimeout = 60 * time.Second

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

func (bl *busLogger) Output(_ int, s string) error {
	bl.logger.Debugf(s)

	return nil
}

func newBus(config Config, logger *zap.SugaredLogger) (*bus, error) {
	dataDir := config.NSQDataDir
	if dataDir == "" {
		dir, err := os.MkdirTemp(os.TempDir(), "nsq")
		if err != nil {
			return nil, err
		}
		dataDir = dir
	}

	logger.Infof("using %s as nsq data dir", dataDir)

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

	// Set default logger and adjust log level based on busLogEnabled
	opts.Logger = zap.NewStdLog(logger.Desugar())
	if !busLogEnabled {
		opts.LogLevel = 2 // Error level
	}

	var err error

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

	// Set default logger and adjust log level based on busLogEnabled
	lookupOptions.Logger = zap.NewStdLog(logger.Desugar())
	if !busLogEnabled {
		lookupOptions.LogLevel = 2 // Error level
	}

	bus.lookup, err = nsqlookupd.New(lookupOptions)
	if err != nil {
		return nil, err
	}

	return bus, nil
}

func (b *bus) stop() {
	b.logger.Debug("Stopping nsqd", b.nsqd == nil, b.lookup == nil)

	if b.nsqd != nil {
		b.nsqd.Exit()
	}

	b.logger.Debug("Stopping nsqd lookup")
	if b.lookup != nil {
		b.lookup.Exit()
	}
}

type producerList struct {
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

//lint:ignore U1000 Ignore unused function for testing
func (b *bus) nodes() (*producerList, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		fmt.Sprintf("http://127.0.0.1:%d/nodes",
			b.config.NSQLookupListenHTTPPort), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bo, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pl producerList
	err = json.Unmarshal(bo, &pl)
	if err != nil {
		return nil, err
	}

	var s string
	for _, n := range pl.Producers {
		s += "," + n.Hostname + ":" + fmt.Sprintf("%v", n.TCPPort)
	}

	return &pl, nil
}

func (b *bus) updateBusNodes(ctx context.Context, nodes []string, client *http.Client) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	b.logger.Debugf("Updating bus nodes: %s", strings.Join(nodes, ", "))

	url := fmt.Sprintf("http://127.0.0.1:%d/config/nsqlookupd_tcp_addresses", b.config.NSQDListenHTTPPort)

	data, err := json.Marshal(nodes)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(data)) //nolint
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("can not set new bus members")
	}

	return nil
}

func (b *bus) start(ctx context.Context, timeout time.Duration) error {
	// Create a new context with a timeout of 10 seconds
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	errChanNsqd := make(chan error, 1)
	errChanLookup := make(chan error, 1)
	go func() {
		err := b.nsqd.Main()
		errChanNsqd <- err
	}()

	go func() {
		err := b.lookup.Main()
		errChanLookup <- err
	}()

	select {
	case errNsqd := <-errChanNsqd:
		if errNsqd != nil {
			return errNsqd
		}
	case errLookup := <-errChanLookup:
		if errLookup != nil {
			return errLookup
		}
	case <-ctx.Done():
		// Cancel both nsqd and lookup in case of timeout
		b.nsqd.Exit()
		b.lookup.Exit()
		return ctx.Err()
	}

	return nil
}

func (b *bus) waitTillConnected(ctx context.Context, client *http.Client, tickerTime, timeout time.Duration) error {
	checkService := func(port int) bool {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/ping", port), nil)
		if err != nil {
			return false
		}

		resp, err := client.Do(req)
		if err != nil {
			return false
		}
		defer resp.Body.Close()

		return resp.StatusCode == http.StatusOK
	}

	busStarted := make(chan struct{})
	go func() {
		for {
			if checkService(b.config.NSQDListenHTTPPort) && checkService(b.config.NSQLookupListenHTTPPort) {
				close(busStarted)

				return
			}

			select {
			case <-time.After(tickerTime):
			case <-ctx.Done():

				return
			}
		}
	}()

	select {
	case <-time.After(timeout):
		return fmt.Errorf("timed out waiting for nsq bus to start")
	case <-busStarted:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

//nolint:unused
func (b *bus) createTopic(ctx context.Context, topic string, client *http.Client) error {
	url := fmt.Sprintf("http://127.0.0.1:%d/topic/create?topic=%s", b.config.NSQDListenHTTPPort, topic)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create channel failed")
	}

	return nil
}

//nolint:unused
func (b *bus) createDeleteChannel(ctx context.Context, client *http.Client, topic, channel string, create bool) error {
	action := "delete"
	if create {
		action = "create"
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/channel/%s?topic=%s&channel=%s", b.config.NSQDListenHTTPPort, action, topic, channel)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create channel failed")
	}

	return nil
}
