package server

import (
    "fmt"
    "net/http"
    "github.com/devChorok/load-balancer/internal/loadbalancer"
    "github.com/devChorok/load-balancer/pkg/types"
)

func StartServer() {
    nodes := []*types.Node{
        {Address: "http://localhost:8081", BPM: 1000, RPM: 10},
        {Address: "http://localhost:8082", BPM: 1500, RPM: 15},
    }

    lb := loadbalancer.NewLoadBalancer(nodes)

    http.Handle("/", lb)

    fmt.Println("Load Balancer running on port 8080")
    http.ListenAndServe(":8080", nil)
}
