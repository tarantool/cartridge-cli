package commands

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/replicasets"
)

func init() {
	var replicasetsCmd = &cobra.Command{
		Use:   "replicasets",
		Short: "Manage application topology",
	}

	rootCmd.AddCommand(replicasetsCmd)

	// replicasets sub-commands

	// setup topology from file
	var setupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Setup topology described in a file",

		Args: cobra.ExactValidArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runReplicasetsCommand(replicasets.Setup, args); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}
	setupCmd.Flags().StringVar(&ctx.Replicasets.File, "file", "", replicasetsFileUsage)
	setupCmd.Flags().BoolVar(
		&ctx.Replicasets.BootstrapVshard, "bootstrap-vshard", false, replicasetsBootstrapVshardUsage,
	)

	// join instances to replicaset
	var joinCmd = &cobra.Command{
		Use:   "join INSTANCE_NAME...",
		Short: "Join instance(s) to replicaset",

		Run: func(cmd *cobra.Command, args []string) {
			if err := runReplicasetsCommand(replicasets.Join, args); err != nil {
				log.Fatalf(err.Error())
			}
		},

		ValidArgsFunction: ShellCompRunningInstances,
	}

	addReplicasetFlag(joinCmd)

	// expel instance from cluster
	var expelCmd = &cobra.Command{
		Use:   "expel INSTANCE_NAME...",
		Short: "Expel instance(s) from topology",

		Run: func(cmd *cobra.Command, args []string) {
			if err := runReplicasetsCommand(replicasets.Expel, args); err != nil {
				log.Fatalf(err.Error())
			}
		},

		ValidArgsFunction: ShellCompRunningInstances,
	}

	addReplicasetFlag(expelCmd)

	// list available roles
	var listRolesCmd = &cobra.Command{
		Use:   "list-roles",
		Short: "List available roles",

		Args: cobra.ExactValidArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runReplicasetsCommand(replicasets.ListRoles, args); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	// add roles to replicaset
	var addRolesCmd = &cobra.Command{
		Use:   "add-roles ROLE_NAME...",
		Short: "Add role(s) to replicaset",

		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runReplicasetsCommand(replicasets.AddRoles, args); err != nil {
				log.Fatalf(err.Error())
			}
		},

		ValidArgsFunction: ShellCompRolesToAdd,
	}

	addReplicasetFlag(addRolesCmd)

	// remove roles from replicaset
	var removeRolesCmd = &cobra.Command{
		Use:   "remove-roles ROLE_NAME...",
		Short: "Remove role(s) from replicaset",

		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runReplicasetsCommand(replicasets.RemoveRoles, args); err != nil {
				log.Fatalf(err.Error())
			}
		},

		ValidArgsFunction: ShellCompReplicasetRoles,
	}

	addReplicasetFlag(removeRolesCmd)

	// set failover priority
	var setFailoverPriorityCmd = &cobra.Command{
		Use:   "set-failover-priority INSTANCE_NAME...",
		Short: "Set replicaset failover priority",

		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runReplicasetsCommand(replicasets.SetFailoverPriority, args); err != nil {
				log.Fatalf(err.Error())
			}
		},

		ValidArgsFunction: ShellCompRunningInstances,
	}

	addReplicasetFlag(setFailoverPriorityCmd)

	// bootstrap vshard
	var bootstrapVshardCmd = &cobra.Command{
		Use:   "bootstrap-vshard",
		Short: "Bootstrap vshard",

		Args: cobra.ExactValidArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runReplicasetsCommand(replicasets.BootstrapVshard, args); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	addReplicasetFlag(bootstrapVshardCmd)

	// set weight
	var setWeightCmd = &cobra.Command{
		Use:   "set-weight WEIGHT",
		Short: "Set replicaset weight",

		Args: cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runReplicasetsCommand(replicasets.SetWeight, args); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	addReplicasetFlag(setWeightCmd)

	// list vshard groups
	var listVshardGroupsCmd = &cobra.Command{
		Use:   "list-vshard-groups",
		Short: "List avaliable vshard groups",

		Args: cobra.ExactValidArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runReplicasetsCommand(replicasets.ListVshardGroups, args); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	addReplicasetFlag(listVshardGroupsCmd)

	// set vshard group
	var setVshardGroupCmd = &cobra.Command{
		Use:   "set-vshard-group GROUP_NAME",
		Short: "Set replicaset vshard group",

		Args: cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runReplicasetsCommand(replicasets.SetVshardGroup, args); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	addReplicasetFlag(setVshardGroupCmd)

	// add all sub-commands

	replicasetsSubCommands := []*cobra.Command{
		setupCmd,
		joinCmd,
		expelCmd,
		listRolesCmd,
		addRolesCmd,
		removeRolesCmd,
		setFailoverPriorityCmd,
		bootstrapVshardCmd,
		setWeightCmd,
		listVshardGroupsCmd,
		setVshardGroupCmd,
	}

	for _, cmd := range replicasetsSubCommands {
		replicasetsCmd.AddCommand(cmd)
		configureFlags(cmd)
		addCommonReplicasetsFlags(cmd)
	}
}

func runReplicasetsCommand(replicasetsFunc func(ctx *context.Ctx, args []string) error, args []string) error {
	if err := replicasets.FillCtx(&ctx); err != nil {
		return err
	}

	if err := replicasetsFunc(&ctx, args); err != nil {
		return err
	}

	return nil
}
