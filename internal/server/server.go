package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/devChorok/load-balancer/internal/loadbalancer"
	"github.com/devChorok/load-balancer/pkg/types"
)

func StartServer() {
    // Nodes could be loaded from a config file, environment variables, or an API
    nodes := []*types.Node{
        {Address: "http://localhost:8081", BPM: 1000, RPM: 10, Weight: 3},
        {Address: "http://localhost:8082", BPM: 1500, RPM: 15, Weight: 1},
    }

    lb := loadbalancer.NewLoadBalancer(nodes, loadbalancer.AlgorithmWeightedRoundRobin)

    // Health check implementation (example: simple ping)
    go startHealthChecks(lb)

    srv := &http.Server{
        Addr:    ":8080",
        Handler: lb,
    }

    // Signal channel to handle shutdown
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    // Run server in a goroutine
    go func() {
        fmt.Println("Load Balancer running on port 8080")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Could not listen on :8080: %v\n", err)
        }
    }()

    <-stop // Wait for interrupt signal

    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    log.Println("Shutting down gracefully, press Ctrl+C again to force")
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exiting")
}

func startHealthChecks(lb *loadbalancer.LoadBalancer) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        var wg sync.WaitGroup

        for _, node := range lb.GetNodes() {
            wg.Add(1)
        
            go func(n *types.Node) {
                defer wg.Done()
        
                if err := healthCheck(n); err != nil {
                    log.Printf("Node %s is unhealthy: %v", n.Address, err)
                    lb.MarkNodeUnhealthy(n.Address) // Pass the address, assuming MarkNodeUnhealthy takes an address
                } else {
                    lb.MarkNodeHealthy(n.Address) // Pass the address, assuming MarkNodeHealthy takes an address
                }
            }(node) // Pass the current value of node to the goroutine
        }
        

        wg.Wait()
    }
}

func healthCheck(node *types.Node) error {
    // Example: Simple HTTP GET request to check health
    resp, err := http.Get(node.Address + "/health")
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("node %s returned non-200 status: %d", node.Address, resp.StatusCode)
    }

    return nil
}
