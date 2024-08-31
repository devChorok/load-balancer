package algorithms

import (
	"log"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/devChorok/load-balancer/pkg/types"
)

func RunWeightedRoundRobinConcurrencyTest(t *testing.T, numRequests int, initialBPM int64, numNodes int, bpmLimit int64, rpmLimit int64, maxRetries int, retryDelay time.Duration) ([]int64, []int64) {
	nodes := make([]*types.Node, numNodes)
	for i := 0; i < numNodes; i++ {
		weight := int64(1)
		if i%2 == 0 {
			weight = 3
		}
		nodes[i] = &types.Node{
			Address:    "http://localhost:808" + strconv.Itoa(1+i),
			CurrentBPM: initialBPM,
			Weight:     int(weight),
			CurrentRPM: 0,
			BPM:        bpmLimit,
			RPM:        rpmLimit,
		}
	}

	wrr := NewWeightedRoundRobin(nodes)

	var wg sync.WaitGroup
	wg.Add(numRequests)

	// Map to track logged messages
	loggedMessages := make(map[string]struct{})
	var mu sync.Mutex

	// Function to log messages uniquely
	logUnique := func(message string) {
		mu.Lock()
		defer mu.Unlock()
		if _, exists := loggedMessages[message]; !exists {
			log.Println(message)
			loggedMessages[message] = struct{}{}
		}
	}

	for i := 0; i < numRequests; i++ {
		go func(reqID int) {
			defer wg.Done()

			content := "RequestContent" + strconv.Itoa(reqID)
			contentLength := int64(len(content))

			var node *types.Node
			for retryCount := 0; retryCount < maxRetries; retryCount++ {
				node = wrr.NextNodeWithRetry(contentLength, maxRetries)
				if node != nil {
					break
				}
				time.Sleep(retryDelay)
			}
			if node == nil {
				logUnique("Error: Expected a node, but got nil after retries for request " + strconv.Itoa(reqID))
				return
			}

			logUnique("Request " + strconv.Itoa(reqID) + " routed to node: " + node.Address + " (CurrentBPM: " + strconv.FormatInt(node.CurrentBPM, 10) + ")")

			wrr.CompleteRequest(node, contentLength)
			logUnique("Request " + strconv.Itoa(reqID) + " completed at node: " + node.Address + " (CurrentBPM after completion: " + strconv.FormatInt(node.CurrentBPM, 10) + ")")
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
