package algorithms_test

import (
	"testing"

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
