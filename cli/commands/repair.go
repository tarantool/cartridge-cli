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

	// list current cluster topology
	var repairListCmd = &cobra.Command{
		Use:   "list-topology",
		Short: "Get current cluster topology summary",
		Long:  `All configuration files across directories <data-dir>/<app-name>.* are read`,

		Args: cobra.ExactValidArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runRepairCommand(repair.List); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	// change advertise URI
	var repairURICmd = &cobra.Command{
		Use:   "set-advertise-uri INSTANCE-UUID NEW-URI",
		Short: "Change instance advertise URI",
		Long: `Rewrite advertise URI for specified instance in the instacnes config files.
All configuration files across directories <data-dir>/<app-name>.* are patched.`,

		Args: cobra.ExactValidArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ctx.Repair.SetURIInstanceUUID = args[0]
			ctx.Repair.NewURI = args[1]

			if err := runRepairCommand(repair.PatchURI); err != nil {
				log.Fatalf(err.Error())
			}
		},
		ValidArgsFunction: ShellCompRepairSetURI,
	}
	addCommonRepairPatchFlags(repairURICmd)

	// remove node from cluster
	var repairRemoveCmd = &cobra.Command{
		Use:   "remove-instance UUID",
		Short: "Remove instance from the cluster",
		Long: `Remove instance with specified UUID from all instances config files.
All configuration files across directories <data-dir>/<app-name>.* are patched.`,

		Args: cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx.Repair.RemoveInstanceUUID = args[0]

			if err := runRepairCommand(repair.RemoveInstance); err != nil {
				log.Fatalf(err.Error())
			}
		},
		ValidArgsFunction: ShellCompRepairRemove,
	}
	addCommonRepairPatchFlags(repairRemoveCmd)

	// set replicaset leader
	var repairSetLeaderCmd = &cobra.Command{
		Use:   "set-leader REPLICASET-UUID INSTANCE-UUID",
		Short: "Change replicaset leader",
		Long: `Set specified replicaset leader to specified instance in all instances config files.
All configuration files across directories <data-dir>/<app-name>.* are patched.`,

		Args: cobra.ExactValidArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ctx.Repair.SetLeaderReplicasetUUID = args[0]
			ctx.Repair.SetLeaderInstanceUUID = args[1]

			if err := runRepairCommand(repair.SetLeader); err != nil {
				log.Fatalf(err.Error())
			}
		},
		ValidArgsFunction: ShellCompRepairSetLeader,
	}
	addCommonRepairPatchFlags(repairSetLeaderCmd)

	repairSubCommands := []*cobra.Command{
		repairListCmd,
		repairURICmd,
		repairRemoveCmd,
		repairSetLeaderCmd,
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
