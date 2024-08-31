package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	loadBalancerURL = "http://localhost:8080" // Load balancer URL
	numRequests     = 1000                    // Number of requests for stress testing
	numWorkers      = 100                     // Number of concurrent workers
	requestBody     = "Sample request data"   // Replace with actual request data if needed
)

func main() {
	var wg sync.WaitGroup
	requests := make(chan int, numRequests)

	// Create workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(&wg, requests)
	}

	start := time.Now()

	// Send requests
	for i := 0; i < numRequests; i++ {
		requests <- i
	}

	close(requests)
	wg.Wait()

	duration := time.Since(start)
	fmt.Printf("Completed %d requests in %s\n", numRequests, duration)
}

func worker(wg *sync.WaitGroup, requests chan int) {
	defer wg.Done()

	for req := range requests {
		resp, err := sendRequest(req)
		if err != nil {
			log.Printf("Request %d failed: %v\n", req, err)
			continue
		}
		log.Printf("Request %d completed with status %s\n", req, resp.Status)
		resp.Body.Close()
	}
}

func sendRequest(reqID int) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", loadBalancerURL, strings.NewReader(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Request-ID", fmt.Sprintf("%d", reqID))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Optionally read the response body if needed
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
