package algorithms

import "github.com/devChorok/load-balancer/pkg/types"

type LeastLoaded struct {
    nodes []*types.Node
}

func NewLeastLoaded(nodes []*types.Node) *LeastLoaded {
    return &LeastLoaded{nodes: nodes}
}

func (ll *LeastLoaded) NextNode() *types.Node {
    minLoad := ll.nodes[0]
    for _, node := range ll.nodes[1:] {
        if node.CurrentBPM < minLoad.CurrentBPM {
            minLoad = node
        }
    }
    return minLoad
}
