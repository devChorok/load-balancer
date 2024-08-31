package loadbalancer_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	lc "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/least_connections"
	ll "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/least_loaded"
	rr "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/round_robin"
	wrr "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/weighted_round_robin"
	"github.com/devChorok/load-balancer/pkg/types"
)
type Algorithm interface {
	NextNode(contentLength int64) *types.Node
}
type AlgorithmFactory func([]*types.Node) Algorithm

type TestScenario struct {
	Name          string
	AlgorithmType string
	NumRequests   int
	NumNodes      int
	BPMLimit      int64
	RPMLimit      int64
	RequestSize   int64
	Weights       []int
}

type ScenarioResult struct {
    Algorithm        string
    MaxConcurrentRequests int
}

func TestLoadBalancersInVariousScenarios(t *testing.T) {
    // Define the algorithms to test
    algorithmsToTest := map[string]AlgorithmFactory{
        "Round Robin":           func(nodes []*types.Node) Algorithm { return rr.NewRoundRobin(nodes) },
        "Least Connections":     func(nodes []*types.Node) Algorithm { return lc.NewLeastConnections(nodes) },
        "Least Loaded":          func(nodes []*types.Node) Algorithm { return ll.NewLeastLoaded(nodes) },
        "Weighted Round Robin":  func(nodes []*types.Node) Algorithm { return wrr.NewWeightedRoundRobin(nodes) },
    }

    // Define test servers
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

    server3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Server3"))
    }))
    defer server3.Close()

    // Define test scenarios
    scenarios := []TestScenario{
        {"Scenario 1: Low Load, Even Distribution", "Round Robin", 100, 3, 10000, 100, 1024, nil},
        {"Scenario 2: High Load, Even Distribution", "Round Robin", 1000, 3, 10000, 100, 1024, nil},
        {"Scenario 3: Low Load, Unbalanced Nodes", "Least Loaded", 100, 3, 5000, 100, 2048, nil},
        {"Scenario 4: High Load, Unbalanced Nodes", "Least Loaded", 1000, 3, 5000, 100, 2048, nil},
        {"Scenario 5: Low Load, Varying Request Sizes", "Least Connections", 100, 3, 10000, 100, 512, nil},
        {"Scenario 6: High Load, Varying Request Sizes", "Least Connections", 1000, 3, 10000, 100, 512, nil},
        {"Scenario 7: High Load, Weighted Distribution", "Weighted Round Robin", 1000, 3, 10000, 100, 1024, []int{1, 2, 3}},
        {"Scenario 8: Low Load, Weighted Distribution", "Weighted Round Robin", 100, 3, 10000, 100, 1024, []int{1, 2, 3}},
    }

    // Map to store the maximum concurrent requests each algorithm could handle before failing, per scenario
    scenarioResults := make(map[string][]ScenarioResult)

    // Run the test scenarios
    for _, scenario := range scenarios {
        t.Run(scenario.Name, func(t *testing.T) {
            nodes := setupNodes(scenario, server1, server2, server3)
            algorithm := createAlgorithm(scenario, nodes, algorithmsToTest)
            if algorithm == nil {
                t.Fatalf("Algorithm %s could not be initialized", scenario.AlgorithmType)
            }

            maxRequests := testAlgorithmGradualLoad(t, algorithm, scenario)
            scenarioResults[scenario.Name] = append(scenarioResults[scenario.Name], ScenarioResult{Algorithm: scenario.AlgorithmType, MaxConcurrentRequests: maxRequests})
        })
    }

    // Summarize the results
    log.Println("===== Summary of Test Results =====")
	for scenarioName, results := range scenarioResults {
		log.Printf("Scenario: %s", scenarioName)
		log.Println("------------------------------------------------")
		
		// Find the best algorithm
		bestAlgorithm := ""
		bestMaxRequests := 0
		for _, result := range results {
			if result.MaxConcurrentRequests > bestMaxRequests {
				bestMaxRequests = result.MaxConcurrentRequests
				bestAlgorithm = result.Algorithm
			}
		}

		// Log the best algorithm or state that no algorithm passed
		if bestMaxRequests > 0 {
			log.Printf("  Best Algorithm: %-20s | Handled %d Concurrent Requests", bestAlgorithm, bestMaxRequests)
		} else {
			log.Println("  No algorithm passed the test.")
		}
		
		log.Println("") // Extra line between scenarios for clarity
	}
log.Println("===== End of Summary =====")

}



func setupNodes(scenario TestScenario, server1, server2, server3 *httptest.Server) []*types.Node {
	// Use weights if provided, otherwise default to equal weight distribution
	weights := scenario.Weights
	if weights == nil {
		weights = []int{1, 1, 1}
	}

	nodes := []*types.Node{
		{Address: server1.URL, BPM: scenario.BPMLimit, RPM: scenario.RPMLimit, Weight: weights[0]},
		{Address: server2.URL, BPM: scenario.BPMLimit, RPM: scenario.RPMLimit, Weight: weights[1]},
		{Address: server3.URL, BPM: scenario.BPMLimit, RPM: scenario.RPMLimit, Weight: weights[2]},
	}

	return nodes
}

func createAlgorithm(scenario TestScenario, nodes []*types.Node, algorithmsToTest map[string]AlgorithmFactory) Algorithm {
	return algorithmsToTest[scenario.AlgorithmType](nodes)
}
func testAlgorithmGradualLoad(t *testing.T, algorithm Algorithm, scenario TestScenario) int {
    var mu sync.Mutex
    var maxConcurrentRequests int
    errorChan := make(chan bool, 1) // Buffer size 1 to prevent deadlocks
    var loggedNilNode bool          // Flag to track if the nil node message has been logged

    for numRequests := scenario.NumRequests; numRequests <= 10000; numRequests += 100 {
        log.Printf("Testing with %d concurrent requests for algorithm %s in scenario %s", numRequests, scenario.AlgorithmType, scenario.Name)

        wg := sync.WaitGroup{}
        wg.Add(numRequests)

        for i := 0; i < numRequests; i++ {
            go func() {
                defer wg.Done()

                node := algorithm.NextNode(scenario.RequestSize)
                if node == nil {
                    mu.Lock()
                    if !loggedNilNode {
                        log.Printf("Nil node returned by algorithm %s in scenario %s at %d concurrent requests", scenario.AlgorithmType, scenario.Name, numRequests)
                        loggedNilNode = true
                    }
                    if maxConcurrentRequests == 0 {
                        maxConcurrentRequests = numRequests - 100
                    }
                    mu.Unlock()
                    select {
                    case errorChan <- true:
                    default:
                        return
                    }
                    return // Ensure the goroutine exits early
                }

                _, err := http.Post(node.Address, "application/json", nil)
                if err != nil {
                    mu.Lock()
                    if maxConcurrentRequests == 0 {
                        maxConcurrentRequests = numRequests - 100
                    }
                    mu.Unlock()
                    select {
                    case errorChan <- true:
                    default:
                        return
                    }
                    return // Ensure the goroutine exits early
                }
            }()
        }

        wg.Wait()

        select {
        case <-errorChan:
            return maxConcurrentRequests
        default:
            log.Printf("Successfully handled %d concurrent requests for algorithm %s in scenario %s", numRequests, scenario.AlgorithmType, scenario.Name)
            maxConcurrentRequests = numRequests
        }
    }

    return maxConcurrentRequests
}






