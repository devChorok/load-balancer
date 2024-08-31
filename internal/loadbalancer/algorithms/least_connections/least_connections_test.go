package algorithms_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
    "log"

	algorithms "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/least_connections"
	"github.com/devChorok/load-balancer/pkg/types"
)

func TestLeastConnections(t *testing.T) {
    nodes := []*types.Node{
        {Address: "http://localhost:8081", CurrentRPM: 2},
        {Address: "http://localhost:8082", CurrentRPM: 1},
    }

    lc := algorithms.NewLeastConnections(nodes)

    node := lc.NextNode()

    if node.Address != "http://localhost:8082" {
        t.Errorf("Expected http://localhost:8082 (least connections), got %s", node.Address)
    }
}

func TestLeastConnectionsConcurrency(t *testing.T) {
    log.Println("Starting TestLeastConnectionsConcurrency")

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
        {Address: server1.URL, CurrentRPM: 2},
        {Address: server2.URL, CurrentRPM: 1},
    }

    lc := algorithms.NewLeastConnections(nodes)

    const numRequests = 100
    var wg sync.WaitGroup
    wg.Add(numRequests)

    for i := 0; i < numRequests; i++ {
        go func(reqID int) {
            defer wg.Done()

            node := lc.NextNode()
            if node == nil {
                t.Errorf("Expected a node, but got nil")
                return
            }

            log.Printf("Request %d routed to node: %s (CurrentRPM: %d)", reqID, node.Address, node.CurrentRPM)

            // Simulate request completion
            lc.CompleteRequest(node)

            log.Printf("Request %d completed at node: %s (CurrentRPM after completion: %d)", reqID, node.Address, node.CurrentRPM)
        }(i)
    }

    wg.Wait()

    // Allow a small margin of error due to concurrency timing issues
    if nodes[0].CurrentRPM <= nodes[1].CurrentRPM {
        t.Errorf("Expected node 1 to have more RPM than node 2, but got %d and %d respectively", nodes[0].CurrentRPM, nodes[1].CurrentRPM)
    }

    log.Printf("Final RPMs - Node 1: %d, Node 2: %d", nodes[0].CurrentRPM, nodes[1].CurrentRPM)
}