package bus

type NodefinderStatic struct {
	nodes []string
}

func NewNodefinderStatic(nodes []string) *NodefinderStatic {
	if len(nodes) == 0 {
		nodes = make([]string, 0)
		nodes = append(nodes, "127.0.0.1:4160")
	}

	return &NodefinderStatic{
		nodes: nodes,
	}
}

func (nfs *NodefinderStatic) GetNodes() ([]string, error) {
	return nfs.nodes, nil
}
