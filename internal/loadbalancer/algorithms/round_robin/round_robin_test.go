package algorithms_test
import (
    "log"
    "net/http"
    "net/http/httptest"
    "sync"
    "testing"

    "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/round_robin"
    "github.com/devChorok/load-balancer/pkg/types"
)

func TestRoundRobin(t *testing.T) {
    nodes := []*types.Node{
        {Address: "http://localhost:8081"},
        {Address: "http://localhost:8082"},
    }

    rr := algorithms.NewRoundRobin(nodes)

    node1 := rr.NextNode()
    node2 := rr.NextNode()
    node3 := rr.NextNode()

    if node1.Address != "http://localhost:8081" {
        t.Errorf("Expected http://localhost:8081, got %s", node1.Address)
    }

    if node2.Address != "http://localhost:8082" {
        t.Errorf("Expected http://localhost:8082, got %s", node2.Address)
    }

    if node3.Address != "http://localhost:8081" {
        t.Errorf("Expected http://localhost:8081, got %s", node3.Address)
    }
}

func TestRoundRobinConcurrency(t *testing.T) {
    log.Println("Starting TestRoundRobinConcurrency")

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
        {Address: server1.URL},
        {Address: server2.URL},
    }

    rr := algorithms.NewRoundRobin(nodes)

    const numRequests = 100
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

            log.Printf("Request %d routed to node: %s (CurrentRPM: %d)", reqID, node.Address, node.CurrentRPM)

            // Simulate request completion
            rr.CompleteRequest(node)

            log.Printf("Request %d completed at node: %s (CurrentRPM after completion: %d)", reqID, node.Address, node.CurrentRPM)
        }(i)
    }

    wg.Wait()

    // Final log of CurrentRPMs for each node
    log.Printf("Final RPMs - Node 1: %d, Node 2: %d", nodes[0].CurrentRPM, nodes[1].CurrentRPM)

    log.Println("TestRoundRobinConcurrency completed.")
}