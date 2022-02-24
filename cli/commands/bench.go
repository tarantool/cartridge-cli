package commands

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/tarantool/cartridge-cli/cli/bench"
)

func init() {
	var benchCmd = &cobra.Command{
		Use:   "bench",
		Short: "Util for running benchmarks for Tarantool",
		Long:  "Benchmark utility that simulates running commands done by N clients at the same time sending M simultaneous queries",
		Run: func(cmd *cobra.Command, args []string) {
			if err := bench.Run(ctx.Bench); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}
	rootCmd.AddCommand(benchCmd)

	configureFlags(benchCmd)

	benchCmd.Flags().StringVar(&ctx.Bench.URL, "url", "127.0.0.1:3301", "Tarantool address")
	benchCmd.Flags().StringVar(&ctx.Bench.User, "user", "guest", "Tarantool user for connection")
	benchCmd.Flags().StringVar(&ctx.Bench.Password, "password", "", "Tarantool password for connection")

	benchCmd.Flags().IntVar(&ctx.Bench.Connections, "connections", 10, "Number of concurrent connections")
	benchCmd.Flags().IntVar(&ctx.Bench.SimultaneousRequests, "requests", 10, "Number of simultaneous requests per connection")
	benchCmd.Flags().IntVar(&ctx.Bench.Duration, "duration", 10, "Duration of benchmark test (seconds)")
	benchCmd.Flags().IntVar(&ctx.Bench.KeySize, "keysize", 10, "Size of key part of benchmark data (bytes)")
	benchCmd.Flags().IntVar(&ctx.Bench.DataSize, "datasize", 20, "Size of value part of benchmark data (bytes)")

	benchCmd.Flags().IntVar(&ctx.Bench.InsertCount, "insert", 100, "percentage of inserts")
	benchCmd.Flags().IntVar(&ctx.Bench.SelectCount, "select", 0, "percentage of selects")
	benchCmd.Flags().IntVar(&ctx.Bench.UpdateCount, "update", 0, "percentage of updates")
	benchCmd.Flags().IntVar(&ctx.Bench.PreFillingCount, "fill", bench.PreFillingCount, "number of records to pre-fill the space")

}
