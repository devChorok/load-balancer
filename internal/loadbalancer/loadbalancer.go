package loadbalancer

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	lc "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/least_connections"
	ll "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/least_loaded"
	rr "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/round_robin"
	wrr "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/weighted_round_robin"
	"github.com/devChorok/load-balancer/pkg/types"
)

const (
	AlgorithmRoundRobin          = "round_robin"
	AlgorithmWeightedRoundRobin = "weighted_round_robin"
	AlgorithmLeastConnections   = "least_connections"
	AlgorithmLeastLoaded        = "least_loaded"
)

type LoadBalancer struct {
	algorithm   Algorithm
	rateLimiter *RateLimiter
    nodeCount   int

	mu          sync.Mutex
}

type Algorithm interface {
	NextNode() *types.Node
}

func NewLoadBalancer(nodes []*types.Node, algorithmType string) *LoadBalancer {
	var algorithm Algorithm
	rateLimiter := NewRateLimiter(nodes)

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

	return &LoadBalancer{
		algorithm:   algorithm,
		rateLimiter: rateLimiter,
        nodeCount:len(nodes),
	}
}
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle each request in a separate goroutine
	go lb.handleRequest(w, r)
}

func (lb *LoadBalancer) handleRequest(w http.ResponseWriter, r *http.Request) {
	maxRetries := lb.nodeCount // Use the nodeCount field
	retries := 0

	for retries < maxRetries {
		node := lb.NextNode()

		if node == nil {
			http.Error(w, "No available nodes", http.StatusServiceUnavailable)
			return
		}

		// Check if the node can handle the request
		result := lb.rateLimiter.AllowRequest(node, int64(r.ContentLength))
		if !result.Allow {
			retries++
			continue
		}

		// Forward the request to the selected node
		resp, err := http.Post(node.Address, r.Header.Get("Content-Type"), r.Body)
		if err != nil {
			log.Printf("Failed to forward request to %s: %v", node.Address, err)
			retries++
			continue
		}
		defer resp.Body.Close()

		// Copy the response from the selected node back to the client
		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		w.Write(body)

		// Successfully handled the request, exit the loop and the function
		return
	}

	// If all retries fail, return an error
	http.Error(w, "Failed to forward request after multiple attempts", http.StatusServiceUnavailable)
}


func (lb *LoadBalancer) NextNode() *types.Node {
	for {
		node := lb.algorithm.NextNode()
		if node != nil && lb.rateLimiter.AllowRequest(node, 0).Allow { // Check limits with zero bytes to select a node
			return node
		}
	}
}
