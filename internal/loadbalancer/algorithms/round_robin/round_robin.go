package algorithms

import (
    "sync"

    "github.com/devChorok/load-balancer/pkg/types"
)

type RoundRobin struct {
    index int
    nodes []*types.Node
    mu    sync.Mutex
}

func NewRoundRobin(nodes []*types.Node) *RoundRobin {
    return &RoundRobin{index: -1, nodes: nodes}
}

func (rr *RoundRobin) NextNode() *types.Node {
    rr.mu.Lock()
    defer rr.mu.Unlock()

    rr.index = (rr.index + 1) % len(rr.nodes)
    selectedNode := rr.nodes[rr.index]

    // Increment the CurrentRPM to indicate this node is handling one more request
    selectedNode.CurrentRPM++

    return selectedNode
}

func (rr *RoundRobin) CompleteRequest(node *types.Node) {
    rr.mu.Lock()
    defer rr.mu.Unlock()

    // Decrement the CurrentRPM to simulate the completion of the request
    if node.CurrentRPM > 0 {
        node.CurrentRPM--
    }
}
