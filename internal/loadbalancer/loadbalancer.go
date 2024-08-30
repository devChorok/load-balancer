package loadbalancer

import (
	"io/ioutil"
	"log"
	"net/http"

	lc "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/least_connections"
	ll "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/least_loaded"
	rr "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/round_robin"
	wrr "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/weighed_round_robin"

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
        algorithm = rr.NewRoundRobin(nodes)
    case AlgorithmWeightedRoundRobin:
        algorithm = wrr.NewWeightedRoundRobin(nodes)
    case AlgorithmLeastConnections:
        algorithm = lc.NewLeastConnections(nodes)
    case AlgorithmLeastLoaded:
        algorithm = ll.NewLeastLoaded(nodes)
    default:
        algorithm = rr.NewRoundRobin(nodes) // Default to round robin
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
