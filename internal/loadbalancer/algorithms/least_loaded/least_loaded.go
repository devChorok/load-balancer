package algorithms

import (
	"log"
	"sync"
	"time"

	"github.com/devChorok/load-balancer/pkg/types"
)

type LeastLoaded struct {
    nodes []*types.Node
    mu    sync.Mutex
}

func NewLeastLoaded(nodes []*types.Node) *LeastLoaded {
    return &LeastLoaded{nodes: nodes}
}

func (ll *LeastLoaded) NextNode(contentLength int64) *types.Node {
    ll.mu.Lock()
    defer ll.mu.Unlock()

    var selectedNode *types.Node
    minLoad := int64(^uint(0) >> 1) // Initialize to maximum int64 value

    for _, node := range ll.nodes {
        // Check if the node can handle another request based on RPM and BPM limits
        if node.CurrentRPM < node.RPM && node.CurrentBPM+contentLength <= node.BPM {
            // Select the node with the least current BPM (i.e., the least loaded)
            if node.CurrentBPM < minLoad {
                minLoad = node.CurrentBPM
                selectedNode = node
            }
        }
    }

    // If we found a suitable node, update its CurrentBPM and CurrentRPM
    if selectedNode != nil {
        selectedNode.CurrentBPM += contentLength
        selectedNode.CurrentRPM++
        selectedNode.LastRequest = time.Now()
        return selectedNode
    }

    // Return nil if no node can handle the request
    return nil
}

func (ll *LeastLoaded) NextNodeWithRetry(contentLength int64, maxRetries int) *types.Node {
	for i := 0; i < maxRetries; i++ {
		node := ll.NextNode(contentLength)
		if node != nil {
			return node
		}
		time.Sleep(10 * time.Millisecond) // Small delay before retrying
	}
	return nil
}



func (ll *LeastLoaded) CompleteRequest(node *types.Node, contentLength int64) {
    ll.mu.Lock()
    defer ll.mu.Unlock()

    // Decrement the CurrentBPM to simulate the completion of the request
	if node.CurrentRPM > 0 {
		node.CurrentRPM--
	}
    // Update CurrentBPM to reflect the total bytes processed
	node.CurrentBPM += contentLength

	// Log the current state for monitoring
	log.Printf("Node %s: Request completed. CurrentRPM: %d, CurrentBPM: %d", node.Address, node.CurrentRPM, node.CurrentBPM)

}
