package types

type Node struct {
    Address      string
    BPM          int64 // Bytes Per Minute
    RPM          int   // Requests Per Minute
    CurrentBPM   int64
    CurrentRPM   int
	Weight       int    // Weight for weighted round-robin
}

type RateLimitResult struct {
    Allow bool
    Error string
}
