package algorithms

import (
	"log"
	"sync"
	"time"

	"github.com/devChorok/load-balancer/pkg/types"
)

type LeastConnections struct {
	nodes []*types.Node
	mu    sync.Mutex
}

func NewLeastConnections(nodes []*types.Node) *LeastConnections {
	return &LeastConnections{nodes: nodes}
}

func (lc *LeastConnections) NextNode(contentLength int64) *types.Node {
    lc.mu.Lock()
    defer lc.mu.Unlock()

    var selectedNode *types.Node
    minConnections := int(^uint(0) >> 1) // Initialize to max int value

    // Iterate over all nodes to find the one with the least active connections
    for _, node := range lc.nodes {
        // Check if the node can handle another request based on RPM and BPM limits
        if node.CurrentRPM < node.RPM && node.CurrentBPM+contentLength <= node.BPM {
            // Select the node with the fewest active connections
            if node.ActiveConnections < minConnections {
                minConnections = node.ActiveConnections
                selectedNode = node
            }
        }
    }

    // If we found a suitable node, update its fields and return it
    if selectedNode != nil {
        selectedNode.CurrentRPM++
        selectedNode.CurrentBPM += contentLength
        selectedNode.ActiveConnections++
        selectedNode.LastRequest = time.Now()
        return selectedNode
    }

    // Return nil if no node can handle the request
    return nil
}



func (lc *LeastConnections) NextNodeWithRetry(contentLength int64, maxRetries int) *types.Node {
	for i := 0; i < maxRetries; i++ {
		node := lc.NextNode(contentLength)
		if node != nil {
			return node
		}
		time.Sleep(10 * time.Millisecond) // Small delay before retrying
	}
	return nil
}

func (lc *LeastConnections) CompleteRequest(node *types.Node, contentLength int64) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Decrement the CurrentRPM to simulate a connection being closed
	if node.CurrentRPM > 0 {
		node.CurrentRPM--
	}
	// Update CurrentBPM to reflect the total bytes processed
	node.CurrentBPM += contentLength

	// Log the current state for monitoring
	log.Printf("Node %s: Request completed. CurrentRPM: %d, CurrentBPM: %d", node.Address, node.CurrentRPM, node.CurrentBPM)
}
