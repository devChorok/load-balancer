# load-balancer

go1.23.0

## Starting the load balancer
go build -o bin/app cmd/app/main.go

./bin/app

## Running unit tests
### all tests
 go test -v ./... 

### specific tests
go test -v -run TestRoundRobin ./internal/loadbalancer/algorithms/${name_of_algorithm}
