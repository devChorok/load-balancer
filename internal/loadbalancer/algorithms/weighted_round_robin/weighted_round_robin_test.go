package algorithms_test

import (
    "log"
    "net/http"
    "net/http/httptest"
    "sync"
    "testing"

	algorithms "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/weighted_round_robin"
	"github.com/devChorok/load-balancer/pkg/types"
)

func TestWeightedRoundRobin(t *testing.T) {
    nodes := []*types.Node{
        {Address: "http://localhost:8081", Weight: 3},
        {Address: "http://localhost:8082", Weight: 1},
    }
    wrr := algorithms.NewWeightedRoundRobin(nodes)

    nodeCount := map[string]int{
        "http://localhost:8081": 0,
        "http://localhost:8082": 0,
    }

    for i := 0; i < 1000; i++ {
        node := wrr.NextNode()
        nodeCount[node.Address]++
    }

    if nodeCount["http://localhost:8081"] <= nodeCount["http://localhost:8082"] {
        t.Errorf("Weighted round robin is not working as expected: %v", nodeCount)
    }
}

func TestWeightedRoundRobinConcurrency(t *testing.T) {
    log.Println("Starting TestWeightedRoundRobinConcurrency")

    server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Server1"))
    }))
    defer server1.Close()

    server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Server2"))
    }))
    defer server2.Close()

    nodes := []*types.Node{
        {Address: server1.URL, Weight: 3}, // Higher weight
        {Address: server2.URL, Weight: 1}, // Lower weight
    }

    wrr := algorithms.NewWeightedRoundRobin(nodes)

    const numRequests = 100
    var wg sync.WaitGroup
    wg.Add(numRequests)

    for i := 0; i < numRequests; i++ {
        go func(reqID int) {
            defer wg.Done()

            node := wrr.NextNode()
            if node == nil {
                t.Errorf("Expected a node, but got nil")
                return
            }

            log.Printf("Request %d routed to node: %s (Weight: %d, CurrentRPM: %d)", reqID, node.Address, node.Weight, node.CurrentRPM)

            // Simulate request completion
            wrr.CompleteRequest(node)

            log.Printf("Request %d completed at node: %s (CurrentRPM after completion: %d)", reqID, node.Address, node.CurrentRPM)
        }(i)
    }

    wg.Wait()

    // Final log of CurrentRPMs for each node
    log.Printf("Final RPMs - Node 1: %d, Node 2: %d", nodes[0].CurrentRPM, nodes[1].CurrentRPM)

    log.Println("TestWeightedRoundRobinConcurrency completed.")
}
