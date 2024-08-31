package algorithms_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	algorithms "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/weighted_round_robin"
	"github.com/devChorok/load-balancer/pkg/types"
)

func TestWeightedRoundRobinConcurrency(t *testing.T) {
	log.Println("Starting TestWeightedRoundRobinConcurrency")

	// Setup mock servers
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server1"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server2"))
	}))
	defer server2.Close()

	maxRequests := 0
	var mu sync.Mutex

	// Use context with a 10-second timeout for the test
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for numRequests := 100; ; numRequests += 100 {
		log.Printf("Testing with %d requests", numRequests)

		initialBPM := int64(0)
		numNodes := 2
		bpmLimit := int64(10000)
		rpmLimit := int64(100)
		maxRetries := 5
		retryDelay := time.Millisecond

		nodes := []*types.Node{
			{Address: server1.URL, CurrentBPM: initialBPM, BPM: bpmLimit, RPM: rpmLimit},
			{Address: server2.URL, CurrentBPM: initialBPM, BPM: bpmLimit, RPM: rpmLimit},
		}

		// Channel to capture errors
		errorChan := make(chan struct{}, 1)

		// Run the test in a goroutine to handle timeouts
		done := make(chan struct{})
		go func() {
			defer close(done)

			finalRPMs, finalBPMs := algorithms.RunWeightedRoundRobinConcurrencyTest(t, numRequests, initialBPM, numNodes, bpmLimit, rpmLimit, maxRetries, retryDelay)

			// Validate results
			for i := range nodes {
				if finalBPMs[i] > bpmLimit {
					mu.Lock()
					errorChan <- struct{}{}
					mu.Unlock()
					return
				}
			}

			// Log results for debug purposes
			log.Printf("Final RPMs: %v", finalRPMs)
			log.Printf("Final BPMs: %v", finalBPMs)
		}()

		// Wait for test completion or timeout
		select {
		case <-done:
			// Test completed successfully
			log.Printf("Test completed with %d requests", numRequests)
			maxRequests = numRequests
		case <-ctx.Done():
			// Timeout occurred
			log.Printf("Test timed out after 10sec: concurrent %v",maxRequests)
			return
		}

		// Check for errors and break if found
		select {
		case <-errorChan:
			log.Printf("Test failed at %d requests", numRequests)
			break
		default:
			log.Println("TestWeightedRoundRobinConcurrency completed.")
		}
	}

}
