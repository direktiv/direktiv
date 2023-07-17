package cluster

import "time"

const (
	defaultSerfPort                = 5222
	defaultReapTimeout             = 5 * time.Minute
	defaultReapInterval            = 10 * time.Second
	defaultTombstoneTimeout        = 6 * time.Hour
	defaultNSQDPort                = 4150
	defaultNSQLookupPort           = 4151
	defaultNSQDListenHTTPPort      = 4152
	defaultNSQLookupListenHTTPPort = 4153
	defaultNSQInactiveTimeout      = 300 * time.Second
	defaultNSQTombstoneTimeout     = 45 * time.Second
	nsqLookupAddress               = "NSQD_LOOKUP_ADDRESS"
	concurrencyHandlers            = 50
	writeTimeout                   = 5 * time.Second
)

var (
	busClientLogEnabled = false
	busLogEnabled       = false
	serfLogEnabled      = false
)

// Config is the configuration structure for one node.
type Config struct {
	// Port for the serf server.
	SerfPort int

	// SerfReapTimeout: Time after other nodes are getting reaped
	// if they have been marked as failed.
	SerfReapTimeout time.Duration

	// SerfReapInterval: Interval how often the reaper runs.
	SerfReapInterval time.Duration

	// SerfTombstoneTimeout: Time after a node is getting
	// removed after it has left the cluster.
	SerfTombstoneTimeout time.Duration

	// nsq settings
	NSQDPort, NSQLookupPort,
	NSQLookupListenHTTPPort, NSQDListenHTTPPort int

	// timeouts
	NSQInactiveTimeout  time.Duration
	NSQTombstoneTimeout time.Duration

	// Topics to handle. They are getting created on startup.
	NSQTopics []string

	NSQDataDir string

	// NOTE:
	// LookupdPollInterval = time.Millisecond * 100
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
	}
}
