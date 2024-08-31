package algorithms_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	algorithms "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/least_connections"
	"github.com/devChorok/load-balancer/pkg/types"
)

// TestLeastConnectionsMaxConcurrency performs a test with retry logic and tracks the number of requests before the first error
func TestLeastConnectionsMaxConcurrency(t *testing.T) {
	log.Println("Starting TestLeastConnectionsMaxConcurrency")

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
	var firstErrorRequestCount int
	var mu sync.Mutex

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
		errorChan := make(chan bool, 1)

		// Run the test in a goroutine to handle timeouts
		done := make(chan bool)
		go func() {
			defer close(done)
			finalRPMs, finalBPMs := algorithms.RunLeastConnectionsConcurrencyTest(t, numRequests, initialBPM, numNodes, bpmLimit, rpmLimit, maxRetries, retryDelay)

		// Validate results
			for i := range nodes {
				if finalBPMs[i] > bpmLimit {
					mu.Lock()
					firstErrorRequestCount = numRequests
					errorChan <- true
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
		case <-time.After(time.Minute):
			// Timeout occurred
			log.Printf("Test timed out after %v", time.Minute)
			return
		}

		// Check for errors and break if found
		if len(errorChan) > 0 {
			log.Printf("Test failed at %d requests", numRequests)
			break
		}
	}

	// Log the number of requests before the first error occurred
	if firstErrorRequestCount > 0 {
		log.Printf("Maximum number of requests handled before error: %d", firstErrorRequestCount)
	} else {
		log.Printf("No errors encountered. Maximum number of requests handled: %d", maxRequests)
	}
	log.Println("TestLeastConnectionsMaxConcurrency completed.")
}
