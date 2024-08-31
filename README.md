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

### Implications of the Results

#### **Round Robin Algorithm:**

- **Low Load, Even Distribution:** Round Robin handled 200 concurrent requests. This aligns with expectations, as Round Robin is known to perform well in scenarios with even distribution and similar node capacity. The performance under low load suggests that Round Robin operates effectively but may be influenced by factors specific to the test environment.
  
- **High Load, Even Distribution:** Round Robin performed significantly better under high load, handling 900 concurrent requests. This indicates that Round Robin scales well with increasing load when the distribution remains even.

#### **Least Loaded Algorithm:**

- **High Load, Unbalanced Nodes:** Least Loaded handled 900 concurrent requests, demonstrating its effectiveness in managing high loads across nodes with varying capacities. This is expected, as Least Loaded algorithms are designed to dynamically balance the load based on current node usage.

#### **Least Connections Algorithm:**

- **High Load, Varying Request Sizes:** Least Connections performed well under high load with varying request sizes, handling 900 concurrent requests. This result is in line with expectations, as the algorithm is tailored to distribute load based on the number of active connections, making it particularly effective when request sizes vary.

#### **Weighted Round Robin Algorithm:**

- **High Load, Weighted Distribution:** Weighted Round Robin handled 900 concurrent requests, confirming its capability to distribute load effectively based on node weights under high load conditions. This algorithm is well-suited for scenarios that require accounting for differences in node capacity.

#### **Failures in Low-Load Scenarios:**

- **Scenarios 3, 5, and 8:** No algorithm passed the test in these scenarios. The failures suggest that the tested algorithms may face challenges under low-load conditions, particularly when nodes are unbalanced or request sizes vary. These results highlight areas where the current algorithms may not be as effective.

### Does It Fit Known Behaviors?

- **Round Robin:** The performance in evenly distributed scenarios fits well with known behaviors. It is effective in environments with consistent, predictable traffic but may face challenges under more complex conditions.

- **Least Loaded and Least Connections:** These results align with expectations. These algorithms excel in scenarios where dynamic load balancing based on current conditions is critical.

- **Weighted Round Robin:** As anticipated, this algorithm performs well in scenarios involving weighted distribution under high load, confirming its intended use case.

- **Failures in Low-Load Scenarios:** The difficulties encountered in low-load, unbalanced, or varying request size scenarios suggest that these algorithms may be particularly optimized for high load or balanced conditions.

### Conclusion

The results show that the algorithms tested performed admirably in high-load scenarios but encountered challenges in certain low-load or unbalanced situations. This is consistent with the known characteristics of these algorithms, which are generally optimized for environments with significant traffic. The findings indicate that while these algorithms are reliable under heavy load, there may be opportunities to develop more adaptive algorithms that can handle a broader range of conditions, especially in environments with uneven node capacity or varying request sizes.
