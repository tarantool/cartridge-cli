package replicasets

import (
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/cluster"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func Expel(ctx *context.Ctx, args []string) error {
	if err := project.FillCtx(ctx); err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("Please, specify at least one instance to expel")
	}

	instancesToExpelNames := args

	instancesConf, err := cluster.GetInstancesConf(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	joinedInstances, err := cluster.GetMembershipInstances(instancesConf, ctx)
	if err != nil {
		return fmt.Errorf("Failed to get instances connected to membership: %s", err)
	}

	instancesToExpelMap := make(map[string]bool)
	for _, instanceName := range instancesToExpelNames {
		instancesToExpelMap[instanceName] = false
	}

	joinedInstanceName := ""

	// We need to find some instance that is joined and not specified to expel
	// (instance can't be expelled via it's own socket).
	// We also need to collect UUIDs of instances specified to expel.
	instancesToExpelUUIDs := make([]string, len(instancesToExpelNames))
	instancesToExpelUUIDsNum := 0
	for instanceURI, instance := range *joinedInstances {
		if instance.UUID != "" {
			if instance.Alias == "" {
				return fmt.Errorf("Failed to get alias for instance %s", instanceURI)
			}

			// check that instance name is specified to expel
			if _, found := instancesToExpelMap[instance.Alias]; !found {
				joinedInstanceName = instance.Alias
				continue
			}

			instancesToExpelMap[instance.Alias] = true

			instancesToExpelUUIDs[instancesToExpelUUIDsNum] = instance.UUID
			instancesToExpelUUIDsNum++
		}
	}

	for instanceName, instanceFound := range instancesToExpelMap {
		if !instanceFound {
			return common.ErrWrapCheckInstanceNameCommonMisprint([]string{instanceName}, ctx.Project.Name,
				fmt.Errorf("Instance %s isn't found in cluster", instanceName))
		}
	}

	if joinedInstanceName == "" {
		return fmt.Errorf("Not found any other non-expelled instance joined to cluster")
	}

	conn, err := cluster.ConnectToInstance(joinedInstanceName, ctx)
	if err != nil {
		return err
	}

	editInstancesOpts, err := getExpelInstancesEditInstancesOpts(instancesToExpelUUIDs)
	if err != nil {
		return fmt.Errorf("Failed to get edit_topology options for expelling instances: %s", err)
	}

	if _, err = editInstances(conn, editInstancesOpts); err != nil {
		return fmt.Errorf("Failed to expel instances: %s", err)
	}

	log.Infof(
		"Instance(s) %s have been successfully expelled",
		strings.Join(instancesToExpelNames, ", "),
	)

	return nil
}

func getExpelInstancesEditInstancesOpts(instancesToExpelUUIDs []string) (*EditInstancesListOpts, error) {
	editInstancesOpts := make(EditInstancesListOpts, len(instancesToExpelUUIDs))

	for i, instanceUUID := range instancesToExpelUUIDs {
		editInstancesOpts[i] = &EditInstanceOpts{
			InstanceUUID: instanceUUID,
			Expelled:     true,
		}
	}

	return &editInstancesOpts, nil
}
