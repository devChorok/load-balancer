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
	mu          sync.RWMutex // Protects access to the nodes slice
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

	log.Printf("[INFO] Load balancer created with %d nodes using algorithm: %s", len(nodes), algorithmType)

	return &LoadBalancer{
		algorithm:   algorithm,
		rateLimiter: rateLimiter,
		nodeCount:   len(nodes),
		client: &http.Client{
			Timeout: 1 * time.Second, // Set a timeout for client requests
		},
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Received request: %s %s", r.Method, r.URL)
	lb.handleRequest(w, r)
}

func (lb *LoadBalancer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Buffer the request body to support retries
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	contentLength := int64(len(bodyBytes)) // Calculate content length from buffered body
	log.Printf("[INFO] Request body size: %d bytes", contentLength)

	maxRetries := lb.nodeCount
	retries := 0

	for retries < maxRetries {
		node := lb.NextNode(contentLength)

		if node == nil {
			log.Printf("[ERROR] No available nodes")
			http.Error(w, "No available nodes", http.StatusServiceUnavailable)
			return
		}

		// Check if the node can handle the request
		if !lb.rateLimiter.AllowRequest(node, contentLength).Allow {
			log.Printf("[INFO] Node %s rate limit exceeded, retrying... (%d/%d)", node.Address, retries+1, maxRetries)
			retries++
			continue
		}

		// Forward the request to the selected node
		log.Printf("[INFO] Forwarding request to node: %s", node.Address)
		req, err := http.NewRequest(r.Method, node.Address, bytes.NewReader(bodyBytes))
		if err != nil {
			log.Printf("[ERROR] Failed to create request to node %s: %v", node.Address, err)
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}

		req.Header = r.Header

		resp, err := lb.client.Do(req)
		if err != nil {
			log.Printf("[ERROR] Failed to forward request to %s: %v (retrying... %d/%d)", node.Address, err, retries+1, maxRetries)
			retries++
			continue
		}
		defer resp.Body.Close()

		log.Printf("[INFO] Received response from node %s: %d", node.Address, resp.StatusCode)

		// Copy the response from the selected node back to the client
		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		w.Write(body)

		// Successfully handled the request, exit the loop and the function
		log.Printf("[INFO] Successfully handled request for %s via node %s", r.URL, node.Address)
		return
	}

	// If all retries fail, return an error
	log.Printf("[ERROR] Failed to forward request after multiple attempts")
	http.Error(w, "Failed to forward request after multiple attempts", http.StatusServiceUnavailable)
}

func (lb *LoadBalancer) NextNode(contentLength int64) *types.Node {
    lb.mu.Lock()
    defer lb.mu.Unlock()

    for _, node := range lb.nodes {
        log.Printf("[INFO] Node %s: Healthy=%t, BPM=%d, RPM=%d", node.Address, node.Healthy, node.BPM, node.RPM)
    }

    for {
        node := lb.algorithm.NextNode(contentLength)
        if node != nil && lb.rateLimiter.AllowRequest(node, contentLength).Allow {
            log.Printf("[INFO] Selected node: %s", node.Address)
            return node
        }
        log.Printf("[WARN] No suitable node found, retrying...")
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
			log.Printf("[WARN] Marked node %s as unhealthy", address)
			break
		}
	}
}

// MarkNodeHealthy marks a node as healthy.
func (lb *LoadBalancer) MarkNodeHealthy(address string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for _, n := range lb.nodes {
		if n.Address == address {
			n.Healthy = true
			log.Printf("[INFO] Marked node %s as healthy", address)
			break
		}
	}
}

// AddNode adds a new node to the load balancer.
func (lb *LoadBalancer) AddNode(node *types.Node) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.nodes = append(lb.nodes, node)
	lb.nodeCount++
	log.Printf("[INFO] Added new node: %s", node.Address)
}

// RemoveNode removes a node from the load balancer by its address.
func (lb *LoadBalancer) RemoveNode(address string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i, n := range lb.nodes {
		if n.Address == address {
			// Remove the node from the slice
			lb.nodes = append(lb.nodes[:i], lb.nodes[i+1:]...)
			lb.nodeCount--
			log.Printf("[INFO] Removed node: %s", address)
			break
		}
	}
}
func (lb *LoadBalancer) GetCurrentRPM() map[string]int64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	currentRPM := make(map[string]int64)

	for _, node := range lb.nodes {
		if node.Healthy {
			currentRPM[node.Address] = node.RPM // Assuming RPM is a field in your Node struct
			log.Printf("[INFO] Node %s current RPM: %d", node.Address, node.RPM)
		} else {
			log.Printf("[WARN] Node %s is unhealthy, skipping RPM calculation", node.Address)
		}
	}

	return currentRPM
}
