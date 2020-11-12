package replicasets

import (
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func Expel(ctx *context.Ctx, args []string) error {
	if err := FillCtx(ctx); err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("Please, specify at least one instance to expel")
	}

	instancesToExpelNames := args

	instancesConf, err := getInstancesConf(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	joinedInstances, err := getJoinedInstances(instancesConf, ctx)
	if err != nil {
		return fmt.Errorf("Failed to get instances connected to membership: %s", err)
	}

	instancesToExpelMap := make(map[string]bool)
	for _, instanceName := range instancesToExpelNames {
		instancesToExpelMap[instanceName] = false
	}

	joinedInstanceName := ""

	instancesToExpelUUIDs := make([]string, len(instancesToExpelNames))
	instancesToExpelURIsNum := 0
	for instanceURI, instance := range *joinedInstances {
		if instance.UUID != "" {
			if instance.Alias == "" {
				return fmt.Errorf("Failed to get alias for instance %s", instanceURI)
			}

			joinedInstanceName = instance.Alias

			// check that instance name is specified to expel
			if _, found := instancesToExpelMap[instance.Alias]; !found {
				continue
			}

			instancesToExpelUUIDs[instancesToExpelURIsNum] = instance.UUID
			instancesToExpelURIsNum++
		}
	}

	if joinedInstanceName == "" {
		return fmt.Errorf("Not found instances joined to cluster")
	}

	consoleSockPath := project.GetInstanceConsoleSock(ctx, joinedInstanceName)
	conn, err := common.ConnectToTarantoolSocket(consoleSockPath)
	if err != nil {
		return fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	log.Debugf("Connected to %s", consoleSockPath)

	editInstancesOpts, err := getExpelInstancesEditReplicasetsOpts(instancesToExpelUUIDs)
	if err != nil {
		return fmt.Errorf("Failed to get edit_topology options for expelling instances: %s", err)
	}

	if _, err = editInstances(conn, editInstancesOpts); err != nil {
		return fmt.Errorf("Failed to expel instances: %s", err)
	}

	log.Infof(
		"Instance(s) %s was successfully expelled",
		strings.Join(instancesToExpelNames, ", "),
	)

	return nil
}

func getExpelInstancesEditReplicasetsOpts(instancesToExpelUUIDs []string) (*EditInstancesOpts, error) {
	editInstancesOpts := make(EditInstancesOpts, len(instancesToExpelUUIDs))

	for i, instanceUUID := range instancesToExpelUUIDs {
		editInstancesOpts[i] = &EditInstanceOpts{
			InstanceUUID: instanceUUID,
			Expelled:     true,
		}
	}

	return &editInstancesOpts, nil
}
