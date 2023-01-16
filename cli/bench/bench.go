package bench

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/tarantool/cartridge-cli/cli/context"
)

// Main benchmark function.
func Run(ctx context.BenchCtx) error {
	rand.Seed(time.Now().UnixNano())

	if err := verifyOperationsPercentage(&ctx); err != nil {
		return err
	}

	// Check cluster topology for further actions.
	cluster, err := isCluster(ctx)
	if err != nil {
		return err
	}

	if cluster {
		// Check cluster for wrong topology.
		if err := verifyClusterTopology(ctx); err != nil {
			return err
		}
		// Get url of one of instances in cluster for space preset and prefill.
		ctx.URL = (*ctx.Leaders)[0]
	}

	// Connect to tarantool and preset space for benchmark.
	tarantoolConnection, err := createConnection(ctx)
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

	if err := preFillBenchmarkSpaceIfRequired(ctx); err != nil {
		return err
	}

	fmt.Println("Benchmark start")
	fmt.Println("...")

	// Bench one instance by default.
	benchStart := benchOneInstance
	if cluster {
		benchStart = benchCluster
	}

	// Prepare data for bench.
	benchData := getBenchData(ctx)

	// Start benching.
	if err := benchStart(ctx, &benchData); err != nil {
		return err
	}

	// Calculate results.
	benchData.results.duration = time.Since(benchData.startTime).Seconds()
	benchData.results.requestsPerSecond = int(float64(benchData.results.handledRequestsCount) / benchData.results.duration)

	// Benchmark space must exist after bench.
	if err := dropBenchmarkSpace(tarantoolConnection); err != nil {
		return err
	}
	fmt.Println("Benchmark stop.")

	printResults(benchData.results)
	return nil
}
