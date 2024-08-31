package algorithms

import (
	"log"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/devChorok/load-balancer/pkg/types"
)

// RunLeastLoadedConcurrencyTest performs a concurrency test on the LeastLoaded algorithm
func RunLeastLoadedConcurrencyTest(t *testing.T, numRequests int, initialBPM int64, numNodes int, bpmLimit int64, rpmLimit int64, maxRetries int, retryDelay time.Duration) ([]int64, []int64) {
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

	ll := NewLeastLoaded(nodes)

	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func(reqID int) {
			defer wg.Done()

			content := "RequestContent" + strconv.Itoa(reqID)
			contentLength := int64(len(content))

			var node *types.Node
			for retryCount := 0; retryCount < maxRetries; retryCount++ {
				node = ll.NextNodeWithRetry(contentLength,maxRetries)
				if node != nil {
					break
				}
				time.Sleep(retryDelay)
			}

			if node == nil {
				log.Printf("Error: Expected a node, but got nil after retries for request %d", reqID)
				return
			}

			// Consider reducing or batching logs if too many logs cause contention
			log.Printf("Request %d routed to node: %s (CurrentBPM: %d)", reqID, node.Address, node.CurrentBPM)

			ll.CompleteRequest(node, contentLength)

			log.Printf("Request %d completed at node: %s (CurrentBPM after completion: %d)", reqID, node.Address, node.CurrentBPM)
		}(i)
	}

	wg.Wait()

	finalRPMs := make([]int64, len(nodes))
	finalBPMs := make([]int64, len(nodes))
	for i, node := range nodes {
		finalRPMs[i] = node.CurrentRPM
		finalBPMs[i] = node.CurrentBPM
	}
	return finalRPMs, finalBPMs
}
