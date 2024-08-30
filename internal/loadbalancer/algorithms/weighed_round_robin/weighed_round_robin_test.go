package algorithms_test

import (
	"testing"

	algorithms "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/weighed_round_robin"
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
