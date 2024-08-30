package loadbalancer

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/devChorok/load-balancer/internal/loadbalancer/algorithms"
	"github.com/devChorok/load-balancer/pkg/types"
)
const (
    AlgorithmRoundRobin        = "round_robin"
    AlgorithmWeightedRoundRobin = "weighted_round_robin"
    AlgorithmLeastConnections   = "least_connections"
    AlgorithmLeastLoaded        = "least_loaded"
)
type LoadBalancer struct {
    algorithm Algorithm
}

type Algorithm interface {
    NextNode() *types.Node
}

func NewLoadBalancer(nodes []*types.Node, algorithmType string) *LoadBalancer {
    var algorithm Algorithm

    switch algorithmType {
    case AlgorithmRoundRobin:
        algorithm = algorithms.NewRoundRobin(nodes)
    case AlgorithmWeightedRoundRobin:
        algorithm = algorithms.NewWeightedRoundRobin(nodes)
    case AlgorithmLeastConnections:
        algorithm = algorithms.NewLeastConnections(nodes)
    case AlgorithmLeastLoaded:
        algorithm = algorithms.NewLeastLoaded(nodes)
    default:
        algorithm = algorithms.NewRoundRobin(nodes) // Default to round robin
    }

    return &LoadBalancer{algorithm: algorithm}
}
// ServeHTTP handles HTTP requests and forwards them to the appropriate backend node
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Select the next node using the chosen algorithm
    node := lb.NextNode()

    // Forward the request to the selected node
    resp, err := http.Post(node.Address, r.Header.Get("Content-Type"), r.Body)
    if err != nil {
        log.Printf("Failed to forward request to %s: %v", node.Address, err)
        http.Error(w, "Failed to forward request", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    // Copy the response from the selected node back to the client
    for k, v := range resp.Header {
        w.Header()[k] = v
    }
    w.WriteHeader(resp.StatusCode)
    body, _ := ioutil.ReadAll(resp.Body)
    w.Write(body)
}

func (lb *LoadBalancer) NextNode() *types.Node {
    return lb.algorithm.NextNode()
}
