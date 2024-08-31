package algorithms

import (
	"log"
	"sync"
	"time"

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

func (rr *RoundRobin) NextNode(contentLength int64) *types.Node {
    rr.mu.Lock()
    defer rr.mu.Unlock()

    // We start from the current index and try to find an eligible node
    originalIndex := rr.index

    for {
        rr.index = (rr.index + 1) % len(rr.nodes)
        selectedNode := rr.nodes[rr.index]

        // Check if the node can handle another request based on RPM and BPM limits
        if selectedNode.CurrentRPM < selectedNode.RPM && selectedNode.CurrentBPM+contentLength <= selectedNode.BPM {
            // Increment the CurrentRPM to indicate this node is handling one more request
            selectedNode.CurrentRPM++
            return selectedNode
        }

        // If we've looped back to the original index, no nodes can handle the request
        if rr.index == originalIndex {
            break
        }
    }

    // Return nil if no node can handle the request
    return nil
}

// Attempt to find a node with a retry mechanism
func (rr *RoundRobin) NextNodeWithRetry(contentLength int64, maxRetries int) *types.Node {
    for i := 0; i < maxRetries; i++ {
        node := rr.NextNode(contentLength)
        if node != nil {
            return node
        }
        time.Sleep(10 * time.Millisecond) // Small delay before retrying
    }
    return nil
}

func (rr *RoundRobin) CompleteRequest(node *types.Node, contentLength int64) {
    rr.mu.Lock()
    defer rr.mu.Unlock()

    // Decrement the CurrentRPM to simulate the completion of the request
    if node.CurrentRPM > 0 {
        node.CurrentRPM--
    }

    // Update CurrentBPM to reflect the total bytes processed
    node.CurrentBPM += contentLength

    // Log the current state for monitoring
    log.Printf("Node %s: Request completed. CurrentRPM: %d, CurrentBPM: %d", node.Address, node.CurrentRPM, node.CurrentBPM)
}
