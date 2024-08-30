package loadbalancer

import (
	"sync"
	"time"

	"github.com/devChorok/load-balancer/pkg/types"
)

type RateLimiter struct {
    mu    sync.Mutex
    nodes map[string]*types.Node
}

func NewRateLimiter(nodes []*types.Node) *RateLimiter {
    nodeMap := make(map[string]*types.Node)
    for _, node := range nodes {
        nodeMap[node.Address] = node
    }
    return &RateLimiter{
        nodes: nodeMap,
    }
}

func (rl *RateLimiter) AllowRequest(node *types.Node, requestSize int64) types.RateLimitResult {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    // Reset counters every minute
    if time.Now().Sub(time.Unix(node.CurrentBPM, 0)) > time.Minute {
        node.CurrentBPM = 0
        node.CurrentRPM = 0
    }

    if node.CurrentBPM+requestSize > node.BPM {
        return types.RateLimitResult{Allow: false, Error: "BPM limit exceeded"}
    }
    if node.CurrentRPM+1 > node.RPM {
        return types.RateLimitResult{Allow: false, Error: "RPM limit exceeded"}
    }

    // Update node stats
    node.CurrentBPM += requestSize
    node.CurrentRPM += 1

    return types.RateLimitResult{Allow: true}
}
