package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/devChorok/load-balancer/internal/loadbalancer"
	"github.com/devChorok/load-balancer/pkg/types"
)

var (
	mu           sync.Mutex
	runningNodes = map[string]*exec.Cmd{}
)

func StartServer() {
   // Ensure backend servers are running
   startBackendServer("8081")
   startBackendServer("8082")

   // Add a small delay to allow backend servers to fully start
   time.Sleep(2 * time.Second)
   // Nodes could be loaded from a config file, environment variables, or an API
   nodes := []*types.Node{
    {
        Address: "http://localhost:8081",
        BPM:     2048 * 1024, // Increase to 2 MB per minute
        RPM:     200,         // Increase to 200 requests per minute
    },
    {
        Address: "http://localhost:8082",
        BPM:     2048 * 1024, // Increase to 2 MB per minute
        RPM:     200,         // Increase to 200 requests per minute
    },
}


    lb := loadbalancer.NewLoadBalancer(nodes, loadbalancer.AlgorithmRoundRobin)

    // Health check implementation (example: simple ping)
    go startHealthChecks(lb)

    // Start monitoring and dynamic scaling
    go monitorLoadAndScale(lb)

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

func startBackendServer(port string) {
	address := "localhost:" + port

	// Check if the port is already in use
	if !isPortOpen(address) {
		log.Printf("Starting backend server on port %s...", port)

		// Command to run the generic backend server with the port as an argument
		cmd := exec.Command("go", "run", "backend/server.go", "--port", port)

		// Redirect stdout and stderr to log output
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Start the command
		err := cmd.Start()
		if err != nil {
			log.Fatalf("Failed to start backend server on port %s: %v", port, err)
		}

		// Store the running command in a map to keep track of it
		runningNodes[port] = cmd

		// Check if the server started successfully, checking every 10 seconds for up to 1 minute
		success := false
		for i := 0; i < 6; i++ { // 6 iterations, 10 seconds each
			time.Sleep(10 * time.Second)

			if isPortOpen(address) {
				success = true
				break
			}

			log.Printf("Attempt %d: Server on port %s not responding, checking again...", i+1, port)
		}

		if !success {
			// If the server did not start correctly within 1 minute, kill the process
			err = cmd.Process.Kill()
			if err != nil {
				log.Printf("Failed to kill backend server process on port %s: %v", port, err)
			} else {
				log.Printf("Killed the failed backend server process on port %s", port)
			}
			log.Fatalf("Backend server on port %s failed to start or is not listening on the expected port.", port)
		}

		log.Printf("Backend server on port %s started successfully.", port)
	} else {
		log.Printf("Backend server on port %s is already running.", port)
	}
}




// isPortOpen checks if a port is open on a given address.
func isPortOpen(address string) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// monitorLoadAndScale checks the load and dynamically scales the number of backend servers.
func monitorLoadAndScale(lb *loadbalancer.LoadBalancer) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
        // Get the current RPM for each node
        currentRPMs := lb.GetCurrentRPM()
    
        // Initialize a flag to determine if all nodes are above the threshold
        allAboveThreshold := true
    
        // Check if all nodes have an RPM above 20
        for _, rpm := range currentRPMs {
            if rpm <= 20 {
                allAboveThreshold = false
                break
            }
        }
    
        if allAboveThreshold {
            // Scale up if all nodes' RPM are above 20
            fmt.Println("All nodes RPM above 20, scaling up...")
			startBackendServer("8083")
			} else if anyBelowThreshold := checkIfAnyBelowThreshold(currentRPMs, 10); anyBelowThreshold && len(lb.GetNodes()) > 2 {
            // Scale down if any node's RPM is below 10 and there are more than 2 nodes
            fmt.Println("Some nodes RPM below 10, scaling down...")
            removeBackendServer(lb, "8083")
        }
    }
}

// Helper function to check if any RPM is below a given threshold
func checkIfAnyBelowThreshold(currentRPMs map[string]int64, threshold int64) bool {
	for _, rpm := range currentRPMs {
		if rpm < threshold {
			return true
		}
	}
	return false
}

// removeBackendServer stops a backend server and removes it from the load balancer.
func removeBackendServer(lb *loadbalancer.LoadBalancer, port string) {
	mu.Lock()
	defer mu.Unlock()

	cmd, exists := runningNodes[port]
	if !exists {
		return // Server not running
	}

	fmt.Printf("Stopping backend server on port %s...\n", port)
	if err := cmd.Process.Kill(); err != nil {
		log.Fatalf("Failed to stop backend server on port %s: %v", port, err)
	}
	delete(runningNodes, port)

	// Remove the node from the load balancer
	lb.RemoveNode("http://localhost:" + port)
}
