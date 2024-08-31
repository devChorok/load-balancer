package algorithms

import (
	"log"
	"sync"
	"time"

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

// Attempt to find a node with a retry mechanism
func (rr *WeightedRoundRobin) NextNodeWithRetry(contentLength int64, maxRetries int) *types.Node {
    for i := 0; i < maxRetries; i++ {
        node := rr.NextNode(contentLength)
        if node != nil {
            return node
        }
        time.Sleep(10 * time.Millisecond) // Small delay before retrying
    }
    return nil
}

func (wrr *WeightedRoundRobin) NextNode(contentLength int64) *types.Node {
    wrr.mu.Lock()
    defer wrr.mu.Unlock()

    attempts := 0 // Track the number of attempts to find a node

    for attempts < len(wrr.nodes) { // Limit attempts to the number of nodes
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

        selectedNode := wrr.nodes[wrr.index]

        // Check if the node can handle the request based on its BPM and RPM limits
        if selectedNode.Weight >= wrr.weight &&
            selectedNode.CurrentRPM < selectedNode.RPM &&
            selectedNode.CurrentBPM+contentLength <= selectedNode.BPM {

            // Node can handle the request
            selectedNode.CurrentRPM++
            selectedNode.CurrentBPM += contentLength
            selectedNode.LastRequest = time.Now()
            return selectedNode
        }

        attempts++ // Increment attempts after each check
    }

    // If no node is found after checking all of them, return nil
    return nil
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

func (wrr *WeightedRoundRobin) CompleteRequest(node *types.Node, contentLength int64) {
    wrr.mu.Lock()
    defer wrr.mu.Unlock()

    // Decrement the CurrentBPM to simulate the completion of the request
	if node.CurrentRPM > 0 {
		node.CurrentRPM--
	}
    // Update CurrentBPM to reflect the total bytes processed
	node.CurrentBPM += contentLength

	// Log the current state for monitoring
	log.Printf("Node %s: Request completed. CurrentRPM: %d, CurrentBPM: %d", node.Address, node.CurrentRPM, node.CurrentBPM)

}
