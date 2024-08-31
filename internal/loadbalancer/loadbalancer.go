package loadbalancer

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	lc "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/least_connections"
	ll "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/least_loaded"
	rr "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/round_robin"
	wrr "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/weighted_round_robin"
	"github.com/devChorok/load-balancer/pkg/types"
)

const (
	AlgorithmRoundRobin         = "round_robin"
	AlgorithmWeightedRoundRobin = "weighted_round_robin"
	AlgorithmLeastConnections   = "least_connections"
	AlgorithmLeastLoaded        = "least_loaded"
)

type LoadBalancer struct {
	nodes       []*types.Node  // Slice to hold the backend nodes
	algorithm   Algorithm
	rateLimiter *RateLimiter
	nodeCount   int
	client      *http.Client // Reusable HTTP client for efficiency
	mu        sync.RWMutex // Protects access to the nodes slice
}

type Algorithm interface {
	NextNode(contentLength int64) *types.Node
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
		nodeCount:   len(nodes),
		client: &http.Client{
			Timeout: 5 * time.Second, // Set a timeout for client requests
		},
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle each request
	lb.handleRequest(w, r)
}

func (lb *LoadBalancer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Buffer the request body to support retries
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	maxRetries := lb.nodeCount
	retries := 0
	contentLength := int64(len(bodyBytes)) // Calculate content length from buffered body

	for retries < maxRetries {
		node := lb.NextNode(contentLength)

		if node == nil {
			http.Error(w, "No available nodes", http.StatusServiceUnavailable)
			return
		}

		// Check if the node can handle the request
		if !lb.rateLimiter.AllowRequest(node, contentLength).Allow {
			retries++
			continue
		}

		// Forward the request to the selected node
		req, err := http.NewRequest(r.Method, node.Address, bytes.NewReader(bodyBytes))
		if err != nil {
			log.Printf("Failed to create request: %v", err)
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}

		req.Header = r.Header

		resp, err := lb.client.Do(req)
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

func (lb *LoadBalancer) NextNode(contentLength int64) *types.Node {
	// Lock the operation to ensure thread safety
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for {
		node := lb.algorithm.NextNode(contentLength)
		if node != nil && lb.rateLimiter.AllowRequest(node, contentLength).Allow {
			return node
		}
	}
}

// GetNodes returns a copy of the current list of nodes.
func (lb *LoadBalancer) GetNodes() []*types.Node {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	nodesCopy := make([]*types.Node, len(lb.nodes))
	copy(nodesCopy, lb.nodes)

	return nodesCopy
}

// MarkNodeUnhealthy marks a node as unhealthy.
func (lb *LoadBalancer) MarkNodeUnhealthy(address string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, n := range lb.nodes {
		if n.Address == address {
			n.Healthy = false
			break
		}
	}
}

// MarkNodeHealthy marks a node as healthy (optional, if you need this functionality).
func (lb *LoadBalancer) MarkNodeHealthy(address string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, n := range lb.nodes {
		if n.Address == address {
			n.Healthy = true
			break
		}
	}
}