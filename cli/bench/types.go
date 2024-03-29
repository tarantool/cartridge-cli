package bench

import (
	"sync"

	"github.com/FZambia/tarantool"
	"github.com/tarantool/cartridge-cli/cli/context"
)

// Results describes set of benchmark results.
type Results struct {
	handledRequestsCount int     // Count of all executed requests.
	successResultCount   int     // Count of successful request in all connections.
	failedResultCount    int     // Count of failed request in all connections.
	duration             float64 // Benchmark duration.
	requestsPerSecond    int     // Cumber of requests per second - the main measured value.
}

// RequestOperaion describes insert, select or update operation in request.
type RequestOperaion func(*Request)

// Request describes various types of requests.
type Request struct {
	operation           RequestOperaion // insertOperation, selectOperation or updateOperation.
	ctx                 context.BenchCtx
	tarantoolConnection *tarantool.Connection
	results             *Results
}

// RequestsGenerator data structure for abstraction of a renewable heap of identical requests.
type RequestsGenerator struct {
	request Request // Request with specified operation.
	count   int     // Count of requests.
}

// RequestsSequence data structure for abstraction for the constant issuance of new requests.
type RequestsSequence struct {
	requests []RequestsGenerator
	// currentRequestIndex describes what type of request will be issued by the sequence.
	currentRequestIndex int
	// currentCounter describes how many requests of the same type
	// are left to issue from RequestsPool.
	currentCounter int
	// findNewRequestsGeneratorMutex provides goroutine-safe search for new generator.
	findNewRequestsGeneratorMutex sync.Mutex
}
