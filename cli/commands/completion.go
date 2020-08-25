package commands

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/repair"
	"github.com/tarantool/cartridge-cli/cli/running"
)

// RUNNING

func ShellCompRunningInstances(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var err error
	specifiedInstances := make(map[string]bool)

	for _, arg := range args {
		specifiedInstances[arg] = true
	}

	ctx.Running.AppDir, err = os.Getwd()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if ctx.Project.Name == "" {
		if ctx.Project.Name, err = project.DetectName(ctx.Running.AppDir); err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}

	if err := project.SetLocalRunningPaths(&ctx); err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	instances, err := running.CollectInstancesFromConf(&ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var notSpecifiedInstances []string
	for _, instance := range instances {
		if _, found := specifiedInstances[instance]; !found {
			notSpecifiedInstances = append(notSpecifiedInstances, instance)
		}
	}

	return notSpecifiedInstances, cobra.ShellCompDirectiveNoFileComp
}

// REPAIR

func ShellCompRepairSetURI(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	if len(args) == 0 {
		// first argument - instance UUID
		instanceUUIDs, err := repair.GetAllInstanceUUIDsComp(&ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return instanceUUIDs, cobra.ShellCompDirectiveNoFileComp
	}

	// second argument - instance URI (complete host)
	instanceUUID := args[0]
	instanceHosts, err := repair.GetInstanceHostsComp(instanceUUID, &ctx)

	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return instanceHosts, cobra.ShellCompDirectiveNoSpace
}

func ShellCompRepairRemove(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// one argument - instance UUID
	instanceUUIDs, err := repair.GetAllInstanceUUIDsComp(&ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return instanceUUIDs, cobra.ShellCompDirectiveNoFileComp
}

func ShellCompRepairSetLeader(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if len(args) == 0 {
		// first argument - replicaset UUID
		replicasetUUIDs, err := repair.GetAllReplicasetUUIDsComp(&ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return replicasetUUIDs, cobra.ShellCompDirectiveNoFileComp
	}

	// second argument - replicaset instance
	replicasetUUID := args[0]
	instanceUUIDs, err := repair.GetReplicasetInstancesComp(replicasetUUID, &ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return instanceUUIDs, cobra.ShellCompDirectiveNoFileComp
}
