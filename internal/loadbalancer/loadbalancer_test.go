package loadbalancer

import (
	"log"
	"testing"

	lc "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/least_connections"
	lltest "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/least_loaded"
	rrtest "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/round_robin"
	wrrtest "github.com/devChorok/load-balancer/internal/loadbalancer/algorithms/weighted_round_robin"
)
func TestCompareAlgorithms(t *testing.T) {
    log.Println("Starting TestCompareAlgorithms")

    // Define the number of requests and initial BPM
    const numRequests = 100000
    const initialBPM = 2
	const numNodes =3

    // Run each algorithm's concurrency test
    log.Println("Running RoundRobin Concurrency Test")
    rrRPMs := rrtest.RunRoundRobinConcurrencyTest(t, numRequests, initialBPM, numNodes)
    log.Printf("Final RPMs for RoundRobin: %v", rrRPMs)

    log.Println("Running LeastLoaded Concurrency Test")
    llRPMs := lltest.RunLeastLoadedConcurrencyTest(t, numRequests, initialBPM,numNodes)
    log.Printf("Final RPMs for LeastLoaded: %v", llRPMs)

    log.Println("Running LeastConnections Concurrency Test")
    lcRPMs := lc.RunLeastConnectionsConcurrencyTest(t, numRequests, initialBPM,numNodes)
    log.Printf("Final RPMs for LeastConnections: %v", lcRPMs)

    log.Println("Running WeightedRoundRobin Concurrency Test")
    wrrRPMs := wrrtest.RunWeightedRoundRobinConcurrencyTest(t, numRequests, initialBPM,numNodes)
    log.Printf("Final RPMs for WeightedRoundRobin: %v", wrrRPMs)

    // Compare effectiveness
    compareResults(rrRPMs, llRPMs, lcRPMs, wrrRPMs)
}

func compareResults(rrRPMs, llRPMs, lcRPMs, wrrRPMs []int64) {
    log.Println("Comparing results across algorithms...")
    log.Printf("RoundRobin Final RPMs: %v", rrRPMs)
    log.Printf("LeastLoaded Final RPMs: %v", llRPMs)
    log.Printf("LeastConnections Final RPMs: %v", lcRPMs)
    log.Printf("WeightedRoundRobin Final RPMs: %v", wrrRPMs)

    // Calculate and compare load balance across all algorithms
    rrBalance := calculateLoadBalance(rrRPMs)
    llBalance := calculateLoadBalance(llRPMs)
    lcBalance := calculateLoadBalance(lcRPMs)
    wrrBalance := calculateLoadBalance(wrrRPMs)

    log.Printf("RoundRobin Load Balance: %f", rrBalance)
    log.Printf("LeastLoaded Load Balance: %f", llBalance)
    log.Printf("LeastConnections Load Balance: %f", lcBalance)
    log.Printf("WeightedRoundRobin Load Balance: %f", wrrBalance)

    // Determine the most effective algorithm
    bestAlgorithm := "RoundRobin"
    bestBalance := rrBalance

    if llBalance < bestBalance {
        bestAlgorithm = "LeastLoaded"
        bestBalance = llBalance
    }

    if lcBalance < bestBalance {
        bestAlgorithm = "LeastConnections"
        bestBalance = lcBalance
    }

    if wrrBalance < bestBalance {
        bestAlgorithm = "WeightedRoundRobin"
        bestBalance = wrrBalance
    }

// Check if all balances are the same
if rrBalance == llBalance && llBalance == lcBalance && lcBalance == wrrBalance {
    log.Println("No algorithm is more effective than the others; all have the same balance score.")
} else {
    log.Printf("Most effective algorithm: %s with balance score: %f", bestAlgorithm, bestBalance)
}}

func calculateLoadBalance(rpms []int64) float64 {
    // A simple metric to calculate load balance could be the variance or standard deviation of RPMs
    var sum, sumSquares float64
    n := float64(len(rpms))

    for _, rpm := range rpms {
        sum += float64(rpm)
        sumSquares += float64(rpm * rpm)
    }

    mean := sum / n
    variance := (sumSquares / n) - (mean * mean)
    return variance // Lower variance indicates more balanced load
}
