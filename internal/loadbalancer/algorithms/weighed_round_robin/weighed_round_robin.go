package algorithms

import (
    "github.com/devChorok/load-balancer/pkg/types"
    "math/rand"
)

type WeightedRoundRobin struct {
    nodes []*types.Node
}

func NewWeightedRoundRobin(nodes []*types.Node) *WeightedRoundRobin {
    return &WeightedRoundRobin{nodes: nodes}
}

func (wrr *WeightedRoundRobin) NextNode() *types.Node {
    totalWeight := 0
    for _, node := range wrr.nodes {
        totalWeight += node.Weight
    }

    randValue := rand.Intn(totalWeight)
    for _, node := range wrr.nodes {
        if randValue < node.Weight {
            return node
        }
        randValue -= node.Weight
    }

    return wrr.nodes[0] // Fallback in case something goes wrong
}
