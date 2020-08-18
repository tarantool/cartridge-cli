package commands

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/repair"
)

func init() {
	var repairCmd = &cobra.Command{
		Use:   "repair",
		Short: "Patch cluster configuration files",
	}

	rootCmd.AddCommand(repairCmd)

	// repair sub-commands

	// change advertise URI
	var repairURICmd = &cobra.Command{
		Use:   "set-uri URI-FROM URI-TO",
		Short: "Change instance advertise URI",
		Long: `Rewrite specified advertise URI in the instacnes config files.
All configuration files across directories <data-dir>/<app-name>.* are patched.`,

		Args: cobra.ExactValidArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ctx.Repair.OldURI = args[0]
			ctx.Repair.NewURI = args[1]

			if err := runRepairCommand(repair.PatchURI); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	repairSubCommands := []*cobra.Command{
		repairURICmd,
	}

	for _, cmd := range repairSubCommands {
		repairCmd.AddCommand(cmd)
		configureFlags(cmd)
		addCommonRepairFlags(cmd)
	}
}

func runRepairCommand(repairFunc func(ctx *context.Ctx) error) error {
	if ctx.Project.Name == "" {
		return fmt.Errorf("Please, specify application name using --name")
	}

	if err := repairFunc(&ctx); err != nil {
		return err
	}

	return nil
}
