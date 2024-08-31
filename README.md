# load-balancer

go1.23.0

## Starting the load balancer
go build -o bin/app cmd/app/main.go

./bin/app

## Tests
Unit tests can be found under internal/loadbalancer/algorithms/{name_of_distribution_algorithm}/{name_of_distribution_algorithm.go}

Integrated stress test can be found under tests/stress.go

## To see Dynamic Scaling, start the load balancer like the above and
echo "GET http://localhost:8080" | vegeta attack -rate=10 -duration=10s | vegeta report

## Findings

### Performance Summary
1. Least Connections

Maximum Requests Before Error: 600
Test Duration: 0.22 seconds
Implication: Efficient and responsive for moderate loads. Limited in handling high concurrency due to its inability to manage large volumes of requests effectively.

2. Least Loaded

Maximum Requests Before Error: 600
Test Duration: 0.25 seconds
Implication: Similar to Least Connections in terms of performance. It performs well under moderate loads but may not scale as effectively under high concurrency.

3. Round Robin

Maximum Requests Before Error: 1200
Test Duration: 0.81 seconds
Implication: Provides a balance between efficiency and scalability. It handles a higher volume of requests compared to Least Connections and Least Loaded but may not be as optimal for very high concurrency.

4. Weighted Round Robin

Maximum Requests Before Error: 1700
Test Duration: 10 seconds
Implication: Exhibits the best scalability under high concurrency. It can handle the highest volume of concurrent requests efficiently. However, it has a longer test duration, which might indicate increased complexity in managing weighted distributions.

### Trade-Offs
- Efficiency vs. Scalability: Algorithms like Least Connections and Least Loaded are efficient and quick but may not handle very high concurrency well. On the other hand, Round Robin and Weighted Round Robin offer better scalability, with Weighted Round Robin handling the highest load, though with a longer processing time.

- Complexity vs. Performance: Simple algorithms like Round Robin are easy to implement and understand but may not be as performant under high load as more complex algorithms like Weighted Round Robin, which take server capacities into account.