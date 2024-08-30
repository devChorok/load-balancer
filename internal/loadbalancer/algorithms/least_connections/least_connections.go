package algorithms

import "github.com/devChorok/load-balancer/pkg/types"

type LeastConnections struct {
    nodes []*types.Node
}

func NewLeastConnections(nodes []*types.Node) *LeastConnections {
    return &LeastConnections{nodes: nodes}
}

func (lc *LeastConnections) NextNode() *types.Node {
    minConnections := lc.nodes[0]
    for _, node := range lc.nodes[1:] {
        if node.CurrentRPM < minConnections.CurrentRPM {
            minConnections = node
        }
    }
    return minConnections
}
