package algorithms

import (
    "sync"

    "github.com/devChorok/load-balancer/pkg/types"
)

type LeastConnections struct {
    nodes []*types.Node
    mu    sync.Mutex
}

func NewLeastConnections(nodes []*types.Node) *LeastConnections {
    return &LeastConnections{nodes: nodes}
}

func (lc *LeastConnections) NextNode() *types.Node {
    lc.mu.Lock()
    defer lc.mu.Unlock()

    minConnections := lc.nodes[0]
    for _, node := range lc.nodes[1:] {
        if node.CurrentRPM < minConnections.CurrentRPM {
            minConnections = node
        }
    }

    // Increment the CurrentRPM of the selected node to simulate handling one more connection
    minConnections.CurrentRPM++

    return minConnections
}

func (lc *LeastConnections) CompleteRequest(node *types.Node) {
    lc.mu.Lock()
    defer lc.mu.Unlock()

    // Decrement the CurrentRPM to simulate a connection being closed
    if node.CurrentRPM > 0 {
        node.CurrentRPM--
    }
}
