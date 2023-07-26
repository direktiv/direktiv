package cluster

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Nodefinders have to return all ips/addr of the serf nodes available on startup
// the minimum is to return one ip/addr for serf to form a cluster.
type NodeFinder interface {
	GetNodes() ([]string, error)
	GetAddr() (string, error)
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

func (nfs *nodeFinderStatic) GetNodes() ([]string, error) {
	return nfs.nodes, nil
}

func (nfs *nodeFinderStatic) GetAddr() (string, error) {
	return os.Hostname()
}

var direktivNamespace = os.Getenv("DIREKTIV_NAMESPACE")

type nodeFinderKube struct{}

// NewNodeFinderKube returns a dynamic list of nodes found in a kubernetes environment.
func NewNodeFinderKube() NodeFinder {
	return &nodeFinderKube{}
}

func (nfk *nodeFinderKube) GetNodes() ([]string, error) {
	nodes := make([]string, 0)
	ips, err := net.LookupIP(fmt.Sprintf("direktiv-headless.%s.svc", direktivNamespace))
	if err != nil {
		return nil, err
	}
	for _, ip := range ips {
		ipNew := fmt.Sprintf("%s.%s.pod:%d", strings.ReplaceAll(ip.String(), ".", "-"), direktivNamespace, defaultSerfPort)
		nodes = append(nodes, ipNew)
	}

	return nodes, nil
}

func (nfk *nodeFinderKube) GetAddr() (string, error) {
	return os.Hostname()
}
