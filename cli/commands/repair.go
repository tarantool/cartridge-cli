package commands

import (
	"fmt"

	"github.com/apex/log"
	"github.com/spf13/cobra"
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
			err := runRepairListCommand(cmd, args)
			if err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	// change advertise URI
	var repairURICmd = &cobra.Command{
		Use:   "set-uri URI-FROM URI-TO",
		Short: "Change instance advertise URI",
		Long: `Rewrite specified advertise URI in the instacnes config files.
All configuration files across directories <data-dir>/<app-name>.* are patched.`,

		Args: cobra.ExactValidArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := runRepairURICommand(cmd, args)
			if err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	// remove node from cluster
	var repairRemoveCmd = &cobra.Command{
		Use:   "remove-instance UUID",
		Short: "Remove instance from the cluster",
		Long: `Remove instance with specified UUID from all instances config files.
All configuration files across directories <data-dir>/<app-name>.* are patched.`,

		Args: cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := runRepairRemoveCommand(cmd, args)
			if err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	// set replicaset master
	var repairSetMasterCmd = &cobra.Command{
		Use:   "set-leader REPLICASET-UUID INSTANCE-UUID",
		Short: "Change replicaset leader",
		Long: `Set specified replicaset leader to specified instance in all instances config files.
All configuration files across directories <data-dir>/<app-name>.* are patched.`,

		Args: cobra.ExactValidArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := runRepairSetMasterCommand(cmd, args)
			if err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	repairSubCommands := []*cobra.Command{
		repairListCmd,
		repairURICmd,
		repairRemoveCmd,
		repairSetMasterCmd,
	}

	for _, cmd := range repairSubCommands {
		repairCmd.AddCommand(cmd)
		configureFlags(cmd)
		addCommonRepairFlags(cmd)
	}
}

func runRepairListCommand(cmd *cobra.Command, args []string) error {
	if ctx.Project.Name == "" {
		return fmt.Errorf("Please, specify application name using --name")
	}

	return repair.List(&ctx)
}

func runRepairURICommand(cmd *cobra.Command, args []string) error {
	if ctx.Project.Name == "" {
		return fmt.Errorf("Please, specify application name using --name")
	}

	ctx.Repair.OldURI = args[0]
	ctx.Repair.NewURI = args[1]

	return repair.PatchURI(&ctx)
}

func runRepairRemoveCommand(cmd *cobra.Command, args []string) error {
	if ctx.Project.Name == "" {
		return fmt.Errorf("Please, specify application name using --name")
	}

	ctx.Repair.RemoveInstanceUUID = args[0]

	return repair.RemoveInstance(&ctx)
}

func runRepairSetMasterCommand(cmd *cobra.Command, args []string) error {
	if ctx.Project.Name == "" {
		return fmt.Errorf("Please, specify application name using --name")
	}

	ctx.Repair.SetMasterReplicasetUUID = args[0]
	ctx.Repair.SetMasterInstanceUUID = args[1]

	return repair.SetMaster(&ctx)
}
