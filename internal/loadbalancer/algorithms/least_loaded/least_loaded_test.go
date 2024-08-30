package algorithms_test

import (
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
