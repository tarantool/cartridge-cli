package bench

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/FZambia/tarantool"
	"github.com/tarantool/cartridge-cli/cli/context"
)

var (
	benchSpaceName             = "__benchmark_space__"
	benchSpacePrimaryIndexName = "__bench_primary_key__"
	PreFillingCount            = 1000000
	getRandomTupleCommand      = fmt.Sprintf(
		"box.space.%s.index.%s:random",
		benchSpaceName,
		benchSpacePrimaryIndexName,
	)
)

// printConfig output formatted config parameters.
func printConfig(ctx context.BenchCtx, tarantoolConnection *tarantool.Connection) {
	fmt.Printf("%s\n", tarantoolConnection.Greeting().Version)
	fmt.Printf("Parameters:\n")
	fmt.Printf("\tURL: %s\n", ctx.URL)
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
