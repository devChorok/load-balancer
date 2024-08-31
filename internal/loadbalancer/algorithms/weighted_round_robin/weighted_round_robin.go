package algorithms

import (
    "sync"

    "github.com/devChorok/load-balancer/pkg/types"
)

type WeightedRoundRobin struct {
    nodes  []*types.Node
    index  int
    weight int
    mu     sync.Mutex
}

func NewWeightedRoundRobin(nodes []*types.Node) *WeightedRoundRobin {
    return &WeightedRoundRobin{
        nodes:  nodes,
        index:  -1,
        weight: 0,
    }
}

func (wrr *WeightedRoundRobin) NextNode() *types.Node {
    wrr.mu.Lock()
    defer wrr.mu.Unlock()

    for {
        wrr.index = (wrr.index + 1) % len(wrr.nodes)
        if wrr.index == 0 {
            wrr.weight--
            if wrr.weight <= 0 {
                maxWeight := wrr.getMaxWeight()
                wrr.weight = maxWeight
                if maxWeight == 0 {
                    return nil
                }
            }
        }

        if wrr.nodes[wrr.index].Weight >= wrr.weight {
            selectedNode := wrr.nodes[wrr.index]
            selectedNode.CurrentRPM++ // Increment the CurrentRPM to reflect the added load
            return selectedNode
        }
    }
}

func (wrr *WeightedRoundRobin) getMaxWeight() int {
    maxWeight := 0
    for _, node := range wrr.nodes {
        if node.Weight > maxWeight {
            maxWeight = node.Weight
        }
    }
    return maxWeight
}

func (wrr *WeightedRoundRobin) CompleteRequest(node *types.Node) {
    wrr.mu.Lock()
    defer wrr.mu.Unlock()

    if node.CurrentRPM > 0 {
        node.CurrentRPM-- // Decrement the CurrentRPM to reflect the completed request
    }
}
