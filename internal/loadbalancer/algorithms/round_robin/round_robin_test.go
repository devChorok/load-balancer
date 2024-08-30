package algorithms_test

import (
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
