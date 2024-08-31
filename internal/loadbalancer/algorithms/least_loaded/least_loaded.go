package algorithms

import (
    "sync"

    "github.com/devChorok/load-balancer/pkg/types"
)

type LeastLoaded struct {
    nodes []*types.Node
    mu    sync.Mutex
}

func NewLeastLoaded(nodes []*types.Node) *LeastLoaded {
    return &LeastLoaded{nodes: nodes}
}

func (ll *LeastLoaded) NextNode() *types.Node {
    ll.mu.Lock()
    defer ll.mu.Unlock()

    minLoad := ll.nodes[0]
    for _, node := range ll.nodes[1:] {
        if node.CurrentBPM < minLoad.CurrentBPM {
            minLoad = node
        }
    }

    // Increment the CurrentBPM to indicate that this node is handling more load
    minLoad.CurrentBPM++

    return minLoad
}

func (ll *LeastLoaded) CompleteRequest(node *types.Node) {
    ll.mu.Lock()
    defer ll.mu.Unlock()

    // Decrement the CurrentBPM to simulate the completion of the request
    if node.CurrentBPM > 0 {
        node.CurrentBPM--
    }
}
