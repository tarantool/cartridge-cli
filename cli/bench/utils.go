package bench

import (
	bctx "context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
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

// printConfig output formatted config parameters.
func printConfig(ctx context.BenchCtx, tarantoolConnection *tarantool.Connection) {
	fmt.Printf("%s\n", tarantoolConnection.Greeting().Version)
	fmt.Printf("Parameters:\n")
	if cluster, _ := isCluster(ctx); cluster {
		fmt.Printf("\tLeaders:\n")
		fmt.Printf("\t\t%s\n", strings.Join(*ctx.Leaders, "\n"))
		fmt.Printf("\tReplicas:\n")
		fmt.Printf("\t\t%s\n", strings.Join(*ctx.Replicas, "\n"))
	} else {
		fmt.Printf("\tURL: %s\n", ctx.URL)
	}
	fmt.Printf("\tuser: %s\n", ctx.User)
	fmt.Printf("\tconnections: %d\n", ctx.Connections)
	fmt.Printf("\tsimultaneous requests: %d\n", ctx.SimultaneousRequests)
	fmt.Printf("\tduration: %d seconds\n", ctx.Duration)
	fmt.Printf("\tkey size: %d bytes\n", ctx.KeySize)
	fmt.Printf("\tdata size: %d bytes\n", ctx.DataSize)
	fmt.Printf("\tinsert: %d percentages\n", ctx.InsertCount)
	fmt.Printf("\tselect: %d percentages\n", ctx.SelectCount)
	fmt.Printf("\tupdate: %d percentages\n\n", ctx.UpdateCount)

	fmt.Printf("Data schema\n")
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintf(w, "|\tkey\t|\tvalue\n")
	fmt.Fprintf(w, "|\t------------------------------\t|\t------------------------------\n")
	fmt.Fprintf(w, "|\trandom(%d)\t|\trandom(%d)\n", ctx.KeySize, ctx.DataSize)
	w.Flush()
}

func isCluster(ctx context.BenchCtx) (bool, error) {
	result := (len(*ctx.Leaders) > 0 && len(*ctx.Replicas) > 0) || (len(*ctx.Leaders) > 1)
	if result == false && (len(*ctx.Leaders) > 0 || len(*ctx.Replicas) > 0) {
		return result, fmt.Errorf("Cluster: at least one leader and replica, or two leaders must be specified")
	}
	return result, nil
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
			request.operation(request)
		}
	}
}

// spacePreset prepares space for a benchmark.
func spacePreset(tarantoolConnection *tarantool.Connection) error {
	dropBenchmarkSpace(tarantoolConnection)
	return createBenchmarkSpace(tarantoolConnection)
}

// preFillBenchmarkSpaceIfRequired fills benchmark space
// if insert count = 0 or PreFillingCount flag is explicitly specified.
func preFillBenchmarkSpaceIfRequired(ctx context.BenchCtx) error {
	if ctx.InsertCount == 0 || ctx.PreFillingCount != PreFillingCount {
		connectionsPool, err := createConnectionsPool(ctx)
		if err != nil {
			return err
		}
		defer deleteConnectionsPool(connectionsPool)

		fmt.Println("\nThe pre-filling of the space has started,\n" +
			"because the insert operation is not specified\n" +
			"or there was an explicit instruction for pre-filling.")
		fmt.Println("...")
		filledCount, err := fillBenchmarkSpace(ctx, connectionsPool)
		if err != nil {
			return err
		}
		fmt.Printf("Pre-filling is finished. Number of records: %d\n\n", filledCount)
	}
	return nil
}

func getBenchData(ctx context.BenchCtx) BenchmarkData {
	backgroundCtx, cancel := bctx.WithCancel(bctx.Background())
	timer := time.NewTimer(time.Duration(ctx.Duration * int(time.Second)))
	return BenchmarkData{
		backgroundCtx: backgroundCtx,
		cancel:        cancel,
		timer:         timer,
	}
}

func waitBenchEnd(benchData *BenchmarkData) {
	// Sends "signal" to all "requestsLoop" and waits for them to complete.
	<-benchData.timer.C
	benchData.cancel()
	benchData.waitGroup.Wait()
}
