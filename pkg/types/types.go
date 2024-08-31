package types

import (
	"time"
)

type Node struct {
	Address          string    // The address of the node (e.g., "http://localhost:8081")
	BPM              int64     // Bytes Per Minute (Capacity metric)
	RPM              int64     // Requests Per Minute (Capacity metric)
	CurrentBPM       int64     // Current Bytes Per Minute (Usage tracking)
	CurrentRPM       int64     // Current Requests Per Minute (Usage tracking)
	Weight           int       // Weight for weighted round-robin
	LastRequest      time.Time // Timestamp of the last request handled by this node
	ActiveConnections int      // Number of active connections currently handled by this node
	Healthy          bool      // Indicates if the node is healthy and can handle requests
}

type RateLimitResult struct {
	Allow bool
	Error string
}
