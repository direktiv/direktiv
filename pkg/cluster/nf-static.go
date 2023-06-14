package cluster

import (
	"fmt"
	"os"
)

type NodefinderStatic struct {
	nodes []string
}

func NewNodefinderStatic(nodes []string) *NodefinderStatic {
	if len(nodes) == 0 {
		nodes = make([]string, 0)
		nodes = append(nodes, fmt.Sprintf("127.0.0.1:%d", defaultSerfPort))
	}

	return &NodefinderStatic{
		nodes: nodes,
	}
}

func (nfs *NodefinderStatic) GetNodes() ([]string, error) {
	return nfs.nodes, nil
}

func (nfs *NodefinderStatic) GetAddr() (string, error) {
	return os.Hostname()
}
