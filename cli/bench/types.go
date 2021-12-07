package bench

// Results describes set of benchmark results.
type Results struct {
	handledRequestsCount int     // Count of all executed requests.
	successResultCount   int     // Count of successful request in all connections.
	failedResultCount    int     // Count of failed request in all connections.
	duration             float64 // Benchmark duration.
	requestsPerSecond    int     // Cumber of requests per second - the main measured value.
}
