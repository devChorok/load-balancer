package algorithms

import (
	"log"
	"sync"
	"time"

	"github.com/devChorok/load-balancer/pkg/types"
)

type WeightedRoundRobin struct {
	nodes     []*types.Node
	index     int
	weight    int // Current weight being used to select nodes
	maxWeight int // Track the maximum weight
	mu        sync.Mutex
}

func NewWeightedRoundRobin(nodes []*types.Node) *WeightedRoundRobin {
	maxWeight := 0
	for _, node := range nodes {
		if node.Weight > maxWeight {
			maxWeight = node.Weight
		}
	}
	log.Printf("Initializing WeightedRoundRobin with maxWeight: %d", maxWeight)
	return &WeightedRoundRobin{
		nodes:     nodes,
		index:     -1,
		weight:    maxWeight,
		maxWeight: maxWeight,
	}
}

// Attempt to find a node with a retry mechanism
func (wrr *WeightedRoundRobin) NextNodeWithRetry(contentLength int64, maxRetries int) *types.Node {
	for i := 0; i < maxRetries; i++ {
		node := wrr.NextNode(contentLength)
		if node != nil {
			return node
		}
		log.Printf("Retrying to find a suitable node, attempt %d", i+1)
		time.Sleep(10 * time.Millisecond) // Small delay before retrying
	}
	log.Printf("Failed to find a suitable node after %d retries", maxRetries)
	return nil
}

func (wrr *WeightedRoundRobin) NextNode(contentLength int64) *types.Node {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	attempts := 0 // Track the number of attempts to find a node
	for attempts < len(wrr.nodes) { // Limit attempts to the number of nodes
		wrr.index = (wrr.index + 1) % len(wrr.nodes)
		selectedNode := wrr.nodes[wrr.index]

		log.Printf("Trying node %s with weight %d", selectedNode.Address, selectedNode.Weight)

		// Check if the node can handle the request based on its BPM and RPM limits
		if selectedNode.Weight >= wrr.weight &&
			selectedNode.CurrentRPM < selectedNode.RPM &&
			selectedNode.CurrentBPM+contentLength <= selectedNode.BPM {

			// Node can handle the request
			selectedNode.CurrentRPM++
			selectedNode.CurrentBPM += contentLength
			selectedNode.LastRequest = time.Now()

			log.Printf("Selected node %s for request. CurrentRPM: %d, CurrentBPM: %d, Weight: %d",
				selectedNode.Address, selectedNode.CurrentRPM, selectedNode.CurrentBPM, selectedNode.Weight)

			return selectedNode
		}

		attempts++ // Increment attempts after each check
	}

	log.Printf("No suitable node found after %d attempts", attempts)
	return nil
}

func (wrr *WeightedRoundRobin) CompleteRequest(node *types.Node, contentLength int64) {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	// Decrement the CurrentRPM to simulate the completion of the request
	if node.CurrentRPM > 0 {
		node.CurrentRPM--
	}
	// Decrement CurrentBPM to reflect the total bytes processed
	if node.CurrentBPM > contentLength {
		node.CurrentBPM -= contentLength
	} else {
		node.CurrentBPM = 0
	}

	// Log the current state for monitoring
	log.Printf("Node %s: Request completed. CurrentRPM: %d, CurrentBPM: %d", node.Address, node.CurrentRPM, node.CurrentBPM)
}

// Improve weight management logic
func (wrr *WeightedRoundRobin) AdjustWeight() {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	maxWeight := 0
	for _, node := range wrr.nodes {
		if node.Weight > maxWeight {
			maxWeight = node.Weight
		}
	}

	// Log weight adjustment
	if wrr.weight != maxWeight {
		log.Printf("Adjusting weight from %d to %d", wrr.weight, maxWeight)
	}

	wrr.weight = maxWeight
}
