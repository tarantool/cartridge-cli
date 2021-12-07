package bench

import (
	bctx "context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/FZambia/tarantool"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

// printResults outputs benchmark foramatted results.
func printResults(results Results) {
	fmt.Printf("\nResults:\n")
	fmt.Printf("\tSuccess operations: %d\n", results.successResultCount)
	fmt.Printf("\tFailed  operations: %d\n", results.failedResultCount)
	fmt.Printf("\tRequest count: %d\n", results.handledRequestsCount)
	fmt.Printf("\tTime (seconds): %f\n", results.duration)
	fmt.Printf("\tRequests per second: %d\n\n", results.requestsPerSecond)
}

// spacePreset prepares space for a benchmark.
func spacePreset(ctx context.BenchCtx, tarantoolConnection *tarantool.Connection) error {
	dropBenchmarkSpace(tarantoolConnection)
	return createBenchmarkSpace(tarantoolConnection)
}

// incrementRequest increases the counter of successful/failed requests depending on the presence of an error.
func incrementRequest(err error, results *Results) {
	if err == nil {
		results.successResultCount++
	} else {
		results.failedResultCount++
	}
	results.handledRequestsCount++
}

// requestsLoop continuously executes the insert query until the benchmark time runs out.
func requestsLoop(ctx context.BenchCtx, tarantoolConnection *tarantool.Connection, results *Results, backgroundCtx bctx.Context) {
	for {
		select {
		case <-backgroundCtx.Done():
			return
		default:
			_, err := tarantoolConnection.Exec(
				tarantool.Insert(
					benchSpaceName,
					[]interface{}{common.RandomString(ctx.KeySize), common.RandomString(ctx.DataSize)}))
			incrementRequest(err, results)
		}
	}
}

// connectionLoop runs "ctx.SimultaneousRequests" requests execution threads through the same connection.
func connectionLoop(ctx context.BenchCtx, tarantoolConnection *tarantool.Connection, results *Results, backgroundCtx bctx.Context) {
	var connectionWait sync.WaitGroup
	for i := 0; i < ctx.SimultaneousRequests; i++ {
		connectionWait.Add(1)
		go func() {
			defer connectionWait.Done()
			requestsLoop(ctx, tarantoolConnection, results, backgroundCtx)
		}()
	}

	connectionWait.Wait()
}

// Main benchmark function.
func Run(ctx context.BenchCtx) error {
	rand.Seed(time.Now().UnixNano())

	// Connect to tarantool and preset space for benchmark.
	tarantoolConnection, err := tarantool.Connect(ctx.URL, tarantool.Opts{
		User:     ctx.User,
		Password: ctx.Password,
	})
	if err != nil {
		return fmt.Errorf(
			"Couldn't connect to Tarantool %s.",
			ctx.URL)
	}
	defer tarantoolConnection.Close()

	printConfig(ctx, tarantoolConnection)

	if err := spacePreset(ctx, tarantoolConnection); err != nil {
		return err
	}

	/// Сreate a "connectionPool" before starting the benchmark to exclude the connection establishment time from measurements.
	connectionPool := make([]*tarantool.Connection, ctx.Connections)
	for i := 0; i < ctx.Connections; i++ {
		connectionPool[i], err = tarantool.Connect(ctx.URL, tarantool.Opts{
			User:     ctx.User,
			Password: ctx.Password,
		})
		if err != nil {
			return err
		}
		defer connectionPool[i].Close()
	}

	fmt.Println("Benchmark start")
	fmt.Println("...")

	// The "context" will be used to stop all "connectionLoop" when the time is out.
	backgroundCtx, cancel := bctx.WithCancel(bctx.Background())
	var waitGroup sync.WaitGroup
	results := Results{}

	startTime := time.Now()
	timer := time.NewTimer(time.Duration(ctx.Duration * int(time.Second)))

	// Start detached connections.
	for i := 0; i < ctx.Connections; i++ {
		waitGroup.Add(1)
		go func(connection *tarantool.Connection) {
			defer waitGroup.Done()
			connectionLoop(ctx, connection, &results, backgroundCtx)
		}(connectionPool[i])
	}
	// Sends "signal" to all "connectionLoop" and waits for them to complete.
	<-timer.C
	cancel()
	waitGroup.Wait()

	results.duration = time.Since(startTime).Seconds()
	results.requestsPerSecond = int(float64(results.handledRequestsCount) / results.duration)

	dropBenchmarkSpace(tarantoolConnection)
	fmt.Println("Benchmark stop")

	printResults(results)
	return nil
}
