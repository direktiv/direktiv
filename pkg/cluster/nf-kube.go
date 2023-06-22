package cluster

import (
	"fmt"
	"net"
	"os"
	"strings"
)

var direktivNamespace = os.Getenv("DIREKTIV_NAMESPACE")

type NodefinderKube struct{}

func NewNodefinderKube() *NodefinderKube {
	return &NodefinderKube{}
}

func (nfk *NodefinderKube) GetNodes() ([]string, error) {
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

func (nfk *NodefinderKube) GetAddr() (string, error) {
	return os.Hostname()
}
