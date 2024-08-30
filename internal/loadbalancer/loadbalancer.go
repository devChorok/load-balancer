package loadbalancer

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/devChorok/load-balancer/pkg/types"
)

type LoadBalancer struct {
    nodes       []*types.Node
    rateLimiter *RateLimiter
}

func NewLoadBalancer(nodes []*types.Node) *LoadBalancer {
    return &LoadBalancer{
        nodes:       nodes,
        rateLimiter: NewRateLimiter(nodes),
    }
}

func (lb *LoadBalancer) getNode() *types.Node {
    node := lb.nodes[0]
    lb.nodes = append(lb.nodes[1:], node)
    return node
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    node := lb.getNode()

    requestSize := int64(r.ContentLength)
    result := lb.rateLimiter.AllowRequest(node, requestSize)

    if !result.Allow {
        http.Error(w, result.Error, http.StatusTooManyRequests)
        return
    }

    // Forward the request to the selected node
    resp, err := http.Post(node.Address, r.Header.Get("Content-Type"), r.Body)
    if err != nil {
        log.Println(err)
        http.Error(w, "Failed to forward request", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
    w.Write(body)
}
