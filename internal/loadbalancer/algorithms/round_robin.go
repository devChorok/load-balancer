package algorithms

import "github.com/devChorok/load-balancer/pkg/types"

type RoundRobin struct {
    index int
    nodes []*types.Node
}

func NewRoundRobin(nodes []*types.Node) *RoundRobin {
    return &RoundRobin{index: -1, nodes: nodes}
}

func (rr *RoundRobin) NextNode() *types.Node {
    rr.index = (rr.index + 1) % len(rr.nodes)
    return rr.nodes[rr.index]
}
