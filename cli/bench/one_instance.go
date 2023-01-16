package bench

import (
	"sync"
	"time"

	"github.com/FZambia/tarantool"
	"github.com/tarantool/cartridge-cli/cli/context"
)

// benchOneInstance execute bench algorithm for only one tarantool instance.
func benchOneInstance(ctx context.BenchCtx, benchData *BenchmarkData) error {
	// Сreate a "connectionsPool" before starting the benchmark
	// to exclude the connection establishment time from measurements.
	connectionsPool, err := createConnectionsPool(ctx)
	if err != nil {
		return err
	}
	defer deleteConnectionsPool(connectionsPool)

	benchData.startTime = time.Now()

	// Start detached connections.
	for i := 0; i < ctx.Connections; i++ {
		benchData.waitGroup.Add(1)
		go func(connection *tarantool.Connection) {
			defer benchData.waitGroup.Done()
			requestsSequence := RequestsSequence{
				requests: []RequestsGenerator{
					{
						request: Request{
							operation:           insertOperation,
							ctx:                 ctx,
							tarantoolConnection: connection,
							results:             &benchData.results,
						},
						count: ctx.InsertCount,
					},
					{
						request: Request{
							operation:           selectOperation,
							ctx:                 ctx,
							tarantoolConnection: connection,
							results:             &benchData.results,
						},
						count: ctx.SelectCount,
					},
					{
						request: Request{
							operation:           updateOperation,
							ctx:                 ctx,
							tarantoolConnection: connection,
							results:             &benchData.results,
						},
						count: ctx.UpdateCount,
					},
				},
				currentRequestIndex:           0,
				currentCounter:                ctx.InsertCount,
				findNewRequestsGeneratorMutex: sync.Mutex{},
			}

			// Start looped requests in connection.
			var connectionWait sync.WaitGroup
			for i := 0; i < ctx.SimultaneousRequests; i++ {
				connectionWait.Add(1)
				go func() {
					defer connectionWait.Done()
					requestsLoop(&requestsSequence, benchData.backgroundCtx)
				}()
			}
			connectionWait.Wait()
		}(connectionsPool[i])
	}

	waitBenchEnd(benchData)
	return nil
}
