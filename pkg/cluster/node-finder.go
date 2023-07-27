package cluster

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// Nodefinders have to return all ips/addr of the serf nodes available on startup
// the minimum is to return one ip/addr for serf to form a cluster.
type NodeFinder interface {
	GetNodes(ctx context.Context) ([]string, error)
	GetAddr(ctx context.Context, nodeID string) (string, error)
}

type nodeFinderStatic struct {
	nodes []string
}

// NewNodeFinderStatic is used when the cluster is not dynamic (or not a cluster).
func NewNodeFinderStatic(nodes []string) NodeFinder {
	if len(nodes) == 0 {
		nodes = make([]string, 0)
		nodes = append(nodes, fmt.Sprintf("127.0.0.1:%d", defaultSerfPort))
	}

	return &nodeFinderStatic{
		nodes: nodes,
	}
}

func (nfs *nodeFinderStatic) GetNodes(ctx context.Context) ([]string, error) {
	_ = ctx

	return nfs.nodes, nil
}

func (nfs *nodeFinderStatic) GetAddr(ctx context.Context, nodeID string) (string, error) {
	_ = ctx
	_ = nodeID

	return os.Hostname()
}

var direktivNamespace = os.Getenv("DIREKTIV_NAMESPACE")

type nodeFinderKube struct{}

// NewNodeFinderKube returns a dynamic list of nodes found in a kubernetes environment.
func NewNodeFinderKube() NodeFinder {
	return &nodeFinderKube{}
}

func (nfk *nodeFinderKube) GetNodes(ctx context.Context) ([]string, error) {
	nodes := make([]string, 0)

	// Use the provided context for the DNS lookup
	ips, err := lookupIPWithContext(ctx, fmt.Sprintf("direktiv-headless.%s.svc", direktivNamespace), 4, time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to look up IP addresses: %v, %w", fmt.Sprintf("direktiv-headless.%s.svc", direktivNamespace), err)
	}

	for _, ip := range ips {
		nodeAddr := fmt.Sprintf("%s.%s.pod:%d", strings.ReplaceAll(ip.String(), ".", "-"), direktivNamespace, defaultSerfPort)
		nodes = append(nodes, nodeAddr)
	}

	return nodes, nil
}

// lookupIPWithContext is a helper function to perform DNS lookup with a provided context.
func lookupIPWithContext(ctx context.Context, host string, retries int, retryDelay time.Duration) ([]net.IP, error) {
	for i := 0; i <= retries; i++ {
		// Perform the DNS lookup using the provided context
		resolver := net.Resolver{}
		ips, err := resolver.LookupIPAddr(ctx, host)
		if err == nil {
			// Extract the IP addresses from IPAddr objects
			result := make([]net.IP, len(ips))
			for i, ipAddr := range ips {
				result[i] = ipAddr.IP
			}
			return result, nil
		}

		// If there was an error, wait for the specified retry delay before trying again
		if i < retries {
			time.Sleep(retryDelay)
		}
	}

	// Return the last error if all retries failed
	return nil, fmt.Errorf("failed to look up IP addresses after %d retries", retries)
}

func (nfk *nodeFinderKube) GetAddr(ctx context.Context, nodeID string) (string, error) {
	_ = ctx
	_ = nodeID

	return os.Hostname()
}
