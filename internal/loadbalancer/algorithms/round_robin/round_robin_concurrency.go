package algorithms

import (
	"log"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/devChorok/load-balancer/pkg/types"
)

// RunRoundRobinConcurrencyTest performs a concurrency test on a round-robin load balancer with retry logic
func RunRoundRobinConcurrencyTest(t *testing.T, numRequests int, initialBPM int64, numNodes int, bpmLimit int64, rpmLimit int64, maxRetries int, retryDelay time.Duration) ([]int64, []int64) {
	nodes := make([]*types.Node, numNodes)
	for i := 0; i < numNodes; i++ {
		nodes[i] = &types.Node{
			Address:    "http://localhost:808" + strconv.Itoa(1+i),
			CurrentBPM: initialBPM,
			CurrentRPM: 0,
			BPM:        bpmLimit,
			RPM:        rpmLimit,
		}
	}

	rr := NewRoundRobin(nodes)

	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func(reqID int) {
			defer wg.Done()

			content := "RequestContent" + strconv.Itoa(reqID)
			contentLength := int64(len(content))

			var node *types.Node
			for retryCount := 0; retryCount < maxRetries; retryCount++ {
				node = rr.NextNodeWithRetry(contentLength,maxRetries)
				if node != nil {
					break
				}
				time.Sleep(retryDelay) // Wait before retrying
			}

			if node == nil {
				log.Printf("Error: Expected a node, but got nil after retries for request %d", reqID)
				return
			}

			log.Printf("Request %d routed to node: %s (CurrentRPM: %d, CurrentBPM: %d)", reqID, node.Address, node.CurrentRPM, node.CurrentBPM)

			rr.CompleteRequest(node, contentLength)

			log.Printf("Request %d completed at node: %s (CurrentRPM after completion: %d, CurrentBPM after completion: %d)", reqID, node.Address, node.CurrentRPM, node.CurrentBPM)
		}(i)
	}

	wg.Wait()

	// Collect final RPM and BPM values
	finalRPMs := make([]int64, len(nodes))
	finalBPMs := make([]int64, len(nodes))
	for i, node := range nodes {
		finalRPMs[i] = node.CurrentRPM
		finalBPMs[i] = node.CurrentBPM
	}

	return finalRPMs, finalBPMs
}
