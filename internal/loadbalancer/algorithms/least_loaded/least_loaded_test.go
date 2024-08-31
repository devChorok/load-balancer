package algorithms_test

import (
    "log"
    "net/http"
    "net/http/httptest"
    "sync"
    "testing"

    "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/least_loaded"
    "github.com/devChorok/load-balancer/pkg/types"
)

func TestLeastLoaded(t *testing.T) {
    nodes := []*types.Node{
        {Address: "http://localhost:8081", CurrentBPM: 500},
        {Address: "http://localhost:8082", CurrentBPM: 200},
    }

    ll := algorithms.NewLeastLoaded(nodes)

    node := ll.NextNode()

    if node.Address != "http://localhost:8082" {
        t.Errorf("Expected http://localhost:8082 (least loaded), got %s", node.Address)
    }
}
func TestLeastLoadedConcurrency(t *testing.T) {
    log.Println("Starting TestLeastLoadedConcurrency")

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
        {Address: server1.URL, CurrentBPM: 500},
        {Address: server2.URL, CurrentBPM: 200},
    }

    ll := algorithms.NewLeastLoaded(nodes)

    const numRequests = 100
    var wg sync.WaitGroup
    wg.Add(numRequests)

    for i := 0; i < numRequests; i++ {
        go func(reqID int) {
            defer wg.Done()

            node := ll.NextNode()
            if node == nil {
                t.Errorf("Expected a node, but got nil")
                return
            }

            log.Printf("Request %d routed to node: %s (CurrentBPM: %d)", reqID, node.Address, node.CurrentBPM)

            // Simulate request completion
            ll.CompleteRequest(node)

            log.Printf("Request %d completed at node: %s (CurrentBPM after completion: %d)", reqID, node.Address, node.CurrentBPM)
        }(i)
    }

    wg.Wait()

    // Final log of CurrentBPMs for each node
    log.Printf("Final BPMs - Node 1: %d, Node 2: %d", nodes[0].CurrentBPM, nodes[1].CurrentBPM)

    log.Println("TestLeastLoadedConcurrency completed.")
}