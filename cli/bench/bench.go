package bench

import (
	bctx "context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/FZambia/tarantool"
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

// verifyOperationsPercentage checks that the amount of operations percentage is 100.
func verifyOperationsPercentage(ctx *context.BenchCtx) error {
	entire_percentage := ctx.InsertCount + ctx.SelectCount + ctx.UpdateCount
	if entire_percentage != 100 {
		return fmt.Errorf(
			"The number of operations as a percentage should be equal to 100, " +
				"note that by default the percentage of inserts is 100")
	}
	return nil
}

// spacePreset prepares space for a benchmark.
func spacePreset(tarantoolConnection *tarantool.Connection) error {
	dropBenchmarkSpace(tarantoolConnection)
	return createBenchmarkSpace(tarantoolConnection)
}

// incrementRequest increases the counter of successful/failed requests depending on the presence of an error.
func (results *Results) incrementRequestsCounters(err error) {
	if err == nil {
		results.successResultCount++
	} else {
		results.failedResultCount++
	}
	results.handledRequestsCount++
}

// requestsLoop continuously executes the insert query until the benchmark time runs out.
func requestsLoop(requestsSequence *RequestsSequence, backgroundCtx bctx.Context) {
	for {
		select {
		case <-backgroundCtx.Done():
			return
		default:
			request := requestsSequence.getNext()
			request.operation(&request)
		}
	}
}

// connectionLoop runs "ctx.SimultaneousRequests" requests execution goroutines
// through the same connection.
func connectionLoop(
	ctx *context.BenchCtx,
	requestsSequence *RequestsSequence,
	backgroundCtx bctx.Context,
) {
	var connectionWait sync.WaitGroup
	for i := 0; i < ctx.SimultaneousRequests; i++ {
		connectionWait.Add(1)
		go func() {
			defer connectionWait.Done()
			requestsLoop(requestsSequence, backgroundCtx)
		}()
	}

	connectionWait.Wait()
}

// preFillBenchmarkSpaceIfRequired fills benchmark space
// if insert count = 0 or PreFillingCount flag is explicitly specified.
func preFillBenchmarkSpaceIfRequired(ctx context.BenchCtx, connectionPool []*tarantool.Connection) error {
	if ctx.InsertCount == 0 || ctx.PreFillingCount != PreFillingCount {
		fmt.Println("\nThe pre-filling of the space has started,\n" +
			"because the insert operation is not specified\n" +
			"or there was an explicit instruction for pre-filling.")
		fmt.Println("...")
		filledCount, err := fillBenchmarkSpace(ctx, connectionPool)
		if err != nil {
			return err
		}
		fmt.Printf("Pre-filling is finished. Number of records: %d\n\n", filledCount)
	}
	return nil
}

// Main benchmark function.
func Run(ctx context.BenchCtx) error {
	rand.Seed(time.Now().UnixNano())

	if err := verifyOperationsPercentage(&ctx); err != nil {
		return err
	}

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

	if err := spacePreset(tarantoolConnection); err != nil {
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

	if err := preFillBenchmarkSpaceIfRequired(ctx, connectionPool); err != nil {
		return err
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
			requestsSequence := RequestsSequence{
				[]RequestsGenerator{
					{
						Request{
							insertOperation,
							ctx,
							connection,
							&results,
						},
						ctx.InsertCount,
					},
					{
						Request{
							selectOperation,
							ctx,
							connection,
							&results,
						},
						ctx.SelectCount,
					},
					{
						Request{
							updateOperation,
							ctx,
							connection,
							&results,
						},
						ctx.UpdateCount,
					},
				},
				0,
				ctx.InsertCount,
				sync.Mutex{},
			}
			connectionLoop(&ctx, &requestsSequence, backgroundCtx)
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
