package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tarantool/cartridge-cli/cli/version"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Cartridge CLI",
	Long:  `All software has versions. This is Cartridge CLI's`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.BuildVersionString())
	},
}
