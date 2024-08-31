package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
    // Define a flag to accept the port number
    port := flag.String("port", "8080", "port to listen on")
    flag.Parse()

    // Correctly format the address with a single colon
	address := "localhost:" + *port

    // Simple handler to respond with a message indicating the port
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, you've hit the server running on port %s\n", *port)
    })


    // Start the server on the specified port
    log.Printf("Starting server on port %s...", address)
    if err := http.ListenAndServe(address, nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
