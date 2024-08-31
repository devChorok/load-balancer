package algorithms

import (
	"log"
	"sync"
	"testing"
    "strconv"

	"github.com/devChorok/load-balancer/pkg/types"
)

func RunRoundRobinConcurrencyTest(t *testing.T, numRequests int, initialBPM int64, numNodes int) []int64 {
    nodes := make([]*types.Node, numNodes)
	for i := 0; i < numNodes; i++ {
		nodes[i] = &types.Node{
			Address:    "http://localhost:808" + strconv.Itoa(1+i),
			CurrentBPM: initialBPM,
		}
	}


    rr := NewRoundRobin(nodes)

    var wg sync.WaitGroup
    wg.Add(numRequests)

    for i := 0; i < numRequests; i++ {
        go func(reqID int) {
            defer wg.Done()

            node := rr.NextNode()
            if node == nil {
                t.Errorf("Expected a node, but got nil")
                return
            }

            log.Printf("Request %d routed to node: %s (CurrentBPM: %d)", reqID, node.Address, node.CurrentBPM)

            rr.CompleteRequest(node)
            log.Printf("Request %d completed at node: %s (CurrentBPM after completion: %d)", reqID, node.Address, node.CurrentBPM)
        }(i)
    }

    wg.Wait()

    // Collect and return final CurrentBPM values for each node
    finalRPMs := make([]int64, len(nodes))
    for i, node := range nodes {
        finalRPMs[i] = node.CurrentBPM
    }
    return finalRPMs
}

